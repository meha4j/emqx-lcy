package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/paraskun/extd/emqx"
	"github.com/paraskun/extd/internal/gate"
	"github.com/paraskun/extd/internal/hook"
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

	slog.Info("starting server. configuration:")
	cfg.DebugTo(log.Writer())

	host := cfg.GetString("extd.emqx.host")
	port := cfg.GetInt("extd.emqx.port")
	base := fmt.Sprintf("http://%s:%d/api/v5", host, port)

	cli, err := emqx.NewClient(base,
		emqx.WithUser(cfg.GetString("extd.emqx.user")),
		emqx.WithPass(cfg.GetString("extd.emqx.pass")),
		emqx.WithRetries(cfg.GetInt("extd.emqx.rmax")),
		emqx.WithTimeout(cfg.GetString("extd.emqx.tout")),
	)

	if err != nil {
		return fmt.Errorf("emqx: %v", err)
	}

	if err := cli.LookupAddress("extd"); err != nil {
		return fmt.Errorf("lookup address: %v", err)
	}

	srv := grpc.NewServer()

	if err := gate.Register(srv, cfg, gate.WithClient(cli)); err != nil {
		return fmt.Errorf("gate: %v", err)
	}

	if err := hook.Register(srv, cfg, hook.WithClient(cli)); err != nil {
		return fmt.Errorf("gate: %v", err)
	}

	net, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GetInt("extd.port")))

	if err != nil {
		return fmt.Errorf("net: %v", err)
	}

	log.Printf("listening on :%d\n", port)

	return srv.Serve(net)
}

func configure(opts []option) (*viper.Viper, error) {
	cfg := viper.New()

	cfg.SetDefault("extd.tz", time.Local.String())
	cfg.SetDefault("extd.port", 9111)
	cfg.SetDefault("extd.emqx.host", "emqx")
	cfg.SetDefault("extd.emqx.port", 18083)
	cfg.SetDefault("extd.emqx.rmax", 5)
	cfg.SetDefault("extd.emqx.tout", "15s")
	cfg.SetDefault("extd.pgsql.host", "pgsql")
	cfg.SetDefault("extd.pgsql.port", 5432)
	cfg.SetDefault("extd.gate.server.port", 9110)
	cfg.SetDefault("extd.gate.name", "exproto")
	cfg.SetDefault("extd.gate.tout", "30s")
	cfg.SetDefault("extd.gate.enable", false)
	cfg.SetDefault("extd.gate.listener.name", "default")
	cfg.SetDefault("extd.gate.listener.type", "tcp")
	cfg.SetDefault("extd.gate.listener.port", 20041)
	cfg.SetDefault("extd.hook.name", "extd")
	cfg.SetDefault("extd.hook.tout", "30s")
	cfg.SetDefault("extd.hook.trec", "60s")
	cfg.SetDefault("extd.hook.enable", false)
	cfg.SetDefault("extd.hook.action", "deny")
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
