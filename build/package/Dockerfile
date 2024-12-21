FROM golang:latest AS builder

WORKDIR /app

RUN mkdir ./internal
RUN mkdir ./pkg
RUN mkdir ./configs

COPY internal/ ./internal
COPY pkg/ ./pkg
COPY go.mod go.sum extd.go ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /extd .

FROM alpine

COPY --from=builder /extd /
COPY configs/ /etc

ENV EXTD_CONFIG=/etc/config.yaml
ENV EXTD_SECRET=/etc/secret.yaml

EXPOSE 9111
CMD ["/extd"]
