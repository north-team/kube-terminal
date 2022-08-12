FROM alpine:latest

LABEL maintainer="kube terminal by makai"

RUN apk add -U tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime

COPY k8s-webshell /opt/kube/k8s-webshell
COPY conf /opt/kube/conf
COPY static /opt/kube/static
COPY views /opt/kube/views

WORKDIR /opt/kube
CMD ./k8s-webshell