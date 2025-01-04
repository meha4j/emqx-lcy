package srv

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"github.com/paraskun/extd/pkg/emqx"
	"github.com/paraskun/extd/srv/auth"
	"github.com/paraskun/extd/srv/proc"
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

	ctl := &auth.ACL{}
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

	if err := proc.Register(srv, ctl, cli, cfg); err != nil {
		return fmt.Errorf("proc: %v", err)
	}

	if err := auth.Register(srv, ctl, cli, cfg); err != nil {
		return fmt.Errorf("auth: %v", err)
	}

	port = cfg.GetInt("extd.port")
	net, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

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
	cfg.SetDefault("extd.proc.server.port", 9110)
	cfg.SetDefault("extd.proc.name", "vcas")
	cfg.SetDefault("extd.proc.tout", "30s")
	cfg.SetDefault("extd.proc.enable", false)
	cfg.SetDefault("extd.proc.listener.name", "default")
	cfg.SetDefault("extd.proc.listener.type", "tcp")
	cfg.SetDefault("extd.proc.listener.port", 20041)
	cfg.SetDefault("extd.auth.name", "extd")
	cfg.SetDefault("extd.auth.tout", "30s")
	cfg.SetDefault("extd.auth.trec", "60s")
	cfg.SetDefault("extd.auth.enable", false)
	cfg.SetDefault("extd.auth.action", "deny")
	cfg.SetDefault("extd.auth.pgsql.name", "postgres")

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
