#!/bin/bash

NAMESPACE="pg"

# Obtener la password del superuser
PGPASSWORD=$(kubectl get secret pg-cluster-superuser -n $NAMESPACE -o jsonpath='{.data.password}' 2>/dev/null | base64 -d)

if [ -z "$PGPASSWORD" ]; then
  echo "❌ No se encontró el secret pg-cluster-superuser"
  kubectl get secrets -n $NAMESPACE | grep pg-cluster
  exit 1
fi

echo "✅ Password obtenida"

# Crear secret
kubectl create secret generic postgres-exporter-secret \
  --from-literal=DATA_SOURCE_NAME="postgresql://postgres:${PGPASSWORD}@pg-cluster-rw:5432/postgres?sslmode=disable" \
  --namespace=$NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

# Aplicar ConfigMap y Deployment con v0.18.1
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-exporter-queries
  namespace: $NAMESPACE
data:
  queries.yaml: |
    custom_pg_stat_statements:
      query: |
        SELECT queryid::text,
               calls,
               total_exec_time,
               mean_exec_time,
               rows
        FROM pg_stat_statements
        WHERE query NOT LIKE '%pg_stat_statements%'
          AND queryid IS NOT NULL
        ORDER BY total_exec_time DESC
        LIMIT 20
      cache_seconds: 120
      metrics:
        - queryid:
            usage: "LABEL"
            description: "Query ID"
        - calls:
            usage: "COUNTER"
            description: "Times executed"
        - total_exec_time:
            usage: "COUNTER"
            description: "Total exec time (ms)"
        - mean_exec_time:
            usage: "GAUGE"
            description: "Mean exec time (ms)"
        - rows:
            usage: "COUNTER"
            description: "Rows affected"

    custom_pg_database_size:
      query: |
        SELECT datname,
               pg_database_size(datname) AS size_bytes
        FROM pg_database
        WHERE datname NOT IN ('template0','template1')
      cache_seconds: 60
      metrics:
        - datname:
            usage: "LABEL"
        - size_bytes:
            usage: "GAUGE"

    custom_pg_table_size:
      query: |
        SELECT schemaname || '.' || tablename AS table_name,
               pg_total_relation_size(quote_ident(schemaname)||'.'||quote_ident(tablename))::bigint AS total_bytes,
               pg_relation_size(quote_ident(schemaname)||'.'||quote_ident(tablename))::bigint AS table_bytes
        FROM pg_tables
        WHERE schemaname NOT IN ('pg_catalog','information_schema')
        ORDER BY total_bytes DESC
        LIMIT 20
      cache_seconds: 300
      metrics:
        - table_name:
            usage: "LABEL"
        - total_bytes:
            usage: "GAUGE"
        - table_bytes:
            usage: "GAUGE"

    custom_pg_cache_hit_ratio:
      query: |
        SELECT 'buffer_cache' AS cache_type,
               COALESCE(
                 sum(heap_blks_hit)::float /
                 NULLIF(sum(heap_blks_hit)+sum(heap_blks_read),0),
                 0
               ) AS ratio
        FROM pg_statio_user_tables
      cache_seconds: 30
      metrics:
        - cache_type:
            usage: "LABEL"
        - ratio:
            usage: "GAUGE"

    custom_pg_connections_state:
      query: |
        SELECT state,
               count(*) AS connections
        FROM pg_stat_activity
        WHERE backend_type = 'client backend'
        GROUP BY state
      cache_seconds: 10
      metrics:
        - state:
            usage: "LABEL"
        - connections:
            usage: "GAUGE"

    custom_pg_slow_queries:
      query: |
        SELECT pid::text,
               usename,
               datname,
               left(query, 80) AS query_text,
               EXTRACT(EPOCH FROM (now() - query_start))::float AS duration_seconds
        FROM pg_stat_activity
        WHERE state = 'active'
          AND backend_type = 'client backend'
          AND usename NOT IN ('streaming_replica')
          AND query NOT ILIKE 'START_REPLICATION%'
          AND query NOT LIKE '%pg_stat_activity%'
          AND (now() - query_start) > interval '5 seconds'
        ORDER BY duration_seconds DESC
        LIMIT 10
      cache_seconds: 10
      metrics:
        - pid:
            usage: "LABEL"
        - usename:
            usage: "LABEL"
        - datname:
            usage: "LABEL"
        - query_text:
            usage: "LABEL"
        - duration_seconds:
            usage: "GAUGE"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres-exporter
  namespace: $NAMESPACE
  labels:
    app: postgres-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres-exporter
  template:
    metadata:
      labels:
        app: postgres-exporter
    spec:
      containers:
      - name: exporter
        image: quay.io/prometheuscommunity/postgres-exporter:v0.18.1
        args:
          - --log.level=debug
          - --web.listen-address=:9187
          - --web.telemetry-path=/metrics
          # flag (deprecated pero OK) como backup
          - --extend.query-path=/etc/postgres_exporter/queries.yaml
        env:
        - name: DATA_SOURCE_NAME
          valueFrom:
            secretKeyRef:
              name: postgres-exporter-secret
              key: DATA_SOURCE_NAME
        # forma recomendada hoy (source of truth)
        - name: PG_EXPORTER_EXTEND_QUERY_PATH
          value: /etc/postgres_exporter/queries.yaml
        ports:
        - name: metrics
          containerPort: 9187
        volumeMounts:
        - name: queries
          mountPath: /etc/postgres_exporter/queries.yaml
          subPath: queries.yaml
          readOnly: true
      volumes:
      - name: queries
        configMap:
          name: postgres-exporter-queries
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-exporter
  namespace: $NAMESPACE
  labels:
    app: postgres-exporter
spec:
  type: ClusterIP
  ports:
  - name: metrics
    port: 9187
    targetPort: 9187
    protocol: TCP
  selector:
    app: postgres-exporter
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: postgres-exporter
  namespace: $NAMESPACE
  labels:
    release: prometheus
    app: postgres-exporter
spec:
  selector:
    matchLabels:
      app: postgres-exporter
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    scheme: http
EOF

echo ""
echo "✅ Postgres Exporter v0.18.1 desplegado"
echo ""
echo "Nuevas features en v0.18.1:"
echo "  - Soporte mejorado para PostgreSQL 17/18"
echo "  - stat_progress_vacuum collector"
echo "  - buffercache_summary collector"
echo "  - Query text exportado junto con queryId en pg_stat_statements"
echo "  - Mejor manejo de conexiones"
echo ""
echo "Verificar:"
echo "  kubectl get pods -n $NAMESPACE -l app=postgres-exporter"
echo "  kubectl logs -n $NAMESPACE -l app=postgres-exporter"
