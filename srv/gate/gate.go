package gate

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/paraskun/extd/api/gate"

	"github.com/paraskun/extd/pkg/emqx"
	"github.com/paraskun/extd/pkg/vcas"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type options struct {
	cli *emqx.Client
}

type Option func(opts *options) error

func WithClient(cli *emqx.Client) Option {
	return func(opts *options) error {
		opts.cli = cli
		return nil
	}
}

func Register(srv *grpc.Server, cfg *viper.Viper, opts ...Option) error {
	var opt options

	for _, exe := range opts {
		if err := exe(&opt); err != nil {
			return fmt.Errorf("opt: %v", err)
		}
	}

	if opt.cli != nil {
		if err := opt.cli.UpdateExProtoGateway(&emqx.ExProtoGatewayUpdateRequest{
			Name:    cfg.GetString("extd.gate.name"),
			Enable:  cfg.GetBool("extd.gate.enable"),
			Timeout: cfg.GetString("extd.gate.tout"),
			Server: emqx.Server{
				Bind: strconv.Itoa(cfg.GetInt("extd.gate.server.port")),
			},
			Handler: emqx.Handler{
				Addr: fmt.Sprintf("http://%s:%d", opt.cli.Addr, cfg.GetInt("extd.port")),
			},
			Listeners: []emqx.Listener{
				{
					Name: cfg.GetString("extd.gate.listener.name"),
					Type: cfg.GetString("extd.gate.listener.type"),
					Bind: strconv.Itoa(cfg.GetInt("extd.gate.listener.port")),
				},
			},
		}); err != nil {
			return fmt.Errorf("upd: %v", err)
		}
	}

	crd := grpc.WithTransportCredentials(insecure.NewCredentials())
	con, err := grpc.NewClient(fmt.Sprintf("%s:%d",
		cfg.GetString("extd.emqx.host"),
		cfg.GetInt("extd.gate.server.port"),
	), crd)

	if err != nil {
		return fmt.Errorf("cli: %v", err)
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
		},
	})

	if err != nil {
		slog.Error("auth", "conn", req.Conninfo.String(), "err", err)
		s.cli.Close(ctx, &gate.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.Code != gate.ResultCode_SUCCESS {
		slog.Error("auth", "conn", req.Conninfo.String(), "code", res.Code)
		s.cli.Close(ctx, &gate.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.Message)
	}

	s.dat.Store(req.Conn, NewClient(req.Conn, s.cli))

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

	if err := v.(*Client).OnReceivedBytes(ctx, req.Bytes); err != nil {
		slog.Error("bytes", "content", string(req.Bytes), "err", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *gate.TimerTimeoutRequest) (*gate.EmptySuccess, error) {
	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *gate.ReceivedMessagesRequest) (*gate.EmptySuccess, error) {
	c, ok := s.dat.Load(req.Conn)

	if !ok {
		return nil, nil
	}

	for _, msg := range req.Messages {
		if err := c.(*Client).OnReceivedMessage(ctx, msg); err != nil {
			slog.Error("message", "content", msg, "err", err)
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
