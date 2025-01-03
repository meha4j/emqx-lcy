package auth

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/paraskun/extd/api/auth"
)

type Action = auth.ClientAuthorizeRequest_AuthorizeReqType

type Rule interface {
	Check(con string, act Action) bool
}

type ExclusiveRule struct {
	Owner atomic.Pointer[string]
}

func (r *ExclusiveRule) Check(con string, act Action) bool {
	if r.Owner.CompareAndSwap(nil, &con) {
		return true
	}

	if *r.Owner.Load() == con {
		return true
	}

	return false
}

type ReadOnlyRule struct{}

func (r *ReadOnlyRule) Check(con string, act Action) bool {
	if act == auth.ClientAuthorizeRequest_PUBLISH {
		return false
	}

	return true
}

type WriteOnlyRule struct{}

func (r *WriteOnlyRule) Check(con string, act Action) bool {
	if act == auth.ClientAuthorizeRequest_SUBSCRIBE {
		return false
	}

	return true
}

type Store struct {
	dat map[string][]Rule
	mux sync.RWMutex
}

func NewStore() (*Store, error) {
	return &Store{}, nil
}

func (s *Store) Fetch(con *sql.DB) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	res, err := con.Query("SELECT COUNT(DISTINCT top) FROM rule")

	if err != nil {
		return fmt.Errorf("query: %v", err)
	}

	var (
		num int
		top string
		mod string
	)

	for res.Next() {
		if err := res.Scan(&num); err != nil {
			return fmt.Errorf("scan: %v", err)
		}
	}

	if res.Err() != nil {
		return fmt.Errorf("scan: %v", res.Err())
	}

	s.dat = make(map[string][]Rule, num)
	res, err = con.Query("SELECT top, mod FROM rule")

	if err != nil {
		return fmt.Errorf("query: %v", err)
	}

	for res.Next() {
		if err := res.Scan(&top, &mod); err != nil {
			return fmt.Errorf("scan: %v", err)
		}

		ctl, ok := s.dat[top]

		if !ok {
			ctl = make([]Rule, 0, 1)
		}

		switch mod {
		case "ex":
			ctl = append(ctl, &ExclusiveRule{})
		case "ro":
			ctl = append(ctl, &ReadOnlyRule{})
		case "wo":
			ctl = append(ctl, &WriteOnlyRule{})
		}

		s.dat[top] = ctl
	}

	return nil
}
