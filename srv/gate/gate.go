package proc

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	procapi "github.com/paraskun/extd/api/proc"

	"github.com/paraskun/extd/pkg/emqx"
	"github.com/paraskun/extd/pkg/vcas"
	"github.com/paraskun/extd/srv/auth"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Register(srv *grpc.Server, ctl *auth.ACL, cli *emqx.Client, cfg *viper.Viper) error {
	if err := updateRemote(cli, cfg); err != nil {
		return fmt.Errorf("remote: %v", err)
	}

	adapter, err := newAdapter(cfg)

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	procapi.RegisterConnectionUnaryHandlerServer(srv, &service{ctl: ctl, adapter: adapter})

	return nil
}

func updateRemote(cli *emqx.Client, cfg *viper.Viper) error {
	err := cli.UpdateExProtoGateway(&emqx.ExProtoGatewayUpdateRequest{
		Name:    cfg.GetString("extd.proc.name"),
		Enable:  cfg.GetBool("extd.proc.enable"),
		Timeout: cfg.GetString("extd.proc.tout"),
		Server: emqx.Server{
			Bind: strconv.Itoa(cfg.GetInt("extd.proc.server.port")),
		},
		Handler: emqx.Handler{
			Addr: fmt.Sprintf("http://%s:%d", cli.Addr, cfg.GetInt("extd.port")),
		},
		Listeners: []emqx.Listener{
			{
				Name: cfg.GetString("extd.proc.listener.name"),
				Type: cfg.GetString("extd.proc.listener.type"),
				Bind: strconv.Itoa(cfg.GetInt("extd.proc.listener.port")),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return nil
}

func newAdapter(cfg *viper.Viper) (procapi.ConnectionAdapterClient, error) {
	addr := fmt.Sprintf("%s:%d", cfg.GetString("extd.emqx.host"), cfg.GetInt("extd.proc.server.port"))
	crd := grpc.WithTransportCredentials(insecure.NewCredentials())
	con, err := grpc.NewClient(addr, crd)

	if err != nil {
		return nil, fmt.Errorf("grpc: %v", err)
	}

	return procapi.NewConnectionAdapterClient(con), nil
}

type service struct {
	ctl     *auth.ACL
	store   sync.Map
	adapter procapi.ConnectionAdapterClient

	procapi.UnimplementedConnectionUnaryHandlerServer
}

func (s *service) OnSocketCreated(ctx context.Context, req *procapi.SocketCreatedRequest) (*procapi.EmptySuccess, error) {
	res, err := s.adapter.Authenticate(ctx, &procapi.AuthenticateRequest{
		Conn: req.Conn,
		Clientinfo: &procapi.ClientInfo{
			ProtoName: vcas.Name,
			ProtoVer:  vcas.Version,
			Clientid:  req.Conn,
		},
	})

	if err != nil {
		slog.Error("auth", "conn", req.Conninfo.String(), "err", err)
		s.adapter.Close(ctx, &procapi.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.Code != procapi.ResultCode_SUCCESS {
		slog.Error("auth", "conn", req.Conninfo.String(), "code", res.Code)
		s.adapter.Close(ctx, &procapi.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.Message)
	}

	s.store.Store(req.Conn, NewClient(req.Conn, s.ctl, s.adapter))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *procapi.SocketClosedRequest) (*procapi.EmptySuccess, error) {
	s.store.Delete(req.Conn)
	s.ctl.Release(req.Conn)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *procapi.ReceivedBytesRequest) (*procapi.EmptySuccess, error) {
	v, ok := s.store.Load(req.Conn)

	if !ok {
		return nil, nil
	}

	if err := v.(*Client).OnReceivedBytes(ctx, req.Bytes); err != nil {
		slog.Error("bytes", "content", string(req.Bytes), "err", err)
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *procapi.TimerTimeoutRequest) (*procapi.EmptySuccess, error) {
	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *procapi.ReceivedMessagesRequest) (*procapi.EmptySuccess, error) {
	c, ok := s.store.Load(req.Conn)

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
