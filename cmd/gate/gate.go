package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/blabtm/emqx-gate/internal/gate"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var etc *viper.Viper

func init() {
	cfg := os.Getenv("CONFIG")
	sec := os.Getenv("SECRET")

	etc = viper.New()

	etc.SetDefault("extd.gate.name", "gate")
	etc.SetDefault("extd.gate.port", 9001)
	etc.SetDefault("emqx.host", "emqx")
	etc.SetDefault("emqx.port", 18083)
	etc.SetDefault("emqx.retry", 5)
	etc.SetDefault("emqx.delay", "5s")
	etc.SetDefault("extd.gate.emqx.auto.enable", true)
	etc.SetDefault("extd.gate.emqx.auto.timeout", "30s")
	etc.SetDefault("extd.gate.emqx.auto.adapter.port", 9100)

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
		Port: etc.GetInt("extd.gate.port"),
	})

	if err != nil {
		log.Fatalf("net: %v", err)
	}

	if err := lookupAddr(etc); err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	if err := gate.Register(srv, etc); err != nil {
		log.Fatalf("reg: %v", err)
	}

	slog.Info("listening", "addr", etc.GetString("extd.gate.addr"))

	if err := srv.Serve(con); err != nil {
		panic(err)
	}
}

func lookupAddr(etc *viper.Viper) error {
	res, err := net.LookupIP(etc.GetString("extd.gate.name"))

	if err != nil {
		return err
	}

	if len(res) == 0 {
		return fmt.Errorf("no record")
	}

	etc.Set("extd.gate.addr", fmt.Sprintf("%s:%d", res[0].String(), etc.GetInt("extd.gate.port")))

	return nil
}
