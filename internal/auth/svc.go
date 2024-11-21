package auth

import (
	context "context"
	"log"
)

type service struct {
	UnimplementedHookProviderServer
}

func NewService() HookProviderServer {
	return &service{}
}

func (*service) OnProviderLoaded(context.Context, *ProviderLoadedRequest) (*LoadedResponse, error) {
	return nil, nil
}

func (*service) OnProviderUnloaded(context.Context, *ProviderUnloadedRequest) (*EmptySuccess, error) {
	return nil, nil
}

func (*service) OnClientAuthorize(_ context.Context, req *ClientAuthorizeRequest) (*ValuedResponse, error) {
	log.Println(req)

	return &ValuedResponse{
		Type: ValuedResponse_STOP_AND_RETURN,
		Value: &ValuedResponse_BoolResult{
			BoolResult: true,
		},
	}, nil
}
