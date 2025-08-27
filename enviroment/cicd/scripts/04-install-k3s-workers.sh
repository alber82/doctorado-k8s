#!/bin/bash

# Instala K3s agent en los workers y los une al cluster

source ./cluster.conf

if [[ ! -f "$TOKEN_FILE" ]]; then
  echo "❌ Token no encontrado en $TOKEN_FILE. Ejecuta 03-install-k3s-master.sh primero."
  exit 1
fi

TOKEN=$(cat "$TOKEN_FILE")
MASTER_IP="${NODE_IPS[0]}"

for i in "${!NODES[@]}"; do
  if [[ $i -eq 0 ]]; then continue; fi  # Saltar master

  NODE_NAME="${NODES[$i]}"
  NODE_IP="${NODE_IPS[$i]}"

  echo "🚀 Instalando K3s agent en $NODE_NAME ($NODE_IP)"

  ssh "$USER@$NODE_IP" "curl -sfL https://get.k3s.io | \
    INSTALL_K3S_VERSION=$K3S_VERSION K3S_URL=https://$MASTER_IP:6443 K3S_TOKEN=$TOKEN \
    sh -s - agent"

done

echo "✅ Todos los workers instalados y unidos al clúster"