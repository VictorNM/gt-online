package friend

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/storage"
)

type (
	Service struct {
		storage Storage
	}

	Storage interface {
		SearchUsers(ctx context.Context, req SearchFriendsRequest) (*SearchFriendsResponse, error)
		ListPendingFriendships(ctx context.Context, email string) ([]*Friendship, error)
		GetFriendship(ctx context.Context, email, friendEmail string) (*Friendship, error)
		InsertFriendship(ctx context.Context, f *Friendship) error
		UpdateFriendship(ctx context.Context, f *Friendship) error
	}

	Friendship struct {
		Email         string
		FriendEmail   string
		Relationship  string
		DateConnected time.Time
	}
)

func NewService(s Storage) *Service {
	return &Service{storage: s}
}

type (
	SearchFriendsRequest struct {
		Email    string `form:"email"`
		Name     string `form:"name"`
		Hometown string `form:"hometown"`
	}

	SearchFriendsResponse struct {
		Count int
		Users []User
	}

	User struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Hometown  string `json:"hometown"`
	}

	ListFriendRequestResponse struct {
		// RequestTo is the requests has been sent from the input email
		RequestTo []Request `json:"request_to"`
		// RequestFrom is the requests has been sent to the input email
		RequestFrom []Request `json:"request_from"`
	}

	Request struct {
		Email        string `json:"email"`
		Relationship string `json:"relationship"`
	}

	CreateFriendRequest struct {
		Email        string
		FriendEmail  string
		Relationship string `json:"relationship"`
	}
)

func (s *Service) SearchFriends(ctx context.Context, req SearchFriendsRequest) (*SearchFriendsResponse, error) {
	res, err := s.storage.SearchUsers(ctx, req)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return res, nil
}

func (s *Service) ListFriendRequests(ctx context.Context, email string) (*ListFriendRequestResponse, error) {
	friendships, err := s.storage.ListPendingFriendships(ctx, email)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	res := new(ListFriendRequestResponse)
	for _, f := range friendships {
		if strings.EqualFold(email, f.Email) {
			res.RequestTo = append(res.RequestTo, Request{
				Email:        f.FriendEmail,
				Relationship: f.Relationship,
			})
		}
		if strings.EqualFold(email, f.FriendEmail) {
			res.RequestFrom = append(res.RequestFrom, Request{
				Email:        f.Email,
				Relationship: f.Relationship,
			})
		}
	}

	return res, nil
}

func (s *Service) CreateFriend(ctx context.Context, req CreateFriendRequest) error {
	if strings.EqualFold(req.Email, req.FriendEmail) {
		return gterr.New(gterr.InvalidArgument, "can't be friend with yourself")
	}

	f, err := s.storage.GetFriendship(ctx, req.Email, req.FriendEmail)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return gterr.New(gterr.Internal, "", fmt.Errorf("get friendship: %v", err))
	}

	if errors.Is(err, storage.ErrNotFound) {
		return s.insertFriendship(ctx, req)
	}

	if !f.DateConnected.IsZero() {
		return gterr.New(gterr.AlreadyExists, fmt.Sprintf("%s and %s already friends", req.Email, req.FriendEmail))
	}

	f.Relationship = req.Relationship
	if err := s.storage.UpdateFriendship(ctx, f); err != nil {
		return gterr.New(gterr.Internal, "", err)
	}
	return nil
}

func (s *Service) insertFriendship(ctx context.Context, req CreateFriendRequest) error {
	err := s.storage.InsertFriendship(ctx, &Friendship{
		Email:        req.Email,
		FriendEmail:  req.FriendEmail,
		Relationship: req.Relationship,
	})
	if errors.Is(err, storage.ErrInvalidArgument) {
		msg := fmt.Sprintf("the requested email is not found: email=%s friend_email=%s", req.Email, req.FriendEmail)
		return gterr.New(gterr.NotFound, msg, err)
	}

	if err != nil {
		return gterr.New(gterr.Internal, "", fmt.Errorf("insert friendship: %v", err))
	}

	return nil
}
