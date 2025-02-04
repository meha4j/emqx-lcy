package gate

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/blabtm/emqx-gate/api"
	"github.com/blabtm/emqx-gate/vcas"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Config struct {
	Port int
	Emqx struct {
		Adapter struct {
			Host string
			Port int
		} `mapstructure:"adapter"`
	} `mapstructure:"emqx"`
}

func Register(srv *grpc.Server, cfg *Config) error {
	con, err := grpc.NewClient(fmt.Sprintf("%s:%d",
		cfg.Emqx.Adapter.Host,
		cfg.Emqx.Adapter.Port,
	), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return fmt.Errorf("grpc: %w", err)
	}

	cli := gate.NewConnectionAdapterClient(con)
	svc := &service{cli: cli}

	gate.RegisterConnectionUnaryHandlerServer(srv, svc)

	return nil
}

type service struct {
	dat sync.Map
	cli gate.ConnectionAdapterClient

	gate.UnimplementedConnectionUnaryHandlerServer
}

func (s *service) OnSocketCreated(ctx context.Context, req *gate.SocketCreatedRequest) (*gate.EmptySuccess, error) {
	res, err := s.cli.Authenticate(ctx, &gate.AuthenticateRequest{
		Conn: req.Conn,
		Clientinfo: &gate.ClientInfo{
			ProtoName: vcas.Name,
			ProtoVer:  vcas.Version,
			Clientid:  req.Conn,
			Username:  req.Conn,
		},
	})

	if err != nil {
		slog.Error("authn", "con", req.Conninfo.String(), "err", err)
		s.cli.Close(ctx, &gate.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.Code != gate.ResultCode_SUCCESS {
		slog.Error("authn", "con", req.Conninfo.String(), "code", res.Code)
		s.cli.Close(ctx, &gate.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.Message)
	}

	s.dat.Store(req.Conn, newClient(req.Conn, s.cli))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *gate.SocketClosedRequest) (*gate.EmptySuccess, error) {
	s.dat.Delete(req.Conn)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *gate.ReceivedBytesRequest) (*gate.EmptySuccess, error) {
	v, ok := s.dat.Load(req.Conn)

	if !ok {
		return nil, nil
	}

	if err := v.(*client).OnReceivedBytes(ctx, req.Bytes); err != nil {
		slog.Error("bytes", "con", req.Conn, "pay", string(req.Bytes), "err", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *gate.TimerTimeoutRequest) (*gate.EmptySuccess, error) {
	return &gate.EmptySuccess{}, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *gate.ReceivedMessagesRequest) (*gate.EmptySuccess, error) {
	c, ok := s.dat.Load(req.Conn)

	if !ok {
		return nil, nil
	}

	for _, msg := range req.Messages {
		if err := c.(*client).OnReceivedMessage(ctx, msg); err != nil {
			slog.Error("msg", "con", req.Conn, "pay", msg, "err", err)
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
