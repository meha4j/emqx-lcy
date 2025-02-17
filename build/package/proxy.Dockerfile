FROM golang:1.24.0-alpine3.21 AS build

COPY scripts/bootstrap.go .
RUN GOOS=linux go build -o /bootstrap bootstrap.go

FROM haproxy:3.1.3-alpine3.21

COPY configs/haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg
COPY --from=build /bootstrap /

CMD ["/bootstrap"]