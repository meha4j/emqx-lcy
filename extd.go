package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/blabtm/extd/emqx"
	"github.com/blabtm/extd/internal/gate"
	"github.com/blabtm/extd/internal/hook"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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

	if err := start(withConfig(cfg), withSecret(sec)); err != nil {
		slog.Error("srv", "err", err)
		os.Exit(-1)
	}
}

type options struct {
	cfg *string
	sec *string
}

type option func(opts *options) error

func withConfig(c string) option {
	return func(opts *options) error {
		opts.cfg = &c
		return nil
	}
}

func withSecret(s string) option {
	return func(opts *options) error {
		opts.sec = &s
		return nil
	}
}

func start(opts ...option) error {
	cfg, err := configure(opts)

	if err != nil {
		return fmt.Errorf("cfg: %v", err)
	}

	var cli *emqx.Client = nil
	addr, err := net.LookupAddr(cfg.GetString("extd.lookup.name"))

	if err != nil {
		return fmt.Errorf("addr lookup: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("addr lookup: no record", err)
	}

  slog.Info("starting extd instance", "lookup", addr[0])

	if cfg.GetBool("extd.emqx.auto") {
		host := cfg.GetString("extd.emqx.host")
		port := cfg.GetInt("extd.emqx.port")
		base := fmt.Sprintf("http://%s:%d/api/v5", host, port)

		cli, err = emqx.NewClient(base,
			emqx.WithUser(cfg.GetString("extd.emqx.user")),
			emqx.WithPass(cfg.GetString("extd.emqx.pass")),
			emqx.WithRetries(cfg.GetInt("extd.emqx.rmax")),
			emqx.WithTimeout(cfg.GetString("extd.emqx.tout")),
			emqx.WithAddr(fmt.Sprintf("http://%s:%d", addr[0], cfg.GetInt("extd.port"))),
		)

		if err != nil {
			return fmt.Errorf("emqx: cli: %v", err)
		}
	}

	srv := grpc.NewServer()

	if err := gate.Register(srv, cfg, gate.WithClient(cli)); err != nil {
		return fmt.Errorf("gate: %v", err)
	}

	if err := hook.Register(srv, cfg, hook.WithClient(cli)); err != nil {
		return fmt.Errorf("hook: %v", err)
	}

  port := cfg.GetInt("extd.port")
	net, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return fmt.Errorf("net: %v", err)
	}

	log.Printf("listening on :%d\n", port)

	return srv.Serve(net)
}

func configure(opts []option) (*viper.Viper, error) {
	cfg := viper.New()

	cfg.SetDefault("extd.port", 9001)
  cfg.SetDefault("extd.emqx.auto", true)
	cfg.SetDefault("extd.emqx.host", "emqx")
	cfg.SetDefault("extd.emqx.port", 18083)
	cfg.SetDefault("extd.emqx.rmax", 5)
	cfg.SetDefault("extd.emqx.tout", "15s")
	cfg.SetDefault("extd.pgsql.host", "pgsql")
	cfg.SetDefault("extd.pgsql.port", 5432)
	cfg.SetDefault("extd.gate.adapter.port", 9100)
	cfg.SetDefault("extd.hook.pgsql.name", "postgres")

	var options options

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, fmt.Errorf("options: %v", err)
		}
	}

	if options.cfg != nil {
		cfg.SetConfigFile(*options.cfg)

		if err := cfg.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge: %v", err)
		}
	}

	if options.sec != nil {
		cfg.SetConfigFile(*options.sec)

		if err := cfg.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge: %v", err)
		}
	}

	return cfg, nil
}
