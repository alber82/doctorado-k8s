#!/bin/bash

# Despliega charts personalizados usando Helm desde el nodo bootstrap

source ./cluster.conf

# Asegurar que KUBECONFIG está presente y actualizado
export KUBECONFIG=~/.kube/config

if [[ ! -f "$KUBECONFIG" ]]; then
  echo "❌ No se encontró kubeconfig en $KUBECONFIG. Ejecuta 05-install-cilium.sh primero."
  exit 1
fi

# Asegurar que helm está instalado
if ! command -v helm &>/dev/null; then
  echo "⚙️ Instalando Helm..."
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

# Añadir repos necesarios
helm repo add influxdata https://helm.influxdata.com/ || true
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
helm repo add grafana https://grafana.github.io/helm-charts || true
helm repo update

# Crear namespace monitoring si no existe
kubectl get ns monitoring >/dev/null 2>&1 || kubectl create ns monitoring

# Desplegar InfluxDB 2
helm upgrade --install influxdb influxdata/influxdb2 -n monitoring -f charts/values-influxdb.yaml

# Crear organización y bucket
influx org create --name uclm --description "uclm"
influx bucket create --name doctorado --org uclm --retention 72h

# Desplegar Telegraf
helm upgrade --install telegraf influxdata/telegraf -n monitoring -f charts/values-telegraf.yaml

# Desplegar Prometheus stack
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace -f charts/values-prometheus-stack.yaml

# Desplegar K6 operator
helm upgrade --install k6-operator grafana/k6-operator -f charts/values.yaml -n monitoring