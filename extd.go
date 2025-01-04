package main

import (
	"flag"
	"log/slog"

	"github.com/paraskun/extd/srv"
)

var cfg string
var sec string

func init() {
	flag.StringVar(&cfg, "cfg", "/etc/config.yaml", "Configuration file.")
	flag.StringVar(&sec, "sec", "/etc/secret.yaml", "Credentials file.")
}

func main() {
	flag.Parse()
	slog.SetLogLoggerLevel(slog.LevelDebug.Level())

	if err := srv.StartServer(srv.WithConfig(cfg), srv.WithSecret(sec)); err != nil {
		slog.Error("srv", "err", err)
		panic(err)
	}
}
