FROM golang:alpine3.18 as build

WORKDIR /opt/helm-api

COPY . .

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

RUN (go get -v -t -d ./...) && (go build -o /kube-terminal)

FROM alpine:latest

LABEL maintainer="kube terminal by makai"

RUN apk add -U tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime && \
    mkdir /opt/kube

COPY --from=build /kube-terminal /opt/kube/kube-terminal
COPY conf /opt/kube/conf
COPY static /opt/kube/static
COPY views /opt/kube/views

WORKDIR /opt/kube
CMD ./kube-terminal