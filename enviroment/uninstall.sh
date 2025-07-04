
#desinstalacion charts de monitoring
helm uninstall -n monitoring kube-prometheus-stack
helm uninstall -n monitoring telegraf
helm uninstall -n monitoring influxdb

curl https://raw.githubusercontent.com/grafana/k6-operator/main/bundle.yaml | kubectl delete -f -

kubectl delete configmap postgres-stress-test
kubectl delete configmap kafka-stress-test