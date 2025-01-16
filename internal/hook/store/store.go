package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/blabtm/extd/internal/hook"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fastjson"

	api "github.com/blabtm/extd/internal/api/hook"
)

const (
	QueryInit = `
    CREATE TABLE IF NOT EXISTS %s (
      stamp TIMESTAMPTZ NOT NULL,
      payload JSONB NOT NULL
    );

    SELECT create_hypertable('%s', by_range('stamp'), if_not_exists => TRUE)
  `
)

func New(con *pgxpool.Pool) hook.Instance {
	return &store{con: con}
}

type store struct {
	hook.Base

	con *pgxpool.Pool
	ret sync.Map
}

func (s *store) OnProviderLoaded(ctx context.Context, req *api.ProviderLoadedRequest) (*api.LoadedResponse, error) {
	rows, err := s.con.Query(ctx, "SELECT top FROM rule WHERE ret = true")

	if err != nil {
		return nil, fmt.Errorf("db: query: %v", err)
	}

	ret, err := pgx.CollectRows(rows, pgx.RowTo[string])

	if err != nil {
		return nil, fmt.Errorf("db: collect: %v", err)
	}

	s.ret.Clear()

	for _, top := range ret {
		s.ret.Store(top, &queryBuffer{
			top: top,
			buf: make([]*record, 0, 5),
		})

		if _, err := s.con.Exec(ctx, fmt.Sprintf(QueryInit, top, top)); err != nil {
			return nil, fmt.Errorf("db: exec: %v", err)
		}
	}

	return &api.LoadedResponse{
		Hooks: []*api.HookSpec{
			{Name: "message.publish"},
		},
	}, nil
}

func (s *store) OnMessagePublish(ctx context.Context, req *api.MessagePublishRequest) (*api.ValuedResponse, error) {
	buf, ok := s.ret.Load(req.Message.Topic)

	if ok {
		rec, err := parseRecord(req)

		if err != nil {
			return nil, fmt.Errorf("parse: %v", err)
		}

		if query, ok := buf.(*queryBuffer).Append(rec); ok {
			if _, err := s.con.Exec(ctx, query); err != nil {
				return nil, fmt.Errorf("db: exec: %v", err)
			}
		}
	}

	return &api.ValuedResponse{
		Type:  api.ValuedResponse_CONTINUE,
		Value: &api.ValuedResponse_Message{Message: req.Message},
	}, nil
}

func parseRecord(req *api.MessagePublishRequest) (*record, error) {
	pay, err := fastjson.ParseBytes(req.Message.Payload)

	if err != nil {
		return nil, fmt.Errorf("json: %v", err)
	}

	obj := pay.GetObject()

	var r record

	r.Stamp = obj.Get("stamp").GetUint64()
	obj.Del("stamp")
	r.Payload = obj.MarshalTo(r.Payload)

	return &r, nil
}

type record struct {
	Stamp   uint64
	Payload []byte
}

type queryBuffer struct {
	top string
	buf []*record
	mux sync.Mutex
}

func (q *queryBuffer) Append(rec *record) (string, bool) {
	q.mux.Lock()
	q.mux.Unlock()

	q.buf = append(q.buf, rec)

	if len(q.buf) == cap(q.buf) {
		var sb strings.Builder

		sb.Grow(40 + len(q.top))
		sb.WriteString("INSERT INTO ")
		sb.WriteString(q.top)
		sb.WriteString(" (stamp, payload) VALUES ")

		for i, r := range q.buf {
			sb.Grow(40 + len(r.Payload))

			if i != 0 {
				sb.WriteRune(',')
			}

			sb.WriteString("(to_timestamp(")
			sb.WriteString(strconv.FormatUint(r.Stamp, 10))
			sb.WriteString("/1000.0), '")
			sb.Write(r.Payload)
			sb.WriteString("')")
		}

		q.buf = q.buf[:0]

		return sb.String(), true
	}

	return "", false
}
