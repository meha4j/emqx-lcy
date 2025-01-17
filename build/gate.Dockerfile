FROM golang:latest AS builder

WORKDIR /app

COPY vcas/ ./vcas
COPY emqx/ ./emqx
COPY internal/gate/ ./internal/gate/
COPY internal/api/ ./internal/api/
COPY cmd/gate/ ./cmd/gate/
COPY go.mod go.sum .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /cmd ./cmd/gate

FROM alpine

RUN apk add --no-cache tzdata

COPY --from=builder /cmd /

ENV TZ="Asia/Novosibirsk"
EXPOSE 9001

CMD /cmd
