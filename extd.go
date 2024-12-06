package main

import (
	"net"

	"github.com/meha4j/extd/internal/proc"
	"github.com/meha4j/extd/internal/proc/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	viper.SetEnvPrefix("extd")

	viper.BindEnv(proc.AdapterAddr)
}

func main() {
	log, err := zap.NewProduction()
	defer log.Sync()

	if err != nil {
		panic(err)
	}

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
