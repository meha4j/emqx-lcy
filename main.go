package main

import (
	"fmt"
	"net"

	"github.com/meha4j/extd/internal/auth"
	"github.com/meha4j/extd/internal/mem"
	"github.com/meha4j/extd/internal/proc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	PORT              = "port"
	PROC_ADAPTER_ADDR = "proc.adapter.addr"
)

func init() {
	viper.SetEnvPrefix("extd")

	viper.BindEnv(PORT)
	viper.BindEnv(PROC_ADAPTER_ADDR)

	viper.SetDefault(PORT, 9111)
	viper.SetDefault(PROC_ADAPTER_ADDR, "10.5.0.5:9110")
}

func main() {
	log, err := zap.NewProduction()
	defer log.Sync()

	if err != nil {
		panic(err)
	}

	net, err := net.Listen("tcp", fmt.Sprintf(":%v", viper.GetInt(PORT)))

	if err != nil {
		panic(err)
	}

	defer net.Close()

	storage := mem.NewStorage()
	adapter, err := proc.NewAdapter(viper.GetString(PROC_ADAPTER_ADDR))

	if err != nil {
		panic(err)
	}

	rpc := grpc.NewServer()

	proc.RegisterConnectionUnaryHandlerServer(rpc, proc.NewService(
		storage,
		adapter,
		log.With(zap.String("svc", "proc")),
	))
	auth.RegisterHookProviderServer(rpc, auth.NewService(log.With(zap.String("svc", "auth"))))

	log.Info("Listening.", zap.String("addr", net.Addr().String()))
	rpc.Serve(net)
}
