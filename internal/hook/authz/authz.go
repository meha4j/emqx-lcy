package authz

import (
	"context"
	"fmt"
	"sync"

	api "github.com/blabtm/extd/internal/api/hook"

	"github.com/blabtm/extd/internal/hook"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/atomic"
)

func New(pdb *pgxpool.Pool) hook.Instance {
	return &authz{pdb: pdb}
}

type authz struct {
	hook.Base

	pdb *pgxpool.Pool
	own sync.Map
}

func (s *authz) OnProviderLoaded(ctx context.Context, req *api.ProviderLoadedRequest) (*api.LoadedResponse, error) {
	rows, err := s.pdb.Query(ctx, "SELECT top FROM rule WHERE mod = 'ex'")

	if err != nil {
		return nil, fmt.Errorf("db: query: %v", err)
	}

	own, err := pgx.CollectRows(rows, pgx.RowTo[string])

	if err != nil {
		return nil, fmt.Errorf("db: collect: %v", err)
	}

	s.own.Clear()

	for _, top := range own {
		s.own.Store(top, atomic.NewString(""))
	}

	return &api.LoadedResponse{
		Hooks: []*api.HookSpec{
			{Name: "client.authorize"},
		},
	}, nil
}

func (s *authz) OnClientAuthorize(_ context.Context, req *api.ClientAuthorizeRequest) (*api.ValuedResponse, error) {
	if req.Type != api.ClientAuthorizeRequest_PUBLISH {
		return &api.ValuedResponse{
			Type:  api.ValuedResponse_CONTINUE,
			Value: &api.ValuedResponse_BoolResult{BoolResult: true},
		}, nil
	}

	own, ok := s.own.Load(req.Topic)

	if !ok {
		return &api.ValuedResponse{
			Type:  api.ValuedResponse_CONTINUE,
			Value: &api.ValuedResponse_BoolResult{BoolResult: true},
		}, nil
	}

	if own.(*atomic.String).CompareAndSwap("", req.Clientinfo.Clientid) {
		return &api.ValuedResponse{
			Type:  api.ValuedResponse_CONTINUE,
			Value: &api.ValuedResponse_BoolResult{BoolResult: true},
		}, nil
	}

	if own.(*atomic.String).Load() == req.Clientinfo.Clientid {
		return &api.ValuedResponse{
			Type:  api.ValuedResponse_CONTINUE,
			Value: &api.ValuedResponse_BoolResult{BoolResult: true},
		}, nil
	}

	return &api.ValuedResponse{
		Type:  api.ValuedResponse_STOP_AND_RETURN,
		Value: &api.ValuedResponse_BoolResult{BoolResult: false},
	}, nil
}

func (s *authz) OnClientDisconnected(_ context.Context, req *api.ClientDisconnectedRequest) (*api.EmptySuccess, error) {
  s.own.Range(func(key, value any) bool {
    s.own.CompareAndSwap(key, req.Clientinfo.Clientid, "")
    return true
  })

	return &api.EmptySuccess{}, nil
}
