package authz

import (
	"context"

	api "github.com/blabtm/extd/internal/api/hook"

	"github.com/blabtm/extd/internal/hook"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(pool *pgxpool.Pool) hook.Instance {
  return &authz{pool: pool}
}

type authz struct {
	hook.Base

  pool *pgxpool.Pool
}

func (s *authz) OnProviderLoaded(ctx context.Context, req *api.ProviderLoadedRequest) (*api.LoadedResponse, error) {
	return nil, nil
}
