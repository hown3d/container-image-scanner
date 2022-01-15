#!/bin/sh
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && /bin/pwd )"

rm $DIR/*.pem

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout $DIR/ca-key.pem -out $DIR/ca-cert.pem -subj "/C=DE/ST=Hessen/L=Darmstadt/O=Kevo/OU=Test/CN=*/emailAddress=ludi.origin@gmail.com"

echo "CA's self-signed certificate"
openssl x509 -in $DIR/ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout $DIR/server-key.pem -out $DIR/server-req.pem -subj "/C=DE/ST=Hessen/L=Darmstadt/O=Kevo/OU=Test/CN=*/emailAddress=ludi.origin@gmail.com"


# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in $DIR/server-req.pem -days 60 -CA $DIR/ca-cert.pem -CAkey $DIR/ca-key.pem -CAcreateserial -out $DIR/server-cert.pem
