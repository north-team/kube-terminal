FROM alpine:latest

LABEL maintainer="kube terminal by makai"

RUN apk add -U tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime && \
    mkdir /opt/kube

COPY kube-terminal /opt/kube/kube-terminal
COPY conf /opt/kube/conf
COPY static /opt/kube/static
COPY views /opt/kube/views

WORKDIR /opt/kube
CMD ./kube-terminal