package hook

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/paraskun/extd/api/hook"
	"github.com/paraskun/extd/pkg/emqx"
	"github.com/valyala/fastjson"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
  SMT = "INSERT INTO %s (timestamp, payload) VALUES (to_timestamp(%d/1000.0), '%s')"
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

	ctx := context.Background()
	con, err := pgx.Connect(ctx, "postgres://postgres:pass@psql:5432/postgres")

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
	var stamp uint64

	json, err := fastjson.ParseBytes(req.Message.Payload)

	if err != nil {
		return nil, fmt.Errorf("json: %v", err)
	}

	obj, err := json.Object()

	if err != nil {
		return nil, fmt.Errorf("json: %v", err)
	}

	val := obj.Get("timestamp")

	if val == nil {
		stamp = req.Message.Timestamp
	} else {
		stamp = val.GetUint64()
	}

	obj.Del("timestamp")
	s.db.Exec(ctx, fmt.Sprintf(SMT, req.Message.Topic, stamp, string(obj.MarshalTo(make([]byte, 0, 60)))))

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
