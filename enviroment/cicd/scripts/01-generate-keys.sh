#!/bin/bash

# Este script genera claves SSH en cada nodo y recoge las claves públicas en local
# Para ser usado con cluster.conf

source ./cluster.conf

mkdir -p "$KEY_DIR"

for i in "${!NODES[@]}"; do
  NODE_NAME="${NODES[$i]}"
  NODE_IP="${NODE_IPS[$i]}"

  echo "🔐 [$NODE_NAME] Generando clave SSH si no existe..."
  ssh -o StrictHostKeyChecking=no "$USER@$NODE_IP" \
    '[[ -f ~/.ssh/id_ed25519 ]] || ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519'

  echo "📥 [$NODE_NAME] Descargando clave pública..."
  scp "$USER@$NODE_IP:~/.ssh/id_ed25519.pub" "$KEY_DIR/$NODE_NAME.pub"

done

echo "✅ Todas las claves generadas y almacenadas en $KEY_DIR"