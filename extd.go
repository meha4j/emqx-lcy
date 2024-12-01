package main

import (
	"fmt"
	"net"

	"github.com/meha4j/extd/internal/proc"
	"github.com/meha4j/extd/internal/proc/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	Port = "port"
)

func init() {
	viper.SetEnvPrefix("extd")
}

func main() {
	log, err := zap.NewProduction()
	defer log.Sync()

	if err != nil {
		panic(err)
	}

	net, err := net.Listen("tcp", fmt.Sprintf(":%v", viper.GetInt(Port)))

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
	log.Info("Listening.", zap.String("addr", net.Addr().String()))
	srv.Serve(net)
}
