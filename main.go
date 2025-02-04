package main

import (
	"log"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/blabtm/emqx-gate/internal/gate"
	"github.com/spf13/viper"

	"google.golang.org/grpc"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	viper.SetDefault("port", 9001)
	viper.SetDefault("emqx.adapter.host", "emqx")
	viper.SetDefault("emqx.adapter.port", 9100)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	cfg := &gate.Config{}

	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer()

	if err := gate.Register(srv, cfg); err != nil {
		log.Fatal(err)
	}

	con, err := net.ListenTCP("tcp", &net.TCPAddr{Port: cfg.Port})

	if err != nil {
		log.Fatal(err)
	}

	if ips, err := net.InterfaceAddrs(); err != nil {
		log.Fatal(err)
	} else {
		for _, ip := range ips {
			log.Println(ip)
		}
	}

	if err := srv.Serve(con); err != nil {
		log.Fatal(err)
	}
}
