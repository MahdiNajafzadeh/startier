#!/bin/bash

set -e

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <openssl-config-file>"
    exit 1
fi

CONFIG=$1

if [ ! -f "$CONFIG" ]; then
    echo "Config file not found: $CONFIG"
    exit 1
fi

CA_KEY="ca.key"
CA_CERT="ca.crt"
NODE_KEY="node.key"
NODE_CERT="node.crt"
NODE_CSR="node.csr"

openssl genrsa -out $CA_KEY 2048
openssl req -x509 -new -nodes -key $CA_KEY -sha256 -days 3650 -out $CA_CERT -subj "/CN=My Root CA"
openssl genrsa -out $NODE_KEY 2048
openssl req -new -key $NODE_KEY -out $NODE_CSR -config $CONFIG
openssl x509 -req -in $NODE_CSR -CA $CA_CERT -CAkey $CA_KEY -CAcreateserial -out $NODE_CERT -days 365 -sha256 -extensions v3_req -extfile $CONFIG
echo "CA: $CA_CERT"
echo "Node Key: $NODE_KEY"
echo "Node Cert: $NODE_CERT"
