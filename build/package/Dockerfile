FROM golang:latest

WORKDIR /app

COPY go.mod go.sum main.go ./

RUN mkdir ./internal
COPY internal/ ./internal

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /cmd .

EXPOSE 9111

CMD ["/cmd"]
