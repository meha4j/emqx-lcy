package auth

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/paraskun/extd/api/auth"
)

type Action = auth.ClientAuthorizeRequest_AuthorizeReqType

type ACL struct {
	dat map[string]*atomic.Pointer[string]
	mux sync.RWMutex
}

func (ctl *ACL) Check(top, con string, act Action) bool {
	if act != auth.ClientAuthorizeRequest_PUBLISH {
		return true
	}

	ctl.mux.RLock()
	defer ctl.mux.RUnlock()

	own, ok := ctl.dat[top]

	if !ok {
		return true
	}

	if own.CompareAndSwap(nil, &con) {
		return true
	}

	if *own.Load() == con {
		return true
	}

	return false
}

func (ctl *ACL) Fetch(con *sql.DB) error {
	ctl.mux.Lock()
	defer ctl.mux.Unlock()

	ctl.dat = make(map[string]*atomic.Pointer[string])
	res, err := con.Query("SELECT top FROM rule WHERE mod = 'ex'")

	if err != nil {
		return fmt.Errorf("query: %v", err)
	}

	var top string

	for res.Next() {
		if err := res.Scan(&top); err != nil {
			return fmt.Errorf("scan: %v", err)
		}

		ctl.dat[top] = &atomic.Pointer[string]{}
	}

	if res.Err() != nil {
		return fmt.Errorf("scan: %v", res.Err())
	}

	return nil
}
