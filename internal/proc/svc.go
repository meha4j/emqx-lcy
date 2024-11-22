package proc

import (
	context "context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.uber.org/zap"
)

const (
	PROTO_NAME    = "VCAS"
	PROTO_VERSION = "FINAL"
)

func NewAdapter(uri string) (ConnectionAdapterClient, error) {
	crd := insecure.NewCredentials()
	cli, err := grpc.NewClient(uri, grpc.WithTransportCredentials(crd))

	if err != nil {
		return nil, err
	}

	return NewConnectionAdapterClient(cli), nil
}

type Storage interface {
	GetClient(conn string) (*Client, bool)
	SetClient(conn string, cli *Client)
}

type service struct {
	storage Storage
	adapter ConnectionAdapterClient
	log     *zap.Logger

	UnimplementedConnectionUnaryHandlerServer
}

func NewService(storage Storage, adapter ConnectionAdapterClient, log *zap.Logger) ConnectionUnaryHandlerServer {
	return &service{
		storage: storage,
		adapter: adapter,
		log:     log,
	}
}

func (s *service) OnSocketCreated(ctx context.Context, req *SocketCreatedRequest) (*EmptySuccess, error) {
	_, err := s.adapter.Authenticate(ctx, &AuthenticateRequest{
		Conn: req.GetConn(),
		Clientinfo: &ClientInfo{
			ProtoName: PROTO_NAME,
			ProtoVer:  PROTO_VERSION,
			Clientid:  req.GetConn(),
		},
	})

	if err != nil {
		s.log.Error("Authentication error.", zap.Error(err))

		_, err := s.adapter.Close(ctx, &CloseSocketRequest{
			Conn: req.Conn,
		})

		if err != nil {
			s.log.Error("Termination error.", zap.Error(err))
		}

		return nil, nil
	}

	s.log.Info("New connection.", zap.String("conn", req.GetConn()))
	s.storage.SetClient(req.GetConn(), NewClient(req.GetConn(), s.adapter, s.log))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *SocketClosedRequest) (*EmptySuccess, error) {
	s.log.Info("Connection closed.", zap.String("conn", req.GetConn()))
	s.storage.SetClient(req.GetConn(), nil)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *ReceivedBytesRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	cli.OnReceivedBytes(ctx, req.GetBytes())

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *TimerTimeoutRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	cli.OnTimerTimeout(ctx, req.GetType())

	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *ReceivedMessagesRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	for _, msg := range req.GetMessages() {
		cli.OnReceivedMessage(ctx, msg)
	}

	return nil, nil
}
