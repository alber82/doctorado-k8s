#!/bin/bash

NAMESPACE="pg"

echo "🔍 Diagnóstico del Postgres Exporter"
echo "===================================="
echo ""

echo "1. Estado del deployment:"
kubectl get deployment postgres-exporter -n $NAMESPACE

echo ""
echo "2. Estado de los pods:"
kubectl get pods -n $NAMESPACE -l app=postgres-exporter

echo ""
echo "3. Logs recientes (últimas 50 líneas):"
kubectl logs -n $NAMESPACE -l app=postgres-exporter --tail=50

echo ""
echo "4. Variables de entorno:"
POD=$(kubectl get pod -n $NAMESPACE -l app=postgres-exporter -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n $NAMESPACE $POD -- env | grep -E "DATA_SOURCE|PG_"

echo ""
echo "5. Connection string (censurada):"
kubectl get secret postgres-exporter-secret -n $NAMESPACE -o jsonpath='{.data.DATA_SOURCE_NAME}' | base64 -d | sed 's/:[^@]*@/:***@/'
echo ""

echo ""
echo "6. Verificar pg_stat_statements en PostgreSQL:"
kubectl exec -it pg-cluster-1 -n $NAMESPACE -- psql -U postgres -c "\dx" 2>/dev/null
kubectl exec -it pg-cluster-1 -n $NAMESPACE -- psql -U postgres -c "SELECT count(*) FROM pg_stat_statements LIMIT 1;" 2>/dev/null

echo ""
echo "7. Probar endpoint de métricas:"
kubectl port-forward -n $NAMESPACE svc/postgres-exporter 9187:9187 >/dev/null 2>&1 &
PF_PID=$!
sleep 2
curl -s http://localhost:9187/metrics | head -20
kill $PF_PID 2>/dev/null
