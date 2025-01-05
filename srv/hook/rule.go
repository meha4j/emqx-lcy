package auth

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/paraskun/extd/api/auth"
)

type Action = auth.ClientAuthorizeRequest_AuthorizeReqType

const (
	PUB = auth.ClientAuthorizeRequest_PUBLISH
	SUB = auth.ClientAuthorizeRequest_SUBSCRIBE
)

type ACL struct {
	sync.Map
}

func (ctl *ACL) Check(top, con string, act Action) bool {
	if act != PUB {
		return true
	}

	ctl.CompareAndSwap(top, nil, con)

	if own, ok := ctl.Load(top); ok && own != con {
		return false
	}

	return true
}

func (ctl *ACL) Release(con string) {
	ctl.Range(func(key, value any) bool {
		ctl.CompareAndSwap(key, con, nil)
		return true
	})
}

func (ctl *ACL) Fetch(con *sql.DB, query string) error {
	ctl.Clear()

	res, err := con.Query(query)

	if err != nil {
		return fmt.Errorf("query: %v", err)
	}

	var top string

	for res.Next() {
		if err := res.Scan(&top); err != nil {
			return fmt.Errorf("scan: %v", err)
		}

		ctl.Store(top, nil)
	}

	if res.Err() != nil {
		return fmt.Errorf("scan: %v", res.Err())
	}

	return nil
}
