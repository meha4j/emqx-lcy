package proc

import (
	context "context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.uber.org/zap"
)

var storage InMemStorage

type service struct {
	cac ConnectionAdapterClient
	log *zap.Logger

	UnimplementedConnectionUnaryHandlerServer
}

func NewService(uri string, log *zap.Logger) (ConnectionUnaryHandlerServer, error) {
	crd := insecure.NewCredentials()
	cli, err := grpc.NewClient(uri, grpc.WithTransportCredentials(crd))

	if err != nil {
		return nil, err
	}

	return &service{
		cac: NewConnectionAdapterClient(cli),
		log: log,
	}, nil
}

func (s *service) OnSocketCreated(_ context.Context, req *SocketCreatedRequest) (*EmptySuccess, error) {
	storage.Set(req.GetConn(), NewClient(req.GetConn(), s.cac, s.log))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *SocketClosedRequest) (*EmptySuccess, error) {
	storage.Set(req.GetConn(), nil)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *ReceivedBytesRequest) (*EmptySuccess, error) {
	cli, ok := storage.Get(req.GetConn())

	if !ok {
		panic("Client supposed to be exists.")
	}

	cli.OnReceivedBytes(ctx, req.GetBytes())

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *TimerTimeoutRequest) (*EmptySuccess, error) {
	cli, ok := storage.Get(req.GetConn())

	if !ok {
		panic("Client supposed to be exists.")
	}

	cli.OnTimerTimeout(ctx, req.GetType())

	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *ReceivedMessagesRequest) (*EmptySuccess, error) {
	cli, ok := storage.Get(req.GetConn())

	if !ok {
		panic("Client supposed to be exists.")
	}

	for _, msg := range req.GetMessages() {
		cli.OnReceivedMessage(ctx, msg)
	}

	return nil, nil
}
