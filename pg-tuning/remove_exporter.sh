#!/bin/bash

NAMESPACE="pg"

echo "🗑️  Eliminando todos los recursos del postgres-exporter..."

# 1. Eliminar ServiceMonitor
echo "Eliminando ServiceMonitor..."
kubectl delete servicemonitor postgres-exporter -n $NAMESPACE 2>/dev/null || true

# 2. Eliminar Service
echo "Eliminando Service..."
kubectl delete service postgres-exporter -n $NAMESPACE 2>/dev/null || true

# 3. Eliminar Deployment
echo "Eliminando Deployment..."
kubectl delete deployment postgres-exporter -n $NAMESPACE 2>/dev/null || true

# 4. Eliminar ConfigMap de queries
echo "Eliminando ConfigMap..."
kubectl delete configmap postgres-exporter-queries -n $NAMESPACE 2>/dev/null || true

# 5. Eliminar Secret
echo "Eliminando Secret..."
kubectl delete secret postgres-exporter-secret -n $NAMESPACE 2>/dev/null || true

# 6. Esperar a que se eliminen los pods
echo "Esperando eliminación de pods..."
kubectl wait --for=delete pod -l app=postgres-exporter -n $NAMESPACE --timeout=60s 2>/dev/null || true

echo ""
echo "✅ Limpieza completada"
echo ""
echo "Verificar que todo se eliminó:"
kubectl get all,configmap,secret,servicemonitor -n $NAMESPACE | grep postgres-exporter
