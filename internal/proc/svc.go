package proc

import (
	context "context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

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
		return nil, fmt.Errorf("grpc: %v", err)
	}

	return NewConnectionAdapterClient(cli), nil
}

type Storage interface {
	GetClient(conn string) (*Client, bool)
	SetClient(conn string, cli *Client)
}

type service struct {
	Log *zap.Logger

	storage Storage
	adapter ConnectionAdapterClient

	UnimplementedConnectionUnaryHandlerServer
}

func NewService(storage Storage, adapter ConnectionAdapterClient, log *zap.Logger) ConnectionUnaryHandlerServer {
	return &service{
		Log: log,

		storage: storage,
		adapter: adapter,
	}
}

func (s *service) OnSocketCreated(ctx context.Context, req *SocketCreatedRequest) (*EmptySuccess, error) {
	log := s.Log.With(
		zap.String("conn", req.GetConn()),
		zap.String("host", req.GetConninfo().GetPeername().GetHost()),
		zap.Uint32("port", req.GetConninfo().GetPeername().GetPort()),
	)

	res, err := s.adapter.Authenticate(ctx, &AuthenticateRequest{
		Conn: req.GetConn(),
		Clientinfo: &ClientInfo{
			ProtoName: PROTO_NAME,
			ProtoVer:  PROTO_VERSION,
			Clientid:  req.GetConn(),
		},
	})

	if err != nil {
		s.Log.Error("adapter", zap.Error(err))

		s.adapter.Close(ctx, &CloseSocketRequest{
			Conn: req.Conn,
		})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.GetCode() != ResultCode_SUCCESS {
		log.Error("authentication", zap.String("error", res.GetMessage()))

		s.adapter.Close(ctx, &CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.GetMessage())
	}

	log.Info("authenticated")
	s.storage.SetClient(req.GetConn(), NewClient(req.GetConn(), s.adapter, log))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *SocketClosedRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	cli.Log.Info("socket closed", zap.String("reason", req.GetReason()))
	s.storage.SetClient(cli.Conn, nil)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *ReceivedBytesRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	if err := cli.OnReceivedBytes(ctx, req.GetBytes()); err != nil {
		cli.Log.Error("received bytes", zap.Error(err))

		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *TimerTimeoutRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	if err := cli.OnTimerTimeout(ctx, req.GetType()); err != nil {
		cli.Log.Error("timer timeout", zap.Error(err))

		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *ReceivedMessagesRequest) (*EmptySuccess, error) {
	cli, ok := s.storage.GetClient(req.GetConn())

	if !ok {
		return nil, nil
	}

	for _, msg := range req.GetMessages() {
		if err := cli.OnReceivedMessage(ctx, msg); err != nil {
			cli.Log.Error("received message", zap.Error(err))

			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
