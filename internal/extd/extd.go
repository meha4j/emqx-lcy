package extd

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/blabtm/extd/emqx"
	"github.com/blabtm/extd/internal/gate"
	"github.com/blabtm/extd/internal/hook"
	"github.com/blabtm/extd/internal/hook/authz"
	"github.com/blabtm/extd/internal/hook/store"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type options struct {
	cfg *string
	sec *string
}

type Option func(opts *options) error

func WithConfig(c string) Option {
	return func(opts *options) error {
		opts.cfg = &c
		return nil
	}
}

func WithSecret(s string) Option {
	return func(opts *options) error {
		opts.sec = &s
		return nil
	}
}

func Start(opts ...Option) error {
	cfg, err := configure(opts)

	if err != nil {
		return fmt.Errorf("cfg: %v", err)
	}

	slog.Info("starting extd instance", "addr", cfg.GetString("extd.addr"))

	if cfg.GetBool("extd.emqx.auto") {
		host := cfg.GetString("extd.emqx.host")
		port := cfg.GetInt("extd.emqx.port")
		base := fmt.Sprintf("http://%s:%d/api/v5", host, port)

		slog.Info("updating emqx configuration", "base", base)

		cli, err := emqx.NewClient(base,
			emqx.WithUser(cfg.GetString("extd.emqx.user")),
			emqx.WithPass(cfg.GetString("extd.emqx.pass")),
			emqx.WithRetries(cfg.GetInt("extd.emqx.rmax")),
			emqx.WithTimeout(cfg.GetString("extd.emqx.tout")),
		)

		if err != nil {
			return fmt.Errorf("emqx: cli: %v", err)
		}

		if err := cli.UpdateGate(newGateConfig(cfg)); err != nil {
			return fmt.Errorf("emqx: cli: upd: %v", err)
		}

		if err := cli.UpdateHook(newHookConfig(cfg)); err != nil {
			return fmt.Errorf("emqx: cli: upd: %v", err)
		}
	}

	srv := grpc.NewServer()

	if err := gate.Register(srv, cfg); err != nil {
		return fmt.Errorf("gate: %v", err)
	}

	if err := hook.Register(srv, cfg,
		hook.Use(authz.New(nil)),
		hook.Use(store.New(nil)),
	); err != nil {
		return fmt.Errorf("hook: %v", err)
	}

	net, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GetInt("extd.port")))

	if err != nil {
		return fmt.Errorf("net: %v", err)
	}

	slog.Info("server started")

	return srv.Serve(net)
}

func configure(opts []Option) (*viper.Viper, error) {
	cfg := viper.New()

	cfg.SetDefault("extd.port", 9001)
	cfg.SetDefault("extd.lookup.name", "extd")
	cfg.SetDefault("extd.emqx.auto", true)
	cfg.SetDefault("extd.emqx.host", "emqx")
	cfg.SetDefault("extd.emqx.port", 18083)
	cfg.SetDefault("extd.emqx.rmax", 5)
	cfg.SetDefault("extd.emqx.tout", "15s")
	cfg.SetDefault("extd.pgsql.host", "pgsql")
	cfg.SetDefault("extd.pgsql.port", 5432)
	cfg.SetDefault("extd.gate.tout", "300s")
	cfg.SetDefault("extd.gate.enable", false)
	cfg.SetDefault("extd.gate.statistics", true)
	cfg.SetDefault("extd.gate.server.port", 9100)
	cfg.SetDefault("extd.gate.listener.name", "vcas")
	cfg.SetDefault("extd.gate.listener.type", "tcp")
	cfg.SetDefault("extd.gate.listener.port", 20041)
	cfg.SetDefault("extd.hook.pgsql.name", "postgres")
	cfg.SetDefault("extd.hook.name", "extd")
	cfg.SetDefault("extd.hook.enable", false)
	cfg.SetDefault("extd.hook.tout", "15s")
	cfg.SetDefault("extd.hook.rout", "15s")
	cfg.SetDefault("extd.hook.action", "deny")
	cfg.SetDefault("extd.hook.pool.size", 16)

	var options options

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, fmt.Errorf("etc: %v", err)
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

	host, err := net.LookupAddr(cfg.GetString("extd.lookup.name"))

	if err != nil {
		return nil, fmt.Errorf("lookup: net: %v", err)
	}

	if len(host) == 0 {
		return nil, fmt.Errorf("lookup: no record")
	}

	cfg.Set("extd.addr", fmt.Sprintf("http://%s:%d", host[0], cfg.GetInt("extd.port")))

	return cfg, nil
}

func newGateConfig(cfg *viper.Viper) *emqx.GateUpdateRequest {
	return &emqx.GateUpdateRequest{
		Name:       "exproto",
		Timeout:    cfg.GetString("extd.gate.tout"),
		Mountpoint: "",
		Enable:     cfg.GetBool("extd.gate.enable"),
		Statistics: cfg.GetBool("extd.gate.statistics"),
		Server: emqx.Server{
			Bind: strconv.Itoa(cfg.GetInt("extd.gate.server.port")),
		},
		Handler: emqx.Handler{
			Addr: cfg.GetString("extd.addr"),
		},
		Listeners: []emqx.Listener{
			{
				Name: cfg.GetString("extd.gate.listener.name"),
				Type: cfg.GetString("extd.gate.listener.type"),
				Bind: strconv.Itoa(cfg.GetInt("extd.gate.listener.port")),
			},
		},
	}
}

func newHookConfig(cfg *viper.Viper) *emqx.HookUpdateRequest {
	return &emqx.HookUpdateRequest{
		Name:      cfg.GetString("extd.hook.name"),
		Enable:    cfg.GetBool("extd.hook.enable"),
		Addr:      cfg.GetString("extd.addr"),
		Timeout:   cfg.GetString("extd.hook.tout"),
		Action:    cfg.GetString("extd.hook.action"),
		Reconnect: cfg.GetString("extd.hook.rout"),
		PoolSize:  cfg.GetInt("extd.hook.pool.size"),
	}
}
