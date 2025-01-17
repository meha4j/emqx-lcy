package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/blabtm/extd/internal/gate"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var etc *viper.Viper

func init() {
	cfg := os.Getenv("CONFIG")
	sec := os.Getenv("SECRET")

	etc = viper.New()

	etc.SetDefault("gate.name", "gate")
	etc.SetDefault("gate.port", 9001)
	etc.SetDefault("gate.emqx.host", "emqx")
	etc.SetDefault("gate.emqx.port", 18083)
	etc.SetDefault("gate.emqx.retry", 5)
	etc.SetDefault("gate.emqx.timeout", "5s")
	etc.SetDefault("gate.emqx.auto.enable", true)
	etc.SetDefault("gate.emqx.auto.timeout", "30s")
	etc.SetDefault("gate.emqx.auto.adapter.port", 9100)
	etc.SetDefault("gate.emqx.auto.listener.name", "default")
	etc.SetDefault("gate.emqx.auto.listener.type", "tcp")
	etc.SetDefault("gate.emqx.auto.listener.port", 20041)

	if cfg != "" {
		etc.SetConfigFile(cfg)

		if err := etc.MergeInConfig(); err != nil {
			panic(fmt.Errorf("etc: %v", err))
		}
	}

	if sec != "" {
		etc.SetConfigFile(sec)

		if err := etc.MergeInConfig(); err != nil {
			panic(fmt.Errorf("etc: %v", err))
		}
	}
}

func main() {
	con, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: etc.GetInt("gate.port"),
	})

	if err != nil {
		panic(fmt.Errorf("net: %v", err))
	}

	if err := lookupAddr(etc); err != nil {
		panic(fmt.Errorf("lookup: %v", err))
	}

	srv := grpc.NewServer()

	if err := gate.Register(srv, etc); err != nil {
		panic(fmt.Errorf("reg: %v", err))
	}

	slog.Info("listening", "addr", etc.GetString("gate.addr"))

	if err := srv.Serve(con); err != nil {
		panic(err)
	}
}

func lookupAddr(etc *viper.Viper) error {
	res, err := net.LookupIP(etc.GetString("gate.name"))

	if err != nil {
		return fmt.Errorf("net: %v", err)
	}

	if len(res) == 0 {
		return fmt.Errorf("no record")
	}

	etc.Set("gate.addr", fmt.Sprintf("%s:%d", res[0].String(), etc.GetInt("gate.port")))

	return nil
}
