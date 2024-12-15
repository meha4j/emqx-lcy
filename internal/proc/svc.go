package proc

import (
	context "context"
	"fmt"
	"os"

	"github.com/paraskun/extd/internal/proc/proto"
	"github.com/paraskun/extd/pkg/emqx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
)

const (
	ProtoName = "VCAS"
	ProtoVer  = "2.0.0-SNAPSHOT"

	EmqxAddr = "EXTD_EMQX_ADDR"
	Username = "EXTD_USERNAME"
	Password = "EXTD_PASSWORD"
)

type service struct {
	Log *zap.Logger

	store   *Store
	adapter proto.ConnectionAdapterClient

	proto.UnimplementedConnectionUnaryHandlerServer
}

func NewAdapter(addr string) (proto.ConnectionAdapterClient, error) {
	cred := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, cred)

	if err != nil {
		return nil, fmt.Errorf("grpc client: %v", err)
	}

	return proto.NewConnectionAdapterClient(conn), nil
}

func NewService(log *zap.Logger) (proto.ConnectionUnaryHandlerServer, error) {
	addr := os.Getenv(EmqxAddr)
	name := os.Getenv(Username)
	pass := os.Getenv(Password)

	if addr == "" {
		return nil, fmt.Errorf("emqx address does not provided.")
	}

	adapter, err := NewAdapter(addr)

	if err != nil {
		return nil, fmt.Errorf("adapter: %v", err)
	}

	client := emqx.NewClient(addr+"/api/v5", name, pass)
	err = client.UpdateProtoGateway()

	if err != nil {
		return nil, fmt.Errorf("update configuration: %v", err)
	}

	return &service{
		Log:     log,
		store:   NewStore(),
		adapter: adapter,
	}, nil
}

func (s *service) OnSocketCreated(ctx context.Context, req *proto.SocketCreatedRequest) (*proto.EmptySuccess, error) {
	log := s.Log.With(
		zap.String("conn", req.GetConn()),
		zap.String("host", req.GetConninfo().GetPeername().GetHost()),
		zap.Uint32("port", req.GetConninfo().GetPeername().GetPort()),
	)

	res, err := s.adapter.Authenticate(ctx, &proto.AuthenticateRequest{
		Conn: req.GetConn(),
		Clientinfo: &proto.ClientInfo{
			ProtoName: ProtoName,
			ProtoVer:  ProtoVer,
			Clientid:  req.GetConn(),
		},
	})

	if err != nil {
		s.Log.Error("adapter", zap.Error(err))
		s.adapter.Close(ctx, &proto.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		log.Error("authentication", zap.String("error", res.GetMessage()))
		s.adapter.Close(ctx, &proto.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.GetMessage())
	}

	log.Info("authenticated")
	s.store.PutClient(req.GetConn(), NewClient(req.GetConn(), s.adapter, log))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *proto.SocketClosedRequest) (*proto.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	c.Log.Info("closed", zap.String("reason", req.GetReason()))
	s.store.PutClient(c.Conn, nil)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *proto.ReceivedBytesRequest) (*proto.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	if err := c.OnReceivedBytes(ctx, req.GetBytes()); err != nil {
		c.Log.Error("handle bytes", zap.Error(err))

		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *proto.TimerTimeoutRequest) (*proto.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	if err := c.OnTimerTimeout(ctx, req.GetType()); err != nil {
		c.Log.Error("handle timer", zap.Error(err))

		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *proto.ReceivedMessagesRequest) (*proto.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	for _, msg := range req.GetMessages() {
		if err := c.OnReceivedMessage(ctx, msg); err != nil {
			c.Log.Error("handle message", zap.Error(err))

			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
