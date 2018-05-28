#!/usr/bin/env bash
# https://www.accuweaver.com/2014/09/19/make-chrome-accept-a-self-signed-certificate-on-osx/

# https://gist.github.com/jessedearing/2351836

# Run using "sudo"

echo "Generating an SSL private key to sign your certificate..."
openssl genrsa -des3 -out greedgame-local.key 1024

echo "Generating a Certificate Signing Request..."
openssl req -new -key greedgame-local.key -out greedgame-local.csr

echo "Removing pass-phrase from key (for nginx)..."
cp greedgame-local.key greedgame-local.key.org
openssl rsa -in greedgame-local.key.org -out greedgame-local.key
rm greedgame-local.key.org

echo "Generating certificate..."
openssl x509 -req -days 365 -in greedgame-local.csr -signkey greedgame-local.key -out greedgame-local.crt

echo "Copying certificate (greedgame-local.crt) to /etc/ssl/certs/"
mkdir -p  /etc/ssl/certs
cp greedgame-local.crt /etc/ssl/certs/

echo "Copying key (greedgame-local.key) to /etc/ssl/private/"
mkdir -p  /etc/ssl/private
cp greedgame-local.key /etc/ssl/private/