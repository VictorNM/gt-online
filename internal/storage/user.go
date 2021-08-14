package storage

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/victornm/gtonline/internal/auth"
)

type Storage struct {
	db *sqlx.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{db: sqlx.NewDb(db, "mysql")}
}

func (s *Storage) FindUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	u := new(auth.User)
	err := s.db.GetContext(ctx, u, `SELECT email, password, first_name, last_name FROM users WHERE email=?;`, email)
	if err == sql.ErrNoRows {
		return nil, auth.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Storage) CreateUser(ctx context.Context, u auth.User) error {
	stmt, err := s.db.PrepareNamed(
		`INSERT INTO users (email, password, first_name, last_name) VALUES(:email, :password, :first_name, :last_name);`,
	)

	if err != nil {
		return err
	}

	_, err = stmt.ExecContext(ctx, u)
	if err != nil {
		return err
	}
	return nil
}
