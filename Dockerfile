FROM golang:latest AS builder

WORKDIR /app

COPY vcas/ ./vcas
COPY api/ ./api
COPY internal/ ./internal
COPY go.mod go.sum main.go .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /cmd .

FROM alpine

RUN apk add --no-cache tzdata

COPY --from=builder /cmd /

ENV TZ="Asia/Novosibirsk"
EXPOSE 9001

CMD /cmd
