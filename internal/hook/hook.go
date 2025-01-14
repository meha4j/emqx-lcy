package hook

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	"github.com/blabtm/extd/internal/api/hook"
	"github.com/spf13/viper"
)

type Instance = hook.HookProviderServer

type options struct {
	ins []Instance
}

type Option func(opts *options) error

func Use(ins Instance) Option {
	return func(opts *options) error {
		opts.ins = append(opts.ins, ins)
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

	hook.RegisterHookProviderServer(srv, &service{ins: opt.ins})

	return nil
}

type service struct {
	ins []Instance

	hook.UnimplementedHookProviderServer
}

func (s *service) OnProviderLoaded(ctx context.Context, req *hook.ProviderLoadedRequest) (*hook.LoadedResponse, error) {
	slog.Debug("hook: provider: loaded", "req", req)

	spec := make([]*hook.HookSpec, 0, len(s.ins))

	for _, i := range s.ins {
		res, err := i.OnProviderLoaded(ctx, req)

		if err != nil {
			return nil, err
		}

		spec = append(spec, res.Hooks...)
	}

	return &hook.LoadedResponse{Hooks: spec}, nil
}

func (s *service) OnProviderUnloaded(ctx context.Context, req *hook.ProviderUnloadedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: provier: unloaded", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnProviderUnloaded(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientConnect(ctx context.Context, req *hook.ClientConnectRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: client: connect", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnClientConnect(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientConnack(ctx context.Context, req *hook.ClientConnackRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: client: connack", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnClientConnack(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientConnected(ctx context.Context, req *hook.ClientConnectedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: client: connected", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnClientConnected(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientDisconnected(ctx context.Context, req *hook.ClientDisconnectedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: client: disconnected", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnClientDisconnected(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientAuthenticate(ctx context.Context, req *hook.ClientAuthenticateRequest) (*hook.ValuedResponse, error) {
	slog.Debug("hook: client: authn", "req", req)

	for _, i := range s.ins {
		res, err := i.OnClientAuthenticate(ctx, req)

		if err != nil {
			return nil, err
		}

		if res.Type == hook.ValuedResponse_STOP_AND_RETURN {
			return res, nil
		}
	}

	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(ctx context.Context, req *hook.ClientAuthorizeRequest) (*hook.ValuedResponse, error) {
	slog.Debug("hook: client: authz", "req", req)

	for _, i := range s.ins {
		res, err := i.OnClientAuthorize(ctx, req)

		if err != nil {
			return nil, err
		}

		if res.Type == hook.ValuedResponse_STOP_AND_RETURN {
			return res, nil
		}
	}

	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientSubscribe(ctx context.Context, req *hook.ClientSubscribeRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: client: sub", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnClientSubscribe(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientUnsubscribe(ctx context.Context, req *hook.ClientUnsubscribeRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: client: usub", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnClientUnsubscribe(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionCreated(ctx context.Context, req *hook.SessionCreatedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: created", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionCreated(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionSubscribed(ctx context.Context, req *hook.SessionSubscribedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: sub", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionSubscribed(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionUnsubscribed(ctx context.Context, req *hook.SessionUnsubscribedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: usub", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionUnsubscribed(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionResumed(ctx context.Context, req *hook.SessionResumedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: resumed", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionResumed(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionDiscarded(ctx context.Context, req *hook.SessionDiscardedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: discarded", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionDiscarded(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionTakenover(ctx context.Context, req *hook.SessionTakenoverRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: takenover", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionTakenover(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnSessionTerminated(ctx context.Context, req *hook.SessionTerminatedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: session: terminated", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnSessionTerminated(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessagePublish(ctx context.Context, req *hook.MessagePublishRequest) (*hook.ValuedResponse, error) {
	slog.Debug("hook: message: pub", "req", req)

	for _, i := range s.ins {
		res, err := i.OnMessagePublish(ctx, req)

		if err != nil {
			return nil, err
		}

		if res.Type == hook.ValuedResponse_STOP_AND_RETURN {
			return res, nil
		}
	}

	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_Message{Message: req.Message},
	}, nil
}

func (s *service) OnMessageDelivered(ctx context.Context, req *hook.MessageDeliveredRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: message: delivered", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnMessageDelivered(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessageDropped(ctx context.Context, req *hook.MessageDroppedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: message: dropped", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnMessageDropped(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessageAcked(ctx context.Context, req *hook.MessageAckedRequest) (*hook.EmptySuccess, error) {
	slog.Debug("hook: message: acked", "req", req)

	for _, i := range s.ins {
		if _, err := i.OnMessageAcked(ctx, req); err != nil {
			return nil, err
		}
	}

	return &hook.EmptySuccess{}, nil
}
