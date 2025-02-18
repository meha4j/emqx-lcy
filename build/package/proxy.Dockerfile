FROM haproxy:alpine
USER root

RUN apk --no-cache add curl

COPY configs/haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg
COPY scripts/bootstrap.sh /bootstrap

CMD [ "sh", "/bootstrap" ]