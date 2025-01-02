package main

import (
	"flag"

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

	if err := srv.StartServer(srv.WithConfig(cfg), srv.WithSecret(sec)); err != nil {
    panic(err)
  }
}
