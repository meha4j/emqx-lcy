FROM golang:latest AS builder

WORKDIR /app

COPY api/ ./api
COPY pkg/ ./pkg
COPY srv/ ./srv
COPY go.mod go.sum extd.go .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /extd .

FROM alpine

RUN apk add --no-cache tzdata

COPY --from=builder /extd /
COPY config/ /etc

ENV TZ="Asia/Novosibirsk"
ENV EXTD_CONFIG="/etc/extd-config.yaml"
ENV EXTD_SECRET="/etc/extd-secret.yaml"

EXPOSE 9111
CMD /extd --cfg=$EXTD_CONFIG --sec=$EXTD_SECRET
