kubectl create secret generic prometheus-tls \
  --from-file=tls.crt=tls/prometheus.crt \
  --from-file=tls.key=tls/prometheus.key \
  --from-file=ca.crt=tls/my-ca.crt \
  --from-file=web-config.yml=<(cat <<EOF
tls_server_config:
  cert_file: /tls/tls.crt
  key_file: /tls/tls.key
EOF
) \
  -n monitoring

kubectl create secret generic telegraf-ca \
  --from-file=ca.crt=tls/my-ca.crt \
  -n monitoring