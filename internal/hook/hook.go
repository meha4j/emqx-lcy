package hook

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
	"github.com/valyala/fastjson"

	"google.golang.org/grpc"

	"github.com/paraskun/extd/emqx"
	"github.com/paraskun/extd/internal/api/hook"
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

	store, err := newStore(
		context.Background(),
		"postgres://postgres:pass@psql:5432/postgres",
		cfg.GetUint("extd.hook.buf.qcap"),
	)

	if err != nil {
		return fmt.Errorf("store: %v", err)
	}

	hook.RegisterHookProviderServer(srv, &service{store: store})

	return nil
}

type service struct {
	store *store

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
	for _, own := range s.store.own {
		own.CompareAndSwap(req.Clientinfo.Clientid, "")
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientAuthenticate(_ context.Context, req *hook.ClientAuthenticateRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(ctx context.Context, req *hook.ClientAuthorizeRequest) (*hook.ValuedResponse, error) {
	if req.Type == hook.ClientAuthorizeRequest_PUBLISH {
    slog.Debug("authz", "top", req.Topic, "con", req.Clientinfo.Clientid)

		if !s.store.authz(req.Topic, req.Clientinfo.Clientid) {
			return &hook.ValuedResponse{
				Type:  hook.ValuedResponse_STOP_AND_RETURN,
				Value: &hook.ValuedResponse_BoolResult{BoolResult: false},
			}, nil
		}
	}

	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
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
	rec, err := parse(req.Message.Payload)

	if err != nil {
		slog.Error("parse", "msg", req.Message, "err", err)
		return nil, fmt.Errorf("parse: %v", err)
	}

	if err := s.store.save(ctx, req.Message.Topic, rec); err != nil {
		slog.Error("save", "msg", req.Message, "err", err)
		return nil, fmt.Errorf("save: %v", err)
	}

	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_Message{
      Message: req.Message,
    },
	}, nil
}

func parse(pay []byte) (rec record, err error) {
	json, err := fastjson.ParseBytes(pay)

	if err != nil {
		return rec, fmt.Errorf("json: %v", err)
	}

	obj, err := json.Object()

	if err != nil {
		return rec, fmt.Errorf("json: %v", err)
	}

	rec.stamp = obj.Get("stamp").GetUint()
	obj.Del("stamp")
	rec.payload = string(obj.MarshalTo(make([]byte, 0, 60)))

	return rec, nil

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
