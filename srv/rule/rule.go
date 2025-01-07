package rule

import (
	"database/sql"
	"fmt"
)

type RuleStore struct {
	db *sql.DB
}

func NewRuleStore(db *sql.DB) RuleStore {
	return RuleStore{db: db}
}

type Rule struct {
	Top string
	Mod string
}

func (s *RuleStore) GetByTop(top string) (res []Rule, err error) {
	rows, err := s.db.Query("SELECT top, mod FROM rule WHERE top = ?", top)

	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	for rows.Next() {
		var (
			top string
			mod string
		)

		if err := rows.Scan(&top, &mod); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}

		res = append(res, Rule{top, mod})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("scan: %v", err)
	}

	return
}

func (s *RuleStore) GetByMod(mod string) (res []Rule, err error) {
	rows, err := s.db.Query("SELECT top, mod FROM rule WHERE mod = ?", mod)

	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	for rows.Next() {
		var (
			top string
			mod string
		)

		if err := rows.Scan(&top, &mod); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}

		res = append(res, Rule{top, mod})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("scan: %v", err)
	}

	return
}
