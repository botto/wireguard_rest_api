#!/bin/bash

# handle wireguard stuff required for the script
ip link add dev "${WIREGUARD_INTERFACE}" type wireguard
wg set "${WIREGUARD_INTERFACE}" listen-port 1337 private-key <(wg genkey)
ip link set up dev "${WIREGUARD_INTERFACE}"

# generate TLS key and cert
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 \
  -subj "/C=RO/ST=B/L=B/O=CG/OU=Infra/CN=CG/emailAddress=gheorghe@linux.com"

# run the webserver
/opt/wireguard_rest_api ":31337"
