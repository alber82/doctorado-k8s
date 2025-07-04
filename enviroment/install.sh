# instalacion suite monitorizacion
helm upgrade --install influxdb influxdata/influxdb2 -n monitoring -f values-influxdb.yaml

influx org create --name uclm --description "uclm"
influx bucket create --name doctorado --org uclm --retention 72h

helm upgrade --install telegraf -n monitoring -f values-telegraf.yaml influxdata/telegraf
helm install kube-prometheus-stack   --create-namespace   --namespace monitoring   prometheus-community/kube-prometheus-stack -f values-prometheus-stack.yaml

helm uninstall kube-prometheus-stack    --namespace monitoring


# instalacion k6 operator
helm install k6-operator grafana/k6-operator -f values.yaml
