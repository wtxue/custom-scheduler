FROM alpine:3.10
MAINTAINER xk <xk.136@qq.com>

RUN apk add --no-cache tzdata curl busybox-extras && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai">>/etc/timezone

COPY bin/custom-scheduler /

EXPOSE 8080

ENTRYPOINT ["/custom-controller"]