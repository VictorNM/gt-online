package storage

import (
	"context"
	"fmt"

	"github.com/victornm/gtonline/internal/profile"
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

func (s *Storage) ListSchools(ctx context.Context) ([]profile.School, error) {
	var schools []profile.School

	stmt := `SELECT school_name, type FROM schools;`
	if err := s.db.SelectContext(ctx, &schools, stmt); err != nil {
		return nil, fmt.Errorf("query schools: %v", err)
	}

	return schools, nil
}

func (s *Storage) ListEmployers(ctx context.Context) ([]profile.Employer, error) {
	var employers []profile.Employer

	stmt := `SELECT employer_name FROM employers;`
	if err := s.db.SelectContext(ctx, &employers, stmt); err != nil {
		return nil, fmt.Errorf("query employers: %v", err)
	}

	return employers, nil
}
