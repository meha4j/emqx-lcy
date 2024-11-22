package auth

import (
	context "context"

	"go.uber.org/zap"
)

type service struct {
	log *zap.Logger

	UnimplementedHookProviderServer
}

func NewService(log *zap.Logger) HookProviderServer {
	return &service{
		log: log,
	}
}

func (s *service) OnProviderLoaded(context.Context, *ProviderLoadedRequest) (*LoadedResponse, error) {
	s.log.Info("Loaded. Development setup.")

	return &LoadedResponse{
		Hooks: []*HookSpec{
			{
				Name:   "message.publish",
				Topics: []string{"test"},
			},
		},
	}, nil
}

func (s *service) OnProviderUnloaded(context.Context, *ProviderUnloadedRequest) (*EmptySuccess, error) {
	s.log.Info("Unloded. Development setup.")

	return nil, nil
}

func (s *service) OnClientAuthorize(_ context.Context, req *ClientAuthorizeRequest) (*ValuedResponse, error) {
	s.log.Info("Authorization request.", zap.String("topic", req.GetTopic()))

	return &ValuedResponse{
		Type: ValuedResponse_CONTINUE,
		Value: &ValuedResponse_BoolResult{
			BoolResult: true,
		},
	}, nil
}
