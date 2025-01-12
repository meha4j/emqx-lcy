package hook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"github.com/valyala/fastjson"

	"google.golang.org/grpc"

	"github.com/paraskun/extd/emqx"
	"github.com/paraskun/extd/internal/api/hook"
)

const (
	SMT = "INSERT INTO %s (timestamp, payload) VALUES (to_timestamp(%d/1000.0), '%s')"
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
		if err := opt.cli.UpdateExHookServer(&emqx.ExHookServerUpdateRequest{
			Name:      cfg.GetString("extd.hook.name"),
			Enable:    cfg.GetBool("extd.hook.enable"),
			Action:    cfg.GetString("extd.hook.action"),
			Timeout:   cfg.GetString("extd.hook.tout"),
			Reconnect: cfg.GetString("extd.hook.trec"),
			Addr:      fmt.Sprintf("http://%s:%d", opt.cli.Addr, cfg.GetInt("extd.port")),
		}); err != nil {
			return fmt.Errorf("upd: %v", err)
		}
	}

	ctx := context.Background()
	con, err := pgxpool.New(ctx, "postgres://postgres:pass@psql:5432/postgres")

	if err != nil {
		return fmt.Errorf("db: %v", err)
	}

	hook.RegisterHookProviderServer(srv, &service{con: con})

	return nil
}

type series struct {
	Cap uint

	len uint
  tpl string
	dat *strings.Builder
	mux sync.Mutex
}

func newSeries(top string, cap uint) *series {
  tpl := fmt.Sprintf("INSERT INTO %s (timestamp, payload) VALUES ", top)
	dat := &strings.Builder{}

	dat.WriteString(tpl)

	return &series{
		Cap: cap,
    tpl: tpl,
		dat: dat,
	}
}

func (s *series) append(qry string) (string, bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.len != 0 {
		s.dat.WriteRune(',')
	}

	s.dat.WriteString(qry)
	s.len += 1

	if s.len == s.Cap {
		qry = s.dat.String()

		s.dat.Reset()
    s.dat.WriteString(s.tpl)
		s.len = 0

		return qry, true
	} else {
		return "", false
	}
}

type service struct {
	con *pgxpool.Pool
	que sync.Map

	hook.UnimplementedHookProviderServer
}

func (s *service) OnProviderLoaded(ctx context.Context, _ *hook.ProviderLoadedRequest) (*hook.LoadedResponse, error) {
	return &hook.LoadedResponse{
		Hooks: []*hook.HookSpec{
			{Name: "client.authorize"},
			{Name: "message.publish"},
		},
	}, nil
}

func (s *service) OnProviderUnloaded(context.Context, *hook.ProviderUnloadedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientConnect(context.Context, *hook.ClientConnectRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientConnack(context.Context, *hook.ClientConnackRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientConnected(context.Context, *hook.ClientConnectedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientDisconnected(_ context.Context, req *hook.ClientDisconnectedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnClientAuthenticate(_ context.Context, req *hook.ClientAuthenticateRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func (s *service) OnClientAuthorize(_ context.Context, req *hook.ClientAuthorizeRequest) (*hook.ValuedResponse, error) {
	return &hook.ValuedResponse{
		Type: hook.ValuedResponse_CONTINUE,
	}, nil
}

func (*service) OnClientSubscribe(context.Context, *hook.ClientSubscribeRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnClientUnsubscribe(context.Context, *hook.ClientUnsubscribeRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionCreated(context.Context, *hook.SessionCreatedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionSubscribed(context.Context, *hook.SessionSubscribedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionUnsubscribed(context.Context, *hook.SessionUnsubscribedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionResumed(context.Context, *hook.SessionResumedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionDiscarded(context.Context, *hook.SessionDiscardedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionTakenover(context.Context, *hook.SessionTakenoverRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnSessionTerminated(context.Context, *hook.SessionTerminatedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessagePublish(ctx context.Context, req *hook.MessagePublishRequest) (*hook.ValuedResponse, error) {
	que, _ := s.que.LoadOrStore(req.Message.Topic, newSeries(req.Message.Topic, 5))
	qry, err := toQuery(req.Message.Payload)

	if err != nil {
		return nil, fmt.Errorf("pay: %v", err)
	}

	if qry, ok := que.(*series).append(qry); ok {
    slog.Debug("store", "query", qry)
		s.con.Exec(ctx, qry)
	}

	return &hook.ValuedResponse{
		Type:  hook.ValuedResponse_CONTINUE,
		Value: &hook.ValuedResponse_BoolResult{BoolResult: true},
	}, nil
}

func toQuery(pay []byte) (string, error) {
	json, err := fastjson.ParseBytes(pay)

	if err != nil {
		return "", fmt.Errorf("json: %v", err)
	}

	obj, err := json.Object()

	if err != nil {
		return "", fmt.Errorf("pay: %v", err)
	}

	time := obj.Get("timestamp").GetUint64()
	obj.Del("timestamp")
	data := obj.MarshalTo(make([]byte, 0, 60))

	return fmt.Sprintf("(to_timestamp(%d/1000.0), '%s')", time, string(data)), nil
}

func (*service) OnMessageDelivered(context.Context, *hook.MessageDeliveredRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (*service) OnMessageDropped(context.Context, *hook.MessageDroppedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}

func (s *service) OnMessageAcked(_ context.Context, req *hook.MessageAckedRequest) (*hook.EmptySuccess, error) {
	return &hook.EmptySuccess{}, nil
}
