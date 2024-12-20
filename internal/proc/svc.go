package proc

import (
	context "context"
	"fmt"
	"time"

	"github.com/paraskun/extd/internal/pkg/emqx"
	"github.com/paraskun/extd/internal/proc/proto"
	"github.com/paraskun/extd/pkg/vcas"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"
)

func init() {
	viper.SetDefault("extd.proc.emqx.adapter.port", 9110)
	viper.SetDefault("extd.proc.emqx.listener.port", 20041)
}

func NewAdapter(addr string) (proto.ConnectionAdapterClient, error) {
	crd := grpc.WithTransportCredentials(insecure.NewCredentials())
	con, err := grpc.NewClient(addr, crd)

	if err != nil {
		return nil, fmt.Errorf("grpc client: %v", err)
	}

	return proto.NewConnectionAdapterClient(con), nil
}

type service struct {
	Log *zap.SugaredLogger

	store   *Store
	adapter proto.ConnectionAdapterClient

	proto.UnimplementedConnectionUnaryHandlerServer
}

func NewService(log *zap.SugaredLogger) (proto.ConnectionUnaryHandlerServer, error) {
	port := viper.GetInt("extd.port")
	user := viper.GetString("extd.secret.user")
	pass := viper.GetString("extd.secret.pass")

	ehost := viper.GetString("extd.emqx.host")
	eport := viper.GetInt("extd.emqx.port")
	aport := viper.GetInt("extd.proc.emqx.adapter.port")
	lport := viper.GetInt("extd.proc.emqx.listener.port")

	base := fmt.Sprintf("http://%s:%d/api/v5", ehost, eport)
	client, err := emqx.NewClient(base, user, pass)

	if err != nil {
		return nil, fmt.Errorf("emqx client: %v", err)
	}

	timeout, err := time.ParseDuration(viper.GetString("extd.emqx.timeout"))

	if err != nil {
		return nil, fmt.Errorf("parse timeout: %v", err)
	}

	for i := 0; true; i++ {
		if err = client.UpdateExProtoGateway(aport, lport, port); err != nil {
			if i == 5 {
				return nil, fmt.Errorf("update configuration: %v", err)
			}

			log.Errorf("update configuration (%d): %v, retry in %d", i+1, err, timeout.String())
			time.Sleep(timeout)
			continue
		}

		break
	}

	addr := fmt.Sprintf("%s:%d", ehost, aport)
	adapter, err := NewAdapter(addr)

	if err != nil {
		return nil, fmt.Errorf("adapter: %v", err)
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

	log.Info("authenticating")

	res, err := s.adapter.Authenticate(ctx, &proto.AuthenticateRequest{
		Conn: req.GetConn(),
		Clientinfo: &proto.ClientInfo{
			ProtoName: vcas.Name,
			ProtoVer:  vcas.Version,
			Clientid:  req.GetConn(),
		},
	})

	if err != nil {
		log.Errorf("authentication: adapter: %v", err)
		s.adapter.Close(ctx, &proto.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		log.Errorf("authentication: %v", res.GetMessage())
		s.adapter.Close(ctx, &proto.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unauthenticated, res.GetMessage())
	}

	s.store.PutClient(req.GetConn(), NewClient(req.GetConn(), s.adapter, log))

	log.Info("authenticated")
	log.Debug("starting keepalive timer")

	res, err = s.adapter.StartTimer(ctx, &proto.TimerRequest{
		Conn:     req.GetConn(),
		Type:     proto.TimerType_KEEPALIVE,
		Interval: 300,
	})

	if err != nil {
		log.Errorf("timer: adapter:", err)
		s.adapter.Close(ctx, &proto.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.GetCode() != proto.ResultCode_SUCCESS {
		log.Errorf("timer: %v", res.GetMessage())
		s.adapter.Close(ctx, &proto.CloseSocketRequest{Conn: req.Conn})

		return nil, status.Error(codes.Unknown, res.GetMessage())
	}

	log.Debug("keepalive timer started")

	return nil, nil
}

func (s *service) OnSocketClosed(_ context.Context, req *proto.SocketClosedRequest) (*proto.EmptySuccess, error) {
	c, ok := s.store.GetClientByConn(req.GetConn())

	if !ok {
		return nil, nil
	}

	c.Log.Infof("closed: %v", req.GetReason())
	s.store.RemoveClient(c.Conn)

	return nil, nil
}

func (s *service) OnReceivedBytes(ctx context.Context, req *proto.ReceivedBytesRequest) (*proto.EmptySuccess, error) {
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

func (s *service) OnTimerTimeout(ctx context.Context, req *proto.TimerTimeoutRequest) (*proto.EmptySuccess, error) {
	return nil, nil
}

func (s *service) OnReceivedMessages(ctx context.Context, req *proto.ReceivedMessagesRequest) (*proto.EmptySuccess, error) {
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
