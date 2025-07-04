#!/bin/bash

# Despliega Cilium como CNI en el clúster desde el nodo bootstrap

source ./cluster.conf

# Asegurarse de tener kubectl configurado con el master
export KUBECONFIG=~/.kube/config

MASTER_IP="${NODE_IPS[0]}"

# Copiar kubeconfig del master si no existe localmente
if [[ ! -f "$KUBECONFIG" ]]; then
  echo "📥 Copiando kubeconfig desde $MASTER_IP"
  scp "$USER@$MASTER_IP:/etc/rancher/k3s/k3s.yaml" "$KUBECONFIG"
  sed -i "s/127.0.0.1/$MASTER_IP/" "$KUBECONFIG"
fi

# Instalar Cilium CLI si no está presente
if ! command -v cilium &>/dev/null; then
  echo "⚙️ Instalando cilium CLI..."
  curl -sSL --remote-name-all https://github.com/cilium/cilium-cli/releases/latest/download/cilium-linux-arm64.tar.gz
  sudo tar xzvf cilium-linux-arm64.tar.gz -C /usr/local/bin
  rm cilium-linux-arm64.tar.gz
fi

# Desplegar Cilium
echo "🚀 Desplegando Cilium..."
cilium install --version 1.16.2

# Verificar
cilium status --wait

echo "✅ Cilium desplegado correctamente"