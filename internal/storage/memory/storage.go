package memory

import (
	"context"
	"sync"

	"github.com/victornm/gtonline/internal/friend"
	"github.com/victornm/gtonline/internal/profile"
	"github.com/victornm/gtonline/internal/storage"
)

type (
	Storage struct {
		usersMu sync.Mutex
		users   []User

		friendshipsMu sync.Mutex
		friendships   []friend.Friendship
	}

	User profile.Profile
)

func (s *Storage) ListFriends(_ context.Context, email string) ([]*friend.Friendship, error) {
	s.friendshipsMu.Lock()
	defer s.friendshipsMu.Unlock()

	var res []*friend.Friendship

	for _, f := range s.friendships {
		if f.Email == email && !f.DateConnected.IsZero() {
			out := f
			res = append(res, &out)
		}
	}

	return res, nil
}

func (s *Storage) ListPendingFriendships(_ context.Context, email string) ([]*friend.Friendship, error) {
	s.friendshipsMu.Lock()
	defer s.friendshipsMu.Unlock()

	var res []*friend.Friendship

	for _, f := range s.friendships {
		if (f.Email == email || f.FriendEmail == email) && f.DateConnected.IsZero() {
			out := f
			res = append(res, &out)
		}
	}

	return res, nil
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) InsertUsers(users []User) {
	s.usersMu.Lock()
	s.users = append(s.users, users...)
	s.usersMu.Unlock()
}

func (s *Storage) SearchUsers(ctx context.Context, req friend.SearchFriendsRequest) (*friend.SearchFriendsResponse, error) {
	panic("implement me")
}

func (s *Storage) GetFriendship(_ context.Context, email, friendEmail string) (*friend.Friendship, error) {
	s.friendshipsMu.Lock()
	defer s.friendshipsMu.Unlock()

	for _, f := range s.friendships {
		if f.Email == email && f.FriendEmail == friendEmail {
			out := f
			return &out, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (s *Storage) InsertFriendship(ctx context.Context, f *friend.Friendship) error {
	if _, err := s.getUser(f.Email); err != nil {
		return storage.ErrInvalidArgument
	}

	if _, err := s.getUser(f.FriendEmail); err != nil {
		return storage.ErrInvalidArgument
	}

	if _, err := s.GetFriendship(ctx, f.Email, f.FriendEmail); err == nil {
		return storage.ErrAlreadyExist
	}

	s.friendshipsMu.Lock()
	s.friendships = append(s.friendships, *f)
	s.friendshipsMu.Unlock()
	return nil
}

func (s *Storage) UpdateFriendship(_ context.Context, f *friend.Friendship) error {
	s.friendshipsMu.Lock()
	defer s.friendshipsMu.Unlock()

	for i, f1 := range s.friendships {
		if f.Email == f1.Email && f.FriendEmail == f1.FriendEmail {
			s.friendships[i] = *f
			return nil
		}
	}

	return storage.ErrNotFound
}

func (s *Storage) getUser(email string) (*User, error) {
	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	for _, u := range s.users {
		if u.Email == email {
			out := u
			return &out, nil
		}
	}
	return nil, storage.ErrNotFound
}
