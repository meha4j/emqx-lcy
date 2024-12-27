package main

import (
	"flag"

	"github.com/paraskun/extd/srv"
)

var cfg string
var sec string

func init() {
	flag.StringVar(&cfg, "cfg", "/etc/config.yaml", "")
	flag.StringVar(&sec, "sec", "/etc/secret.yaml", "")
}

func main() {
	flag.Parse()

	err := srv.StartServer(srv.WithConfig(cfg), srv.WithSecret(sec))

	if err != nil {
		panic(err)
	}
}
