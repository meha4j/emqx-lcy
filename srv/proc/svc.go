package proc

import (
	"context"
	"fmt"
	"time"

	"github.com/paraskun/extd/api/proc"
	"github.com/paraskun/extd/internal/emqx"
	"github.com/paraskun/extd/pkg/vcas"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Register(srv *grpc.Server, cfg *viper.Viper, log *zap.Logger) error {
	svc, err := newService(cfg, log.Sugar())

	if err != nil {
		return fmt.Errorf("service: %v", err)
	}

	proc.RegisterConnectionUnaryHandlerServer(srv, svc)

	return nil
}

func newService(cfg *viper.Viper, log *zap.SugaredLogger) (proc.ConnectionUnaryHandlerServer, error) {
	if err := updateRemote(cfg); err != nil {
		return nil, fmt.Errorf("remote: %v", err)
	}

	adapter, err := newAdapter(cfg)

	if err != nil {
		return nil, fmt.Errorf("adapter: %v", err)
	}

	return &service{
		Log:     log,
		store:   NewStore(),
		adapter: adapter,
	}, nil
}

func updateRemote(cfg *viper.Viper) error {
	cli, err := newClient(cfg)

	if err != nil {
		return fmt.Errorf("emqx client: %v", err)
	}

	r := cfg.GetInt("extd.emqx.retry")
	t, err := time.ParseDuration(cfg.GetString("extd.emqx.timeout"))

	if err != nil {
		return fmt.Errorf("timeout: %v", err)
	}

	ap := cfg.GetInt("extd.proc.emqx.adapter.port")
	lp := cfg.GetInt("extd.proc.emqx.listener.port")
	hp := cfg.GetInt("extd.port")

	for i := 0; true; i++ {
		if err = cli.UpdateExProtoGateway(ap, lp, hp); err != nil {
			if i == r {
				return fmt.Errorf("update: %v", err)
			}

			time.Sleep(t)

			continue
		}

		break
	}

	return nil
}

func newClient(cfg *viper.Viper) (*emqx.Client, error) {
	port := cfg.GetInt("extd.emqx.port")
	host := cfg.GetString("extd.emqx.host")
	user := cfg.GetString("extd.emqx.user")
	pass := cfg.GetString("extd.emqx.pass")
	base := fmt.Sprintf("http://%s:%d/api/v5", host, port)

	return emqx.NewClient(base, user, pass)
}

func newAdapter(cfg *viper.Viper) (proc.ConnectionAdapterClient, error) {
	port := cfg.GetInt("extd.proc.emqx.adapter.port")
	host := cfg.GetString("extd.emqx.host")
	addr := fmt.Sprintf("%s:%d", host, port)

	crd := grpc.WithTransportCredentials(insecure.NewCredentials())
	con, err := grpc.NewClient(addr, crd)

	if err != nil {
		return nil, fmt.Errorf("grpc: %v", err)
	}

	return proc.NewConnectionAdapterClient(con), nil
}

type service struct {
	Log *zap.SugaredLogger

	store   *Store
	adapter proc.ConnectionAdapterClient

	proc.UnimplementedConnectionUnaryHandlerServer
}

func (s *service) OnSocketCreated(ctx context.Context, req *proc.SocketCreatedRequest) (*proc.EmptySuccess, error) {
	log := s.Log.With(
		zap.String("conn", req.GetConn()),
		zap.String("host", req.GetConninfo().GetPeername().GetHost()),
		zap.Uint32("port", req.GetConninfo().GetPeername().GetPort()),
	)

	res, err := s.adapter.Authenticate(ctx, &proc.AuthenticateRequest{
		Conn: req.GetConn(),
		Clientinfo: &proc.ClientInfo{
			ProtoName: vcas.Name,
			ProtoVer:  vcas.Version,
			Clientid:  req.GetConn(),
		},
	})

	if err != nil {
		log.Errorf("auth: adapter: %v", err)
		s.adapter.Close(ctx, &proc.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.GetCode() != proc.ResultCode_SUCCESS {
		log.Errorf("auth: %v", res.GetMessage())
		s.adapter.Close(ctx, &proc.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.GetMessage())
	}

	s.store.PutClient(req.GetConn(), NewClient(req.GetConn(), s.adapter, log))

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *proc.SocketClosedRequest) (*proc.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	s.store.RemoveClient(c.Conn)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *proc.ReceivedBytesRequest) (*proc.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	c.Log.Debugf("bytes received: %v", string(req.GetBytes()))

	if err := c.OnReceivedBytes(ctx, req.GetBytes()); err != nil {
		c.Log.Errorf("handle bytes: %v", err)

		return nil, status.Error(codes.Unknown, err.Error())
	}

	return nil, nil
}

func (s *service) OnTimerTimeout(ctx context.Context, req *proc.TimerTimeoutRequest) (*proc.EmptySuccess, error) {
	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *proc.ReceivedMessagesRequest) (*proc.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	for _, msg := range req.GetMessages() {
		c.Log.Debugf("message received: %v", msg)

		if err := c.OnReceivedMessage(ctx, msg); err != nil {
			c.Log.Errorf("handle message: %v", err)

			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return nil, nil
}
