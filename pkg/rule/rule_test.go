//go:build integration

package rule

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	Name = "postgres"
	User = "postgres"
	Pass = "password"
)

func TestGetByTop(t *testing.T) {
	ctx := context.Background()

	cnt, err := postgres.Run(ctx, "postgres:latest",
		postgres.WithPassword(Pass),
		postgres.WithInitScripts(filepath.Join(".", "docker-entrypoint-initdb.d", "init.sql")),
		testcontainers.WithWaitStrategy(wait.
			ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5*time.Second),
		),
	)

	assert.Nil(t, err)

	con, err := cnt.ConnectionString(ctx)
	assert.Nil(t, err)

	db, err := sql.Open("pgx", con)
	assert.Nil(t, err)

	rs := NewRuleStore(db)

	res, err := rs.GetByTop("test1")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "test1", res[0].Top)
	assert.Equal(t, "ex", res[0].Mod)

	testcontainers.CleanupContainer(t, cnt)
}

func TestGetByMod(t *testing.T) {
}
