FROM kubeoperator/webkubectl

LABEL maintainer="kube terminal by makai"

RUN apk add -U tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime && \
    mkdir /opt/kube

COPY kube-terminal /opt/kube/kube-terminal
COPY conf /opt/kube/conf
COPY static /opt/kube/static
COPY views /opt/kube/views
COPY start.sh /opt/kube/

WORKDIR /opt/kube
CMD ["sh","/opt/kube/start.sh"]