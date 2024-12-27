package srv

import (
	"fmt"
	"net"

	"github.com/paraskun/extd/srv/auth"
	"github.com/paraskun/extd/srv/proc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	cfg, err := getConfig(opts)

	if err != nil {
		return fmt.Errorf("config: %v", err)
	}

	log, err := getLogger(cfg)

	if err != nil {
		return fmt.Errorf("logger: %v", err)
	}

	srv := grpc.NewServer()

	if err := proc.Register(srv, cfg, log.With(zap.String("svc", "proc"))); err != nil {
		return fmt.Errorf("proc: %v", err)
	}

	if err := auth.Register(srv, cfg, log.With(zap.String("svc", "auth"))); err != nil {
		return fmt.Errorf("auth: %v", err)
	}

	port := cfg.GetInt("extd.port")
	net, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return fmt.Errorf("network: %v", err)
	}

	return srv.Serve(net)
}

func getConfig(opts []Option) (*viper.Viper, error) {
	var options options

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, fmt.Errorf("options: %v", err)
		}
	}

	cfg := viper.New()

	cfg.SetDefault("extd.port", 9111)
	cfg.SetDefault("extd.log", "debug")
	cfg.SetDefault("extd.emqx.host", "emqx")
	cfg.SetDefault("extd.emqx.port", 18083)
	cfg.SetDefault("extd.emqx.retry", 5)
	cfg.SetDefault("extd.emqx.timeout", "15s")
	cfg.SetDefault("extd.proc.emqx.adater.port", 9110)
	cfg.SetDefault("extd.proc.emqx.listener.port", 20041)
	cfg.SetDefault("extd.auth.pgsql.host", "pgsql")
	cfg.SetDefault("extd.auth.pgsql.port", 5432)

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

func getLogger(cfg *viper.Viper) (*zap.Logger, error) {
	conf := zap.NewProductionConfig()

	conf.Development = true
	conf.Encoding = "console"

	if lvl, err := zap.ParseAtomicLevel(cfg.GetString("extd.log")); err != nil {
		conf.Level.SetLevel(zap.InfoLevel)
	} else {
		conf.Level.SetLevel(lvl.Level())
	}

	return conf.Build()
}
