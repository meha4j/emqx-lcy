package srv

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/paraskun/extd/api/proc"
	"github.com/paraskun/extd/pkg/emqx"
	"github.com/paraskun/extd/pkg/vcas"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Cache struct {
}

func NewProc(srv *grpc.Server, cli *emqx.Client, cfg *viper.Viper, log *zap.SugaredLogger) error {
	if err := updateExProtoGateway(cli, cfg); err != nil {
		return fmt.Errorf("remote: %v", err)
	}

	adapter, err := newAdapter(cfg)

	if err != nil {
		return fmt.Errorf("adapter: %v", err)
	}

	proc.RegisterConnectionUnaryHandlerServer(srv, &Proc{
		Log:     log,
		adapter: adapter,
	})

	return nil
}

func updateExProtoGateway(cli *emqx.Client, cfg *viper.Viper) error {
	r := cfg.GetInt("extd.emqx.retry")
	t, err := time.ParseDuration(cfg.GetString("extd.emqx.timeout"))

	if err != nil {
		return fmt.Errorf("timeout: %v", err)
	}

	ap := cfg.GetInt("extd.proc.emqx.adapter.port")
	lp := cfg.GetInt("extd.proc.emqx.listener.port")
	hp := cfg.GetInt("extd.port")

	for i := 0; true; i++ {
		if err = cli.UpdateExProtoGateway(&emqx.ExProtoGatewayUpdateRequest{
			Name: "exproto",
			Server: emqx.Server{
				Bind: fmt.Sprintf(":%d", ap),
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
		}); err != nil {
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

type Proc struct {
	Log *zap.SugaredLogger

	store   *Store
	adapter proc.ConnectionAdapterClient

	proc.UnimplementedConnectionUnaryHandlerServer
}

func (s *Proc) OnSocketCreated(ctx context.Context, req *proc.SocketCreatedRequest) (*proc.EmptySuccess, error) {
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

func (s *Proc) OnSocketClosed(_ context.Context, req *proc.SocketClosedRequest) (*proc.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	s.store.RemoveClient(c.Conn)

	return nil, nil
}

func (s *Proc) OnReceivedBytes(ctx context.Context, req *proc.ReceivedBytesRequest) (*proc.EmptySuccess, error) {
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

func (s *Proc) OnTimerTimeout(ctx context.Context, req *proc.TimerTimeoutRequest) (*proc.EmptySuccess, error) {
	return nil, nil
}

func (s *Proc) OnReceivedMessages(ctx context.Context, req *proc.ReceivedMessagesRequest) (*proc.EmptySuccess, error) {
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
