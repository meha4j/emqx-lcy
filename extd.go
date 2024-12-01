package main

import (
	"fmt"
	"net"

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

	srv := grpc.NewServer()

	log.Info("Listening.", zap.String("addr", net.Addr().String()))
	srv.Serve(net)
}
