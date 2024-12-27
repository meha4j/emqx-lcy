FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum .
COPY api/ ./api
COPY cmd/ ./cmd
COPY internal/ ./internal
COPY pkg/ ./pkg
COPY srv/ ./srv

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /extd ./cmd/extd

FROM alpine

COPY --from=builder /extd /
COPY config/ /etc

ENV EXTD_CONFIG=/etc/extd-config.yaml
ENV EXTD_SECRET=/etc/extd-secret.yaml

EXPOSE 9111
CMD /extd --cfg=$EXTD_CONFIG --sec=$EXTD_SECRET
