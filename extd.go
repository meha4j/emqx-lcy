package main

import (
	"net"

	"github.com/paraskun/extd/internal/proc"
	"github.com/paraskun/extd/internal/proc/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log, err := zap.NewProduction()

	if err != nil {
		panic(err)
	}

	defer log.Sync()

	net, err := net.Listen("tcp", ":80")

	if err != nil {
		panic(err)
	}

	defer net.Close()

	srv := grpc.NewServer()
	svc, err := proc.NewService(log.With(zap.String("svc", "proc")))

	if err != nil {
		panic(err)
	}

	proto.RegisterConnectionUnaryHandlerServer(srv, svc)
	srv.Serve(net)
}
