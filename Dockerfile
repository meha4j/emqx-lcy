FROM golang:latest AS builder

WORKDIR /app

COPY vcas/ ./vcas
COPY emqx/ ./emqx
COPY internal/ ./internal
COPY go.mod go.sum main.go .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /cmd .

FROM alpine

RUN apk add --no-cache tzdata

COPY --from=builder /cmd /
COPY etc/ /etc

ENV TZ="Asia/Novosibirsk"
ENV EXTD_CONFIG="/etc/extd-config.yaml"
ENV EXTD_SECRET="/etc/extd-secret.yaml"

EXPOSE 9001

CMD /cmd --cfg=$EXTD_CONFIG --sec=$EXTD_SECRET
