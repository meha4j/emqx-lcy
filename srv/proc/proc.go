package proc

import (
	"context"
	"fmt"
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
	ap := cfg.GetInt("extd.proc.emqx.adapter.port")
	lp := cfg.GetInt("extd.proc.emqx.listener.port")
	hp := cfg.GetInt("extd.port")

	err := cli.UpdateExProtoGateway(&emqx.ExProtoGatewayUpdateRequest{
		Name:    "exproto",
		Enable:  false,
		Timeout: "30s",
		Server: emqx.Server{
			Bind: strconv.Itoa(ap),
		},
		Handler: emqx.Handler{
			Addr: fmt.Sprintf("http://%s:%d", cli.Addr, hp),
		},
		Listeners: []emqx.Listener{
			{
				Name: "default",
				Type: "tcp",
				Bind: strconv.Itoa(lp),
			},
		},
	})

	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return nil
}

func newAdapter(cfg *viper.Viper) (procapi.ConnectionAdapterClient, error) {
	port := cfg.GetInt("extd.proc.emqx.adapter.port")
	host := cfg.GetString("extd.emqx.host")
	addr := fmt.Sprintf("%s:%d", host, port)

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
		Conn: req.GetConn(),
		Clientinfo: &procapi.ClientInfo{
			ProtoName: vcas.Name,
			ProtoVer:  vcas.Version,
			Clientid:  req.GetConn(),
		},
	})

	if err != nil {
		s.adapter.Close(ctx, &procapi.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.GetCode() != procapi.ResultCode_SUCCESS {
		s.adapter.Close(ctx, &procapi.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.GetMessage())
	}

	s.store.Store(req.GetConn(), NewClient(req.GetConn(), s.ctl, s.adapter))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *procapi.SocketClosedRequest) (*procapi.EmptySuccess, error) {
	s.store.Delete(req.Conn)
	s.ctl.Release(req.Conn)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *procapi.ReceivedBytesRequest) (*procapi.EmptySuccess, error) {
	v, ok := s.store.Load(req.GetConn())

	if !ok {
		return nil, nil
	}

	if err := v.(*Client).OnReceivedBytes(ctx, req.GetBytes()); err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *procapi.TimerTimeoutRequest) (*procapi.EmptySuccess, error) {
	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *procapi.ReceivedMessagesRequest) (*procapi.EmptySuccess, error) {
	c, ok := s.store.Load(req.GetConn())

	if !ok {
		return nil, nil
	}

	for _, msg := range req.GetMessages() {
		if err := c.(*Client).OnReceivedMessage(ctx, msg); err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
