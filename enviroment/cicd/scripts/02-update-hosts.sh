#!/bin/bash

# Este script actualiza /etc/hosts en todos los nodos del clúster
# Usando los nombres e IPs definidos en cluster.conf

source ./cluster.conf

HOSTS_BLOCK="# cluster nodes\n"
for i in "${!NODES[@]}"; do
  HOSTS_BLOCK+="${NODE_IPS[$i]} ${NODES[$i]}\n"
  # También puedes añadir FQDN si quieres
  # HOSTS_BLOCK+="${NODE_IPS[$i]} ${NODES[$i]}.local ${NODES[$i]}\n"
done

for i in "${!NODES[@]}"; do
  NODE_NAME="${NODES[$i]}"
  NODE_IP="${NODE_IPS[$i]}"
  echo "🖋️ [$NODE_NAME] Actualizando /etc/hosts..."
  ssh "$USER@$NODE_IP" "sudo bash -c 'grep -v "# cluster nodes" /etc/hosts > /tmp/hosts && \
    echo -e \"$HOSTS_BLOCK\" >> /tmp/hosts && mv /tmp/hosts /etc/hosts'"
done

echo "✅ /etc/hosts actualizado en todos los nodos."