kubectl create secret generic prometheus-tls \
  --from-file=tls.crt=tls/prometheus.crt \
  --from-file=tls.key=tls/prometheus.key \
  -n monitoring

kubectl create secret generic telegraf-ca \
  --from-file=ca.crt=tls/ca.crt \
  -n monitoring

kubectl create configmap node-exporter-web-config \
  --from-file=web-config.yml=web-config.yml \
  -n monitoring