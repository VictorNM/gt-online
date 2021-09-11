package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/victornm/gtonline/internal/auth"
	"github.com/victornm/gtonline/internal/storage"
)

type (
	Storage struct {
		db *sqlx.DB
	}

	Config struct {
		Addr string
		User string
		Pass string
		Name string
	}
)

func New(cfg Config) (*Storage, error) {
	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", cfg.User, cfg.Pass, cfg.Addr, cfg.Name))
	if err != nil {
		return nil, fmt.Errorf("open db: %v", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Ping() error {
	return s.db.Ping()
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) FindUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	u := new(auth.User)
	err := s.db.GetContext(ctx, u, `SELECT email, password, first_name, last_name FROM users WHERE email=?;`, email)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
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
