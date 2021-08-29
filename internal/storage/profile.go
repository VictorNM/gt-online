package storage

import (
	"context"
	"fmt"
)

func (s *Storage) isEmailExist(ctx context.Context, table string, email string) (bool, error) {
	stmt := fmt.Sprintf(`SELECT EXISTS(SELECT email FROM %s WHERE email=?)`, table)

	r, err := s.db.QueryContext(ctx, stmt, email)
	if err != nil {
		return false, fmt.Errorf("query: %v", err)
	}
	var exist bool
	for r.Next() {
		if err := r.Scan(&exist); err != nil {
			return false, fmt.Errorf("scan: %v", err)
		}
	}

	return exist, nil
}
