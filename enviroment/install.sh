# instalacion suite monitorizacion
helm upgrade --install influxdb influxdata/influxdb2 -n monitoring -f values-influxdb.yaml

influx org create --name uclm --description "uclm"
influx bucket create --name doctorado --org uclm --retention 72h

helm upgrade --install telegraf -n monitoring -f values-telegraf.yaml influxdata/telegraf
helm install kube-prometheus-stack   --create-namespace   --namespace monitoring   prometheus-community/kube-prometheus-stack -f values-prometheus-stack.yaml

# instalacion k6 operator
helm install k6-operator grafana/k6-operator -f values.yaml

# instalacion kafka
#kubectl apply -f https://strimzi.io/examples/latest/kafka/kraft/kafka-single-node.yaml -n kafka
#kubectl -n kafka apply -f 'https://strimzi.io/install/latest?namespace=kafka'
#
#
##test
#kubectl create configmap postgres-stress-test --from-file test/postgres/postgres-test.js
#kubectl create configmap kafka-stress-test --from-file test/kafka/kafka-test.js