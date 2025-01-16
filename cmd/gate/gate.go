package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/blabtm/extd/internal/extd"
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

	if err := extd.Start(
		extd.WithConfig(cfg),
		extd.WithSecret(sec),
	); err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
}
