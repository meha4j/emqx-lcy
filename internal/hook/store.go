package hook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/atomic"
)

type record struct {
	stamp   uint
	payload string
}

func (r *record) String() string {
	return fmt.Sprintf("(to_timestamp(%d/1000.0),'%s')", r.stamp, r.payload)
}

type qbuf struct {
	cap uint
	len uint
	req string
	dat strings.Builder
	mux sync.Mutex
}

func (buf *qbuf) add(rec record) (string, bool) {
	buf.mux.Lock()
	defer buf.mux.Unlock()

	if buf.len != 0 {
		buf.dat.WriteRune(',')
	}

	buf.dat.WriteString(rec.String())
	buf.len += 1

	if buf.len == buf.cap {
		query := buf.dat.String()

		buf.dat.Reset()
		buf.dat.WriteString(buf.req)
    buf.len = 0

		return query, true
	}

	return "", false
}

type rule struct {
	top string
	mod string
	ret bool
}

type store struct {
	con *pgxpool.Pool
	ret map[string]*qbuf
	own map[string]*atomic.String
}

func newStore(ctx context.Context, addr string, qcap uint) (*store, error) {
	con, err := pgxpool.New(ctx, addr)

	if err != nil {
		return nil, fmt.Errorf("con: %v", err)
	}

	rows, err := con.Query(ctx, "SELECT top, mod, ret FROM rule WHERE mod = 'ex' OR ret = true")

	if err != nil {
		return nil, fmt.Errorf("qry: %v", err)
	}

	enum := 0
	rnum := 0

	rules, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*rule, error) {
		rule := &rule{}

		if err := row.Scan(&rule.top, &rule.mod, &rule.ret); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}

		if rule.mod == "ex" {
			enum += 1
		}

		if rule.ret {
			rnum += 1
		}

		return rule, nil
	})

	if err != nil {
		return nil, fmt.Errorf("col: %v", err)
	}

	store := &store{
		con: con,
		ret: make(map[string]*qbuf, rnum),
		own: make(map[string]*atomic.String, enum),
	}

	for _, rule := range rules {
		if rule.mod == "ex" {
			store.own[rule.top] = atomic.NewString("")
		}

		if rule.ret {
			store.ret[rule.top] = &qbuf{
				cap: qcap,
				req: fmt.Sprintf("INSERT INTO %s (stamp, payload) VALUES ", rule.top),
			}
		}

		if _, err := con.Exec(ctx,
			fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (stamp TIMESTAMPTZ NOT NULL, payload JSONB NOT NULL)", rule.top),
		); err != nil {
			return nil, fmt.Errorf("create table: %v", err)
		}

		if _, err := con.Exec(ctx,
			fmt.Sprintf("SELECT create_hypertable('%s', by_range('stamp'), if_not_exists => TRUE)", rule.top),
		); err != nil {
			return nil, fmt.Errorf("tune table: %v", err)
		}
	}

	return store, nil
}

func (s *store) save(ctx context.Context, top string, rec record) error {
	buf, ok := s.ret[top]

	if !ok {
		return nil
	}

  slog.Debug("rec", "top", top, "rec", rec)

	if qry, ok := buf.add(rec); ok {
    slog.Debug("save", "qry", qry)

		if _, err := s.con.Exec(ctx, qry); err != nil {
			return fmt.Errorf("exec: %v", err)
		}
	}

	return nil
}

func (s *store) authz(top, con string) bool {
  own, ok := s.own[top]

  if !ok {
    return true
  }

  if own.CompareAndSwap("", con) {
    return true
  }

  if own.Load() == con {
    return true
  }

  return false
}
