package gate

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	api "github.com/blabtm/extd/internal/api/gate"

	"github.com/blabtm/extd/vcas"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Register(srv *grpc.Server, cfg *viper.Viper) error {
	crd := grpc.WithTransportCredentials(insecure.NewCredentials())
	con, err := grpc.NewClient(fmt.Sprintf("%s:%d",
		cfg.GetString("extd.emqx.host"),
		cfg.GetInt("extd.gate.server.port"),
	), crd)

	if err != nil {
		return fmt.Errorf("cli: %v", err)
	}

	cli := api.NewConnectionAdapterClient(con)
	svc := &service{cli: cli}

	api.RegisterConnectionUnaryHandlerServer(srv, svc)

	return nil
}

type service struct {
	dat sync.Map
	cli api.ConnectionAdapterClient

	api.UnimplementedConnectionUnaryHandlerServer
}

func (s *service) OnSocketCreated(ctx context.Context, req *api.SocketCreatedRequest) (*api.EmptySuccess, error) {
	slog.Debug("gate: created", "con", req.Conninfo.String())

	res, err := s.cli.Authenticate(ctx, &api.AuthenticateRequest{
		Conn: req.Conn,
		Clientinfo: &api.ClientInfo{
			ProtoName: vcas.Name,
			ProtoVer:  vcas.Version,
			Clientid:  req.Conn,
			Username:  req.Conn,
		},
	})

	if err != nil {
		slog.Error("gate: authn", "con", req.Conninfo.String(), "err", err)
		s.cli.Close(ctx, &api.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.Code != api.ResultCode_SUCCESS {
		slog.Error("gate: authn", "con", req.Conninfo.String(), "code", res.Code)
		s.cli.Close(ctx, &api.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.Message)
	}

	s.dat.Store(req.Conn, NewClient(req.Conn, s.cli))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *api.SocketClosedRequest) (*api.EmptySuccess, error) {
	slog.Debug("gate: closed", "con", req.Conn)
	s.dat.Delete(req.Conn)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *api.ReceivedBytesRequest) (*api.EmptySuccess, error) {
	slog.Debug("gate: bytes", "con", req.Conn, "pay", string(req.Bytes))

	v, ok := s.dat.Load(req.Conn)

	if !ok {
		return nil, nil
	}

	if err := v.(*Client).OnReceivedBytes(ctx, req.Bytes); err != nil {
		slog.Error("gate: bytes", "con", req.Conn, "pay", string(req.Bytes), "err", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *api.TimerTimeoutRequest) (*api.EmptySuccess, error) {
	return &api.EmptySuccess{}, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *api.ReceivedMessagesRequest) (*api.EmptySuccess, error) {
	c, ok := s.dat.Load(req.Conn)

	if !ok {
		return nil, nil
	}

	for _, msg := range req.Messages {
		slog.Debug("gate: msg", "con", req.Conn, "pay", msg)

		if err := c.(*Client).OnReceivedMessage(ctx, msg); err != nil {
			slog.Error("gate: msg", "con", req.Conn, "pay", msg, "err", err)
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
