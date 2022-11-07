#!/bin/bash
mkdir bin
echo "start package"
GOOS=linux go build -o bin/k8s-webshell
#GOOS=darwin go build -o bin/k8s-webshell-mac
#GOOS=windows go build -o bin/k8s-webshell.exe

# Copy
echo "start copy"
pwd
cp -rf conf bin/
cp -rf static bin/
cp -rf views bin/
cp Dockerfile bin/

# Build
echo "start build"
cd bin/
pwd
docker build -t registry.fit2cloud.com/north/k8s-webshell:master .
docker push registry.fit2cloud.com/north/k8s-webshell:master

# Delete
cd ..
rm -rf bin/
