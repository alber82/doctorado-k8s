#!/bin/bash

DEPLOYMENT_NAME="nginx-deployment"
NAMESPACE="nginx"
MAX_REPLICAS=20
INCREMENT=1
INTERVAL=300  # 5 minutos en segundos

while true; do
  # Obtener el número actual de réplicas
  CURRENT_REPLICAS=$(kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o=jsonpath='{.spec.replicas}')

  # Verificar si ya alcanzamos el límite
  if [ "$CURRENT_REPLICAS" -ge "$MAX_REPLICAS" ]; then
    echo "Se alcanzó el número máximo de réplicas: $MAX_REPLICAS"
    break
  fi

  # Calcular el nuevo número de réplicas
  NEW_REPLICAS=$((CURRENT_REPLICAS + INCREMENT))

  # No exceder el máximo permitido
  if [ "$NEW_REPLICAS" -gt "$MAX_REPLICAS" ]; then
    NEW_REPLICAS=$MAX_REPLICAS
  fi

  # Escalar el Deployment
  echo "Escalando $DEPLOYMENT_NAME a $NEW_REPLICAS réplicas..."
  kubectl scale deployment $DEPLOYMENT_NAME --replicas=$NEW_REPLICAS -n $NAMESPACE

  # Esperar el intervalo antes de la siguiente iteración
  sleep $INTERVAL
done
