package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/paraskun/extd/api/hook"
	"github.com/paraskun/extd/pkg/emqx"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type options struct {
	cli *emqx.Client
}

type Option func(opts *options) error

func WithClient(cli *emqx.Client) Option {
	return func(opts *options) error {
		opts.cli = cli
		return nil
	}
}

func Register(srv *grpc.Server, cfg *viper.Viper, opts ...Option) error {
	var opt options

	for _, exe := range opts {
		if err := exe(&opt); err != nil {
			return fmt.Errorf("opt: %v", err)
		}
	}

	if opt.cli != nil {
		if err := opt.cli.UpdateExHookServer(&emqx.ExHookServerUpdateRequest{
			Name:      cfg.GetString("extd.hook.name"),
			Enable:    cfg.GetBool("extd.hook.enable"),
			Action:    cfg.GetString("extd.hook.action"),
			Timeout:   cfg.GetString("extd.hook.tout"),
			Reconnect: cfg.GetString("extd.hook.trec"),
			Addr:      fmt.Sprintf("http://%s:%d", opt.cli.Addr, cfg.GetInt("extd.port")),
		}); err != nil {
			return fmt.Errorf("upd: %v", err)
		}
	}

	con, err := pgx.Connect(context.Background(), "postgres://postgres:pass@psql:5432/postgres")

	if err != nil {
		return fmt.Errorf("db: %v", err)
	}

	hook.RegisterHookProviderServer(srv, &service{db: con})

	return nil
}

type service struct {
	db *pgx.Conn

	hook.UnimplementedHookProviderServer
}

func (s *service) OnProviderLoaded(ctx context.Context, _ *hook.ProviderLoadedRequest) (*hook.LoadedResponse, error) {
	return &hook.LoadedResponse{
		Hooks: []*hook.HookSpec{
			{Name: "client.authorize"},
			{Name: "message.publish"},
		},
	}, nil
}

func (s *service) OnProviderUnloaded(context.Context, *hook.ProviderUnloadedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientConnect(context.Context, *hook.ClientConnectRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientConnack(context.Context, *hook.ClientConnackRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientConnected(context.Context, *hook.ClientConnectedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientDisconnected(_ context.Context, req *hook.ClientDisconnectedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientAuthenticate(_ context.Context, req *hook.ClientAuthenticateRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(_ context.Context, req *hook.ClientAuthorizeRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type: hook.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientSubscribe(context.Context, *hook.ClientSubscribeRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientUnsubscribe(context.Context, *hook.ClientUnsubscribeRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionCreated(context.Context, *hook.SessionCreatedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionSubscribed(context.Context, *hook.SessionSubscribedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionUnsubscribed(context.Context, *hook.SessionUnsubscribedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionResumed(context.Context, *hook.SessionResumedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionDiscarded(context.Context, *hook.SessionDiscardedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionTakenover(context.Context, *hook.SessionTakenoverRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionTerminated(context.Context, *hook.SessionTerminatedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessagePublish(ctx context.Context, req *hook.MessagePublishRequest) (*hook.ValuedResponse, error) {
	slog.Debug("pub", "top", req.Message.Topic, "pay", string(req.Message.Payload))

	pay := make(map[string]any)

	if err := json.Unmarshal(req.Message.Payload, &pay); err != nil {
		return nil, fmt.Errorf("pay: %v", err)
	}

	var keys strings.Builder
	var vals strings.Builder

	for key, val := range pay {
		if keys.Len() != 0 {
			keys.WriteString(", ")
			vals.WriteString(", ")
		}

		keys.WriteString(key)

		switch key {
		case "timestamp":
			vals.WriteString(fmt.Sprintf("to_timestamp(%d/1000.0) at time zone 'Asia/Novosibirsk'", uint64(val.(float64))))
		default:
			if s, ok := val.(string); ok {
				vals.WriteString(fmt.Sprintf("'%s'", s))
			} else {
				vals.WriteString(fmt.Sprintf("%v", val))
			}
		}
	}

	s.db.Exec(ctx, fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", req.Message.Topic, keys.String(), vals.String()))

	return &hook.ValuedResponse{
		Type: hook.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnMessageDelivered(context.Context, *hook.MessageDeliveredRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnMessageDropped(context.Context, *hook.MessageDroppedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessageAcked(_ context.Context, req *hook.MessageAckedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}
