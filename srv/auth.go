package srv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/paraskun/extd/api/auth"
	"github.com/paraskun/extd/pkg/emqx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jackc/pgx/v5"
)

func NewAuth(srv *grpc.Server, cfg *viper.Viper, log *zap.SugaredLogger) error {
	if err := updateRemote(cfg); err != nil {
		return fmt.Errorf("remote: %v", err)
	}

	host := cfg.GetString("extd.pgsql.host")
	port := cfg.GetInt("extd.pgsql.port")
	user := cfg.GetString("extd.pgsql.user")
	pass := cfg.GetString("extd.pgsql.pass")
	name := cfg.GetString("extd.auth.pgsql.name")
	addr := fmt.Sprintf("postgres://%s:%v/%s?user=%s&password=%s", host, port, name, user, pass)

	log.Infof("register", zap.String("pgsql", addr))

	auth.RegisterHookProviderServer(srv, &service{Log: log, addr: addr})

	return nil
}

func updateRemote(cfg *viper.Viper) error {
	cli, err := newClient(cfg)

	if err != nil {
		return fmt.Errorf("emqx client: %v", err)
	}

	r := cfg.GetUint("extd.emqx.retry.num")
	t, err := time.ParseDuration(cfg.GetString("extd.emqx.retry.timeout"))

	if err != nil {
		return fmt.Errorf("timeout: %v", err)
	}

	port := cfg.GetInt("extd.port")

	for i := uint(0); true; i++ {
		if err = cli.UpdateExHookServer("extd", port); err != nil {
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

type service struct {
	Log *zap.SugaredLogger

	addr  string
	store sync.Map

	auth.UnimplementedHookProviderServer
}

func (s *service) OnProviderLoaded(ctx context.Context, _ *auth.ProviderLoadedRequest) (*auth.LoadedResponse, error) {
	con, err := pgx.Connect(ctx, s.addr)

	if err != nil {
		return nil, fmt.Errorf("pgx: con: %v", err)
	}

	defer con.Close(ctx)

	top, err := QueryExclusive(ctx, con)

	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	s.Log.Infof("exclusive: %v", top)

	return &auth.LoadedResponse{
		Hooks: []*auth.HookSpec{
			{Name: "client.authorize", Topics: top},
		},
	}, nil
}

func QueryExclusive(ctx context.Context, con *pgx.Conn) ([]string, error) {
	set, err := con.Query(ctx, "SELECT top FROM rule WHERE mod = 'ex'")

	if err != nil {
		return nil, fmt.Errorf("exec: %v", err)
	}

	res, err := pgx.CollectRows(set, pgx.RowTo[string])

	if err != nil {
		return nil, fmt.Errorf("collect: %v", err)
	}

	return res, nil
}

func (s *service) OnProviderUnloaded(context.Context, *auth.ProviderUnloadedRequest) (*auth.EmptySuccess, error) {
	s.store.Clear()

	return &auth.EmptySuccess{}, nil
}

func (*service) OnClientConnect(context.Context, *auth.ClientConnectRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnClientConnack(context.Context, *auth.ClientConnackRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnClientConnected(context.Context, *auth.ClientConnectedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *service) OnClientDisconnected(_ context.Context, req *auth.ClientDisconnectedRequest) (*auth.EmptySuccess, error) {
	s.store.Range(func(key, val any) bool {
		s.store.CompareAndDelete(key, req.Clientinfo.Clientid)
		return true
	})

	return &auth.EmptySuccess{}, nil
}

func (s *service) OnClientAuthenticate(_ context.Context, req *auth.ClientAuthenticateRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type:  auth.ValuedResponse_CONTINUE,
		Value: &auth.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(_ context.Context, req *auth.ClientAuthorizeRequest) (*auth.ValuedResponse, error) {
	s.Log.Infof("priv: %v", req)
	res := auth.ValuedResponse{
		Type:  auth.ValuedResponse_STOP_AND_RETURN,
		Value: &auth.ValuedResponse_BoolResult{BoolResult: true},
	}

	if req.Type != auth.ClientAuthorizeRequest_PUBLISH {
		s.Log.Infof("ok: %v", req)
		return &res, nil
	}

	v, ok := s.store.LoadOrStore(req.Topic, req.Clientinfo.Clientid)

	if ok && v != req.Clientinfo.Clientid {
		res.Value = &auth.ValuedResponse_BoolResult{BoolResult: false}
		return &res, nil
	}

	return &res, nil
}

func (*service) OnClientSubscribe(context.Context, *auth.ClientSubscribeRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnClientUnsubscribe(context.Context, *auth.ClientUnsubscribeRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionCreated(context.Context, *auth.SessionCreatedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionSubscribed(context.Context, *auth.SessionSubscribedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionUnsubscribed(context.Context, *auth.SessionUnsubscribedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionResumed(context.Context, *auth.SessionResumedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionDiscarded(context.Context, *auth.SessionDiscardedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionTakenover(context.Context, *auth.SessionTakenoverRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnSessionTerminated(context.Context, *auth.SessionTerminatedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *service) OnMessagePublish(_ context.Context, req *auth.MessagePublishRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type:  auth.ValuedResponse_CONTINUE,
		Value: &auth.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (*service) OnMessageDelivered(context.Context, *auth.MessageDeliveredRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*service) OnMessageDropped(context.Context, *auth.MessageDroppedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *service) OnMessageAcked(_ context.Context, req *auth.MessageAckedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}
