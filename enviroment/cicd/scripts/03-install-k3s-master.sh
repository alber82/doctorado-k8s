#!/bin/bash

# Instala el nodo master de K3s en master01
# Guarda el token en $TOKEN_FILE para que los agentes lo usen

source ./cluster.conf

MASTER_IP="${NODE_IPS[0]}"
MASTER_NAME="${NODES[0]}"

echo "🚀 Instalando K3s master en $MASTER_NAME ($MASTER_IP)"

ssh "$USER@$MASTER_IP" "curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=$K3S_VERSION sh -s - server --write-kubeconfig-mode 644"

# Esperar a que se genere el token
sleep 5

echo "📥 Obteniendo token de registro..."
ssh "$USER@$MASTER_IP" "sudo cat /var/lib/rancher/k3s/server/node-token" > "$TOKEN_FILE"

echo "✅ K3s master instalado y token guardado en $TOKEN_FILE"
