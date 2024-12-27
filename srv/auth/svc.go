package auth

import (
	"context"

	"github.com/paraskun/extd/api/auth"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Register(srv *grpc.Server, cfg *viper.Viper, log *zap.Logger) error {
	return nil
}

type service struct {
	Log *zap.SugaredLogger

	auth.UnimplementedHookProviderServer
}

func newService(log *zap.SugaredLogger) (auth.HookProviderServer, error) {
	return &service{Log: log}, nil
}

func (*service) OnProviderLoaded(context.Context, *auth.ProviderLoadedRequest) (*auth.LoadedResponse, error) {
	return &auth.LoadedResponse{
		Hooks: nil,
	}, nil
}

func (*service) OnProviderUnloaded(context.Context, *auth.ProviderUnloadedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientConnect(context.Context, *auth.ClientConnectRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientConnack(context.Context, *auth.ClientConnackRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientConnected(context.Context, *auth.ClientConnectedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientDisconnected(context.Context, *auth.ClientDisconnectedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientAuthenticate(context.Context, *auth.ClientAuthenticateRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type: auth.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientAuthorize(_ context.Context, req *auth.ClientAuthorizeRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type: auth.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientSubscribe(context.Context, *auth.ClientSubscribeRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientUnsubscribe(context.Context, *auth.ClientUnsubscribeRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionCreated(context.Context, *auth.SessionCreatedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionSubscribed(context.Context, *auth.SessionSubscribedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionUnsubscribed(context.Context, *auth.SessionUnsubscribedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionResumed(context.Context, *auth.SessionResumedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionDiscarded(context.Context, *auth.SessionDiscardedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionTakenover(context.Context, *auth.SessionTakenoverRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionTerminated(context.Context, *auth.SessionTerminatedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnMessagePublish(context.Context, *auth.MessagePublishRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type: auth.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnMessageDelivered(context.Context, *auth.MessageDeliveredRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnMessageDropped(context.Context, *auth.MessageDroppedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnMessageAcked(context.Context, *auth.MessageAckedRequest) (*auth.EmptySuccess, error) {
	return nil, nil
}
