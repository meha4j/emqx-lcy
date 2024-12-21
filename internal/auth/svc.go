package auth

import (
	"context"

	"github.com/paraskun/extd/internal/auth/proto"
	"go.uber.org/zap"
)

type service struct {
	Log *zap.SugaredLogger

	proto.UnimplementedHookProviderServer
}

func (*service) OnProviderLoaded(context.Context, *proto.ProviderLoadedRequest) (*proto.LoadedResponse, error) {
	return &proto.LoadedResponse{
		Hooks: nil,
	}, nil
}

func (*service) OnProviderUnloaded(context.Context, *proto.ProviderUnloadedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientConnect(context.Context, *proto.ClientConnectRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientConnack(context.Context, *proto.ClientConnackRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientConnected(context.Context, *proto.ClientConnectedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientDisconnected(context.Context, *proto.ClientDisconnectedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientAuthenticate(context.Context, *proto.ClientAuthenticateRequest) (*proto.ValuedResponse, error) {
	return &proto.ValuedResponse{
		Type: proto.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientAuthorize(_ context.Context, req *proto.ClientAuthorizeRequest) (*proto.ValuedResponse, error) {
	return &proto.ValuedResponse{
		Type: proto.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientSubscribe(context.Context, *proto.ClientSubscribeRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientUnsubscribe(context.Context, *proto.ClientUnsubscribeRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionCreated(context.Context, *proto.SessionCreatedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionSubscribed(context.Context, *proto.SessionSubscribedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionUnsubscribed(context.Context, *proto.SessionUnsubscribedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionResumed(context.Context, *proto.SessionResumedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionDiscarded(context.Context, *proto.SessionDiscardedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionTakenover(context.Context, *proto.SessionTakenoverRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnSessionTerminated(context.Context, *proto.SessionTerminatedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnMessagePublish(context.Context, *proto.MessagePublishRequest) (*proto.ValuedResponse, error) {
	return &proto.ValuedResponse{
		Type: proto.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnMessageDelivered(context.Context, *proto.MessageDeliveredRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnMessageDropped(context.Context, *proto.MessageDroppedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (*service) OnMessageAcked(context.Context, *proto.MessageAckedRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}
