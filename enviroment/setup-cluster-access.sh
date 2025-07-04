#!/bin/bash

set -euo pipefail

# Definir nodos y IPs
declare -A NODES=(
  [bootstrap]="192.168.1.229"
  [master01]="192.168.1.220"
  [worker01]="192.168.1.221"
  [worker02]="192.168.1.222"
  [worker03]="192.168.1.223"
  [worker04]="192.168.1.224"
  [worker05]="192.168.1.225"
)

# Usuario SSH (ajusta si no es ubuntu)
SSH_USER="doctorado"

echo "🔑 Generando claves públicas en todos los nodos..."
for name in "${!NODES[@]}"; do
  ip="${NODES[$name]}"
  echo "➡️ $name ($ip): generando clave si no existe"
  ssh "${SSH_USER}@${ip}" '[[ -f ~/.ssh/id_ed25519 ]] || ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519'
done

mkdir -p ./node_keys

echo "📥 Recogiendo claves públicas..."
for name in "${!NODES[@]}"; do
  ip="${NODES[$name]}"
  scp "${SSH_USER}@${ip}:~/.ssh/id_ed25519.pub" "./node_keys/${name}.pub"
done

echo "📤 Distribuyendo claves a todos los nodos..."
for target_name in "${!NODES[@]}"; do
  target_ip="${NODES[$target_name]}"
  for source_name in "${!NODES[@]}"; do
    cat "./node_keys/${source_name}.pub" | ssh "${SSH_USER}@${target_ip}" "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
  done
done

echo "📓 Actualizando /etc/hosts en todos los nodos..."
HOSTS_ENTRIES=""
for name in "${!NODES[@]}"; do
  HOSTS_ENTRIES+="${NODES[$name]} ${name}\n"
done

for name in "${!NODES[@]}"; do
  ip="${NODES[$name]}"
  ssh "${SSH_USER}@${ip}" "sudo bash -c 'grep -v \"# cluster nodes\" /etc/hosts > /tmp/hosts && \
    echo \"# cluster nodes\" >> /tmp/hosts && echo -e \"$HOSTS_ENTRIES\" >> /tmp/hosts && mv /tmp/hosts /etc/hosts'"
done

echo "✅ Todo listo: acceso SSH mutuo configurado y /etc/hosts actualizado."