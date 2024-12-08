FROM golang:latest

WORKDIR /app

COPY go.mod go.sum extd.go ./

RUN mkdir ./internal
RUN mkdir ./pkg

COPY internal/ ./internal
COPY pkg/ ./pkg

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /extd .

EXPOSE 9111

CMD ["/extd"]
