FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN mkdir ./internal
COPY internal/ ./internal

RUN mkdir ./cmd
COPY cmd/ ./cmd

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/proc/proc.go

EXPOSE 9111

CMD ["/app/proc"]
