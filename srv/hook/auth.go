package auth

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paraskun/extd/api/auth"
	"github.com/paraskun/extd/pkg/emqx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func Register(srv *grpc.Server, ctl *ACL, cli *emqx.Client, cfg *viper.Viper) error {
	if err := updateRemote(cli, cfg); err != nil {
		return fmt.Errorf("remote: %v", err)
	}

	auth.RegisterHookProviderServer(srv, &service{
		dbQuery: cfg.GetString("extd.auth.pgsql.query"),
		dbAddr: fmt.Sprintf("postgres://%s:%v/%s?user=%s&password=%s",
			cfg.GetString("extd.pgsql.host"),
			cfg.GetInt("extd.pgsql.port"),
			cfg.GetString("extd.auth.pgsql.name"),
			cfg.GetString("extd.pgsql.user"),
			cfg.GetString("extd.pgsql.pass"),
		),
		ctl: ctl,
	})

	return nil
}

func updateRemote(cli *emqx.Client, cfg *viper.Viper) error {
	err := cli.UpdateExHookServer(&emqx.ExHookServerUpdateRequest{
		Name:      cfg.GetString("extd.auth.name"),
		Enable:    cfg.GetBool("extd.auth.enable"),
		Action:    cfg.GetString("extd.auth.action"),
		Timeout:   cfg.GetString("extd.auth.tout"),
		Reconnect: cfg.GetString("extd.auth.trec"),
		Addr:      fmt.Sprintf("http://%s:%d", cli.Addr, cfg.GetInt("extd.port")),
	})

	if err != nil {
		return fmt.Errorf("update: %v", err)
	}

	return nil
}

type service struct {
	dbQuery string
	dbAddr  string
	ctl     *ACL

	auth.UnimplementedHookProviderServer
}

func (s *service) OnProviderLoaded(ctx context.Context, _ *auth.ProviderLoadedRequest) (*auth.LoadedResponse, error) {
	con, err := sql.Open("pgx", s.dbAddr)

	if err != nil {
		return nil, fmt.Errorf("postgres: %v", err)
	}

	defer con.Close()

	if err := s.ctl.Fetch(con, s.dbQuery); err != nil {
		return nil, fmt.Errorf("fetch: %v", err)
	}

	return &auth.LoadedResponse{
		Hooks: []*auth.HookSpec{
			{Name: "client.authorize"},
		},
	}, nil
}

func (s *service) OnProviderUnloaded(context.Context, *auth.ProviderUnloadedRequest) (*auth.EmptySuccess, error) {
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
	s.ctl.Release(req.Clientinfo.Clientid)
	return &auth.EmptySuccess{}, nil
}

func (s *service) OnClientserviceenticate(_ context.Context, req *auth.ClientAuthenticateRequest) (*auth.ValuedResponse, error) {
	return &auth.ValuedResponse{
		Type:  auth.ValuedResponse_CONTINUE,
		Value: &auth.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(_ context.Context, req *auth.ClientAuthorizeRequest) (*auth.ValuedResponse, error) {
	if !s.ctl.Check(req.Topic, req.Clientinfo.Clientid, req.Type) {
		return &auth.ValuedResponse{
			Type:  auth.ValuedResponse_STOP_AND_RETURN,
			Value: &auth.ValuedResponse_BoolResult{BoolResult: false},
		}, nil
	}

	return &auth.ValuedResponse{
		Type: auth.ValuedResponse_CONTINUE,
	}, nil
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
		Type: auth.ValuedResponse_CONTINUE,
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
