package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/victornm/gtonline/internal/friend"
	"github.com/victornm/gtonline/internal/storage"
)

type friendship struct {
	Email         string         `db:"email"`
	FriendEmail   string         `db:"friend_email"`
	Relationship  sql.NullString `db:"relationship"`
	DateConnected sql.NullTime   `db:"date_connected"`
}

func (s *Storage) ListPendingFriendships(ctx context.Context, email string) ([]*friend.Friendship, error) {
	stmt := `
SELECT email, friend_email, relationship
FROM friendships
WHERE (email=? OR friend_email=?) AND date_connected IS NULL;
`
	var rows []friendship
	err := s.db.SelectContext(ctx, &rows, stmt, email, email)
	if err != nil {
		return nil, err
	}

	res := make([]*friend.Friendship, 0, len(rows))
	for _, r := range rows {
		res = append(res, &friend.Friendship{
			Email:        r.Email,
			FriendEmail:  r.FriendEmail,
			Relationship: r.Relationship.String,
		})
	}
	return res, nil
}

func (s *Storage) GetFriendship(ctx context.Context, email, friendEmail string) (*friend.Friendship, error) {
	var row friendship

	err := s.db.GetContext(ctx, &row, `
SELECT email, friend_email, relationship, date_connected 
FROM friendships 
WHERE email=? 
  AND friend_email=?;`, email, friendEmail)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &friend.Friendship{
		Email:         row.Email,
		FriendEmail:   row.FriendEmail,
		Relationship:  row.Relationship.String,
		DateConnected: row.DateConnected.Time,
	}, nil
}

func (s *Storage) InsertFriendship(ctx context.Context, f *friend.Friendship) error {
	row := friendship{
		Email:       f.Email,
		FriendEmail: f.FriendEmail,
	}
	if f.Relationship != "" {
		row.Relationship = sql.NullString{
			String: f.Relationship,
			Valid:  true,
		}
	}

	stmt := `
INSERT INTO friendships (email, friend_email, relationship)
VALUES (:email, :friend_email, :relationship);`

	_, err := s.db.NamedExecContext(ctx, stmt, row)
	if isDuplicate(err) {
		return fmt.Errorf("%w: %v", storage.ErrAlreadyExist, err)
	}
	if isErrForeignKeyConstraint(err) {
		return fmt.Errorf("%w: %v", storage.ErrInvalidArgument, err)
	}
	return err
}

func (s *Storage) UpdateFriendship(ctx context.Context, f *friend.Friendship) error {
	row := friendship{
		Email:       f.Email,
		FriendEmail: f.FriendEmail,
	}
	if f.Relationship != "" {
		row.Relationship = sql.NullString{
			String: f.Relationship,
			Valid:  true,
		}
	}

	if !f.DateConnected.IsZero() {
		row.DateConnected = sql.NullTime{
			Time:  f.DateConnected,
			Valid: true,
		}
	}

	stmt := `
UPDATE friendships 
SET relationship=:relationship, date_connected=:date_connected
WHERE email=:email AND friend_email=:friend_email;
`
	_, err := s.db.NamedExecContext(ctx, stmt, row)
	if err != nil {
		return err
	}
	return nil
}
