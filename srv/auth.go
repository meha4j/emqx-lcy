package srv

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/paraskun/extd/api/auth"
	"github.com/paraskun/extd/pkg/emqx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/jackc/pgx/v5"
)

type Action = auth.ClientAuthorizeRequest_AuthorizeReqType

type Rule interface {
	Check(top, con string, act Action) bool
}

type ExclusiveRule struct {
	own atomic.Pointer[string]
}

func (r *ExclusiveRule) Check(top, con string, act Action) bool {
	if r.own.CompareAndSwap(nil, &con) {
		return true
	}

	if *r.own.Load() == con {
		return true
	}

	return false
}

type ACL struct {
	dat map[string][]Rule
	mux sync.RWMutex
}

func (a *ACL) Check(top, con string, act Action) bool {
	a.mux.RLock()
	defer a.mux.RUnlock()

	l, ok := a.dat[top]

	if !ok {
		return true
	}

	for _, r := range l {
		if ok := r.Check(top, con, act); !ok {
			return false
		}
	}

	return true
}

func (acc *ACL) Fetch(ctx context.Context, addr string) error {
	acc.mux.Lock()
	defer acc.mux.Unlock()

	con, err := pgx.Connect(ctx, addr)

	if err != nil {
		return fmt.Errorf("connect: %v", err)
	}

	res, err := con.Query(ctx, "SELECT COUNT(DISTINCT top) FROM rule")

	if err != nil {
		return fmt.Errorf("query: %v", err)
	}

	num, err := pgx.CollectExactlyOneRow(res, pgx.RowTo[int])

	if err != nil {
		return fmt.Errorf("scan: %v", err)
	}

	acc.dat = make(map[string][]Rule, num)
	res, err = con.Query(ctx, "SELECT top, mod FROM rule")

	if err != nil {
		return fmt.Errorf("query: %v", err)
	}

	var top, mod string

	for res.Next() {
		if err := res.Scan(&top, &mod); err != nil {
			return fmt.Errorf("scan: %v", err)
		}

		a, ok := acc.dat[top]

		if !ok {
			a = make([]Rule, 0, 1)
		}

		switch mod {
		case "ex":
			a = append(a, &ExclusiveRule{})
		}

		acc.dat[top] = a
	}

	return nil
}

func (acc *ACL) Clear() {
	acc.mux.Lock()
	defer acc.mux.Unlock()

	for k := range acc.dat {
		delete(acc.dat, k)
	}
}

var acl ACL

func NewAuth(srv *grpc.Server, cli *emqx.Client, cfg *viper.Viper, log *zap.SugaredLogger) error {
	if err := updateExHookServer(cli, cfg); err != nil {
		return fmt.Errorf("remote: %v", err)
	}

	host := cfg.GetString("extd.pgsql.host")
	port := cfg.GetInt("extd.pgsql.port")
	user := cfg.GetString("extd.pgsql.user")
	pass := cfg.GetString("extd.pgsql.pass")
	name := cfg.GetString("extd.auth.pgsql.name")
	addr := fmt.Sprintf("postgres://%s:%v/%s?user=%s&password=%s", host, port, name, user, pass)

	log.Infof("register", zap.String("pgsql", addr))

	auth.RegisterHookProviderServer(srv, &Auth{Log: log, Addr: addr})

	return nil
}

func updateExHookServer(cli *emqx.Client, cfg *viper.Viper) error {
	r := cfg.GetUint("extd.emqx.retry.num")
	t, err := time.ParseDuration(cfg.GetString("extd.emqx.retry.timeout"))

	if err != nil {
		return fmt.Errorf("timeout: %v", err)
	}

	port := cfg.GetInt("extd.port")

	for i := uint(0); true; i++ {
		if err = cli.UpdateExHookServer(&emqx.ExHookServerUpdateRequest{
			Name: "extd",
			Addr: fmt.Sprintf("http://%s:%d", cli.Addr, port),
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

type Auth struct {
	Log  *zap.SugaredLogger
	Addr string

	auth.UnimplementedHookProviderServer
}

func (s *Auth) OnProviderLoaded(ctx context.Context, _ *auth.ProviderLoadedRequest) (*auth.LoadedResponse, error) {
	if err := acl.Fetch(ctx, s.Addr); err != nil {
		return nil, fmt.Errorf("fetch: %v", err)
	}

	return &auth.LoadedResponse{
		Hooks: []*auth.HookSpec{
			{Name: "client.authorize"},
		},
	}, nil
}

func (s *Auth) OnProviderUnloaded(context.Context, *auth.ProviderUnloadedRequest) (*auth.EmptySuccess, error) {
	acl.Clear()

	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnClientConnect(context.Context, *auth.ClientConnectRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnClientConnack(context.Context, *auth.ClientConnackRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnClientConnected(context.Context, *auth.ClientConnectedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *Auth) OnClientDisconnected(_ context.Context, req *auth.ClientDisconnectedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *Auth) OnClientAuthenticate(_ context.Context, req *auth.ClientAuthenticateRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type:  auth.ValuedResponse_CONTINUE,
		Value: &auth.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *Auth) OnClientAuthorize(_ context.Context, req *auth.ClientAuthorizeRequest) (*auth.ValuedResponse, error) {
	if !acl.Check(req.Topic, req.Clientinfo.Clientid, req.Type) {
		return &auth.ValuedResponse{
			Type:  auth.ValuedResponse_STOP_AND_RETURN,
			Value: &auth.ValuedResponse_BoolResult{BoolResult: false},
		}, nil
	}

	return &auth.ValuedResponse{
		Type: auth.ValuedResponse_CONTINUE,
	}, nil
}

func (*Auth) OnClientSubscribe(context.Context, *auth.ClientSubscribeRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnClientUnsubscribe(context.Context, *auth.ClientUnsubscribeRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionCreated(context.Context, *auth.SessionCreatedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionSubscribed(context.Context, *auth.SessionSubscribedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionUnsubscribed(context.Context, *auth.SessionUnsubscribedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionResumed(context.Context, *auth.SessionResumedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionDiscarded(context.Context, *auth.SessionDiscardedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionTakenover(context.Context, *auth.SessionTakenoverRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnSessionTerminated(context.Context, *auth.SessionTerminatedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *Auth) OnMessagePublish(_ context.Context, req *auth.MessagePublishRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type:  auth.ValuedResponse_CONTINUE,
		Value: &auth.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (*Auth) OnMessageDelivered(context.Context, *auth.MessageDeliveredRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (*Auth) OnMessageDropped(context.Context, *auth.MessageDroppedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}

func (s *Auth) OnMessageAcked(_ context.Context, req *auth.MessageAckedRequest) (*auth.EmptySuccess, error) {
	return &auth.EmptySuccess{}, nil
}
