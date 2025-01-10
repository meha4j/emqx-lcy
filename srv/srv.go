package srv

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/paraskun/extd/pkg/emqx"
	"github.com/paraskun/extd/srv/gate"
	"github.com/paraskun/extd/srv/hook"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type options struct {
	config *string
	secret *string
}

type Option func(opts *options) error

func WithConfig(c string) Option {
	return func(opts *options) error {
		opts.config = &c
		return nil
	}
}

func WithSecret(s string) Option {
	return func(opts *options) error {
		opts.secret = &s
		return nil
	}
}

func StartServer(opts ...Option) error {
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

func configure(opts []Option) (*viper.Viper, error) {
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

	if options.config != nil {
		cfg.SetConfigFile(*options.config)

		if err := cfg.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge: %v", err)
		}
	}

	if options.secret != nil {
		cfg.SetConfigFile(*options.secret)

		if err := cfg.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("merge: %v", err)
		}
	}

	return cfg, nil
}
