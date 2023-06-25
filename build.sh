#!/bin/bash
mkdir bin
echo "start package"
GOOS=linux go build -o bin/kube-terminal
#GOOS=darwin go build -o bin/kube-terminal-mac
#GOOS=windows go build -o bin/kube-terminal.exe

# Copy
echo "start copy"
pwd
cp -rf conf bin/
cp -rf static bin/
cp -rf views bin/
cp start.sh bin/
cp Dockerfile bin/

# Build
echo "start build"
cd bin/
pwd
docker build -t registry.fit2cloud.com/north/kube-terminal:dev .
docker push registry.fit2cloud.com/north/kube-terminal:dev

# Delete
cd ..
rm -rf bin/
