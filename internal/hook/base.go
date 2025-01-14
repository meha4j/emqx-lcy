package hook

import (
	"context"

	"github.com/blabtm/extd/internal/api/hook"
)

type Base struct{
  hook.UnimplementedHookProviderServer
}

func (*Base) OnProviderUnloaded(context.Context, *hook.ProviderUnloadedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnClientConnect(ctx context.Context, req *hook.ClientConnectRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnClientConnack(context.Context, *hook.ClientConnackRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnClientConnected(context.Context, *hook.ClientConnectedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnClientDisconnected(context.Context, *hook.ClientDisconnectedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnClientAuthenticate(context.Context, *hook.ClientAuthenticateRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (*Base) OnClientAuthorize(context.Context, *hook.ClientAuthorizeRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (*Base) OnClientSubscribe(context.Context, *hook.ClientSubscribeRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnClientUnsubscribe(context.Context, *hook.ClientUnsubscribeRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionCreated(context.Context, *hook.SessionCreatedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionSubscribed(context.Context, *hook.SessionSubscribedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionUnsubscribed(context.Context, *hook.SessionUnsubscribedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionResumed(context.Context, *hook.SessionResumedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionDiscarded(context.Context, *hook.SessionDiscardedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionTakenover(context.Context, *hook.SessionTakenoverRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnSessionTerminated(context.Context, *hook.SessionTerminatedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnMessagePublish(_ context.Context, req *hook.MessagePublishRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_Message{Message: req.Message},
	}, nil
}

func (*Base) OnMessageDelivered(context.Context, *hook.MessageDeliveredRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnMessageDropped(context.Context, *hook.MessageDroppedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*Base) OnMessageAcked(context.Context, *hook.MessageAckedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}
