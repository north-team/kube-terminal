#!/bin/bash
mkdir bin
GOOS=linux go build -o bin/k8s-webshell
#GOOS=darwin go build -o bin/k8s-webshell-mac
#GOOS=windows go build -o bin/k8s-webshell.exe

# Copy
cp k8s-webshell bin/
cp conf bin/
cp static bin/
cp views bin/

# Build
cd bin/
docker build -t registry.fit2cloud.com/north/k8s-webshell:master .
docker push registry.fit2cloud.com/north/k8s-webshell:master

# Delete
cd ..
rm -rf bin/
