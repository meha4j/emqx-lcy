package hook

import (
	"context"
	"database/sql"
	"fmt"

	hookapi "github.com/paraskun/extd/api/hook"

	"github.com/paraskun/extd/pkg/emqx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Register(srv *grpc.Server, ctl *ACL, cli *emqx.Client, cfg *viper.Viper) error {
	if err := updateRemote(cli, cfg); err != nil {
		return fmt.Errorf("remote: %v", err)
	}

	hookapi.RegisterHookProviderServer(srv, &service{
		dbQuery: cfg.GetString("extd.hookapi.pgsql.query"),
		dbAddr: fmt.Sprintf("postgres://%s:%v/%s?user=%s&password=%s",
			cfg.GetString("extd.pgsql.host"),
			cfg.GetInt("extd.pgsql.port"),
			cfg.GetString("extd.hookapi.pgsql.name"),
			cfg.GetString("extd.pgsql.user"),
			cfg.GetString("extd.pgsql.pass"),
		),
		ctl: ctl,
	})

	return nil
}

func updateRemote(cli *emqx.Client, cfg *viper.Viper) error {
	err := cli.UpdateExHookServer(&emqx.ExHookServerUpdateRequest{
		Name:      cfg.GetString("extd.hookapi.name"),
		Enable:    cfg.GetBool("extd.hookapi.enable"),
		Action:    cfg.GetString("extd.hookapi.action"),
		Timeout:   cfg.GetString("extd.hookapi.tout"),
		Reconnect: cfg.GetString("extd.hookapi.trec"),
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

	hookapi.UnimplementedHookProviderServer
}

func (s *service) OnProviderLoaded(ctx context.Context, _ *hookapi.ProviderLoadedRequest) (*hookapi.LoadedResponse, error) {
	con, err := sql.Open("pgx", s.dbAddr)

	if err != nil {
		return nil, fmt.Errorf("postgres: %v", err)
	}

	defer con.Close()

	if err := s.ctl.Fetch(con, s.dbQuery); err != nil {
		return nil, fmt.Errorf("fetch: %v", err)
	}

	return &hookapi.LoadedResponse{
		Hooks: []*hookapi.HookSpec{
			{Name: "client.hookapiorize"},
		},
	}, nil
}

func (s *service) OnProviderUnloaded(context.Context, *hookapi.ProviderUnloadedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnClientConnect(context.Context, *hookapi.ClientConnectRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnClientConnack(context.Context, *hookapi.ClientConnackRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnClientConnected(context.Context, *hookapi.ClientConnectedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (s *service) OnClientDisconnected(_ context.Context, req *hookapi.ClientDisconnectedRequest) (*hookapi.EmptySuccess, error) {
	s.ctl.Release(req.Clientinfo.Clientid)
	return &hookapi.EmptySuccess{}, nil
}

func (s *service) OnClientserviceenticate(_ context.Context, req *hookapi.ClientAuthenticateRequest) (*hookapi.ValuedResponse, error) {
	return &hookapi.ValuedResponse{
		Type:  hookapi.ValuedResponse_CONTINUE,
		Value: &hookapi.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(_ context.Context, req *hookapi.ClientAuthorizeRequest) (*hookapi.ValuedResponse, error) {
	if !s.ctl.Check(req.Topic, req.Clientinfo.Clientid, req.Type) {
		return &hookapi.ValuedResponse{
			Type:  hookapi.ValuedResponse_STOP_AND_RETURN,
			Value: &hookapi.ValuedResponse_BoolResult{BoolResult: false},
		}, nil
	}

	return &hookapi.ValuedResponse{
		Type: hookapi.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientSubscribe(context.Context, *hookapi.ClientSubscribeRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnClientUnsubscribe(context.Context, *hookapi.ClientUnsubscribeRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionCreated(context.Context, *hookapi.SessionCreatedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionSubscribed(context.Context, *hookapi.SessionSubscribedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionUnsubscribed(context.Context, *hookapi.SessionUnsubscribedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionResumed(context.Context, *hookapi.SessionResumedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionDiscarded(context.Context, *hookapi.SessionDiscardedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionTakenover(context.Context, *hookapi.SessionTakenoverRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnSessionTerminated(context.Context, *hookapi.SessionTerminatedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (s *service) OnMessagePublish(_ context.Context, req *hookapi.MessagePublishRequest) (*hookapi.ValuedResponse, error) {
	return &hookapi.ValuedResponse{
		Type: hookapi.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnMessageDelivered(context.Context, *hookapi.MessageDeliveredRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (*service) OnMessageDropped(context.Context, *hookapi.MessageDroppedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}

func (s *service) OnMessageAcked(_ context.Context, req *hookapi.MessageAckedRequest) (*hookapi.EmptySuccess, error) {
	return &hookapi.EmptySuccess{}, nil
}
