FROM golang:latest AS builder

WORKDIR /app

RUN mkdir ./internal
RUN mkdir ./pkg

COPY go.mod go.sum extd.go ./
COPY internal/ ./internal
COPY pkg/ ./pkg

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /extd .

FROM alpine

COPY --from=builder /extd /

EXPOSE 9111

CMD ["/extd"]
