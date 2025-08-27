#!/bin/bash

set -e

# CONFIGURACIÓN PERSONALIZABLE
CN="prometheus.monitoring.svc"
DAYS=365
DIR="tls"
CA_NAME="ca"
CERT_NAME="prometheus"

# SANs personalizados (puedes añadir más)
ALT_NAMES="
DNS.1 = ${CN}
DNS.2 = prometheus
DNS.3 = node-exporter.monitoring.svc
DNS.4 = node-exporter
DNS.5 = node-exporter.monitoring
IP.1 = 127.0.0.1
"

# CREA DIRECTORIO DE SALIDA
mkdir -p "$DIR"
cd "$DIR"

echo "🔐 Generando CA propia..."

# Clave privada de la CA
openssl genrsa -out "${CA_NAME}.key" 4096

# Certificado de la CA
openssl req -x509 -new -nodes -key "${CA_NAME}.key" -sha256 -days 3650 \
  -out "${CA_NAME}.crt" \
  -subj "/C=ES/ST=Albacete/O=uclm/CN=CA"

echo "📜 Generando CSR para ${CERT_NAME}..."

# Clave privada del servidor (Prometheus o node_exporter)
openssl genrsa -out "${CERT_NAME}.key" 4096

# CSR del servidor
openssl req -new -key "${CERT_NAME}.key" -out "${CERT_NAME}.csr" \
  -subj "/C=ES/ST=Albacete/O=uclm/CN=${CN}"

# Archivo de configuración para SANs
cat > extfile.cnf <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,nonRepudiation,keyEncipherment,dataEncipherment
subjectAltName=@alt_names

[alt_names]
${ALT_NAMES}
EOF

echo "✍️ Firmando el certificado TLS con la CA..."

# Certificado TLS firmado
openssl x509 -req -in "${CERT_NAME}.csr" -CA "${CA_NAME}.crt" -CAkey "${CA_NAME}.key" \
  -CAcreateserial -out "${CERT_NAME}.crt" -days ${DAYS} -sha256 \
  -extfile extfile.cnf

# Limpiar archivos intermedios
rm -f "${CERT_NAME}.csr" extfile.cnf "${CA_NAME}.srl"

echo "✅ Listo: certificados generados en ./${DIR}/"
echo "- CA: ${CA_NAME}.crt"
echo "- TLS Cert: ${CERT_NAME}.crt"
echo "- TLS Key: ${CERT_NAME}.key"