package main

import (
	"fmt"
	"net"

	"github.com/paraskun/extd/internal/proc"
	"github.com/paraskun/extd/internal/proc/proto"
	"github.com/spf13/viper"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	viper.SetDefault("extd.config", "/etc/config.yaml")
	viper.SetDefault("extd.secret", "/etc/secret.yaml")
	viper.SetDefault("extd.log", "debug")

	viper.BindEnv("extd.config")
	viper.BindEnv("extd.secret")

	viper.SetDefault("extd.port", 9111)

	viper.SetDefault("extd.emqx.host", "emqx")
	viper.SetDefault("extd.emqx.port", 18083)
	viper.SetDefault("extd.emqx.timeout", "15s")
}

func main() {
	cfg := zap.NewProductionConfig()

	cfg.Development = true
	cfg.Encoding = "console"

	if lvl, err := zap.ParseAtomicLevel(viper.GetString("extd.log")); err != nil {
		cfg.Level.SetLevel(zap.InfoLevel)
	} else {
		cfg.Level.SetLevel(lvl.Level())
	}

	log, err := cfg.Build()

	if err != nil {
		panic(fmt.Errorf("logger: %v", err))
	}

	viper.SetConfigFile(viper.GetString("extd.config"))

	if err := viper.ReadInConfig(); err != nil {
		log.Sugar().Errorf("config file (using defaults): %v", err)
	}

	viper.SetConfigFile(viper.GetString("extd.secret"))

	if err := viper.MergeInConfig(); err != nil {
		log.Sugar().Panicf("secret file: %v", err)
	}

	port := viper.GetInt("extd.port")
	net, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})

	if err != nil {
		log.Sugar().Panicf("network: %v", err)
	}

	defer net.Close()

	srv := grpc.NewServer()
	proc, err := proc.NewService(log.With(zap.String("svc", "proc")).Sugar())

	if err != nil {
		log.Sugar().Panicf("proc: %v", err)
	}

	proto.RegisterConnectionUnaryHandlerServer(srv, proc)
	log.Sugar().Infof("listener started at :%d", port)

	if err := srv.Serve(net); err != nil {
		log.Sugar().Errorf("listener: %v", err)
	}
}
