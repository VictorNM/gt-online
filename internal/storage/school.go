package storage

import (
	"context"
	"fmt"

	"github.com/victornm/gtonline/internal/profile"
)

func (s *Storage) ListSchools(ctx context.Context) ([]profile.School, error) {
	var schools []profile.School

	stmt := `SELECT school_name, type FROM schools;`
	if err := s.db.SelectContext(ctx, &schools, stmt); err != nil {
		return nil, fmt.Errorf("query schools: %v", err)
	}

	return schools, nil
}
