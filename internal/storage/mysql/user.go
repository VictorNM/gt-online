package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/victornm/gtonline/internal/auth"
	"github.com/victornm/gtonline/internal/friend"
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

func (s *Storage) SearchUsers(ctx context.Context, req friend.SearchFriendsRequest) (*friend.SearchFriendsResponse, error) {
	type row struct {
		Email     string         `db:"email"`
		FirstName string         `db:"first_name"`
		LastName  string         `db:"last_name"`
		Hometown  sql.NullString `db:"hometown"`
	}

	buildWhere := func(req friend.SearchFriendsRequest) (string, []interface{}) {
		var (
			condition []string
			args      []interface{}
		)

		if req.Email != nil {
			condition = append(condition, "email=?")
			args = append(args, *req.Email)
		}

		if req.Name != nil {
			condition = append(condition, "(first_name LIKE ? OR last_name LIKE ?)")
			pattern := "%" + *req.Name + "%"
			args = append(args, pattern, pattern)
		}

		if req.Hometown != nil {
			condition = append(condition, "hometown LIKE ?")
			args = append(args, "%"+*req.Hometown+"%")
		}

		return strings.Join(condition, " OR "), args
	}

	where, args := buildWhere(req)

	stmt := `
SELECT u.email, first_name, last_name, hometown 
FROM users as u 
JOIN regular_users as ru 
USING (email)
WHERE ` + where + ";"

	var rows []row
	err := s.db.SelectContext(ctx, &rows, stmt, args...)
	if err != nil {
		return nil, err
	}

	res := &friend.SearchFriendsResponse{
		Count: len(rows),
	}
	for _, r := range rows {
		res.Users = append(res.Users, friend.User{
			Email:     r.Email,
			FirstName: r.FirstName,
			LastName:  r.LastName,
			Hometown:  r.Hometown.String,
		})
	}
	return res, nil
}
