package friend

import (
	"context"
	"encoding/json"
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
		ListFriends(ctx context.Context, email string) ([]*Friendship, error)
		ListPendingFriendships(ctx context.Context, email string) ([]*Friendship, error)
		GetFriendship(ctx context.Context, email, friendEmail string) (*Friendship, error)
		InsertFriendship(ctx context.Context, f *Friendship) error
		UpdateFriendship(ctx context.Context, f *Friendship) error
		DeleteFriendRequest(ctx context.Context, email, friendEmail string) error
	}
)

func NewService(s Storage) *Service {
	return &Service{storage: s}
}

type (
	Friendship struct {
		Email         string    `json:"-"`
		FriendEmail   string    `json:"friend_email"`
		Relationship  string    `json:"relationship"`
		DateConnected time.Time `json:"date_connected"`
	}

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

	ListFriendsResponse struct {
		Friends []Friendship `json:"friends"`
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

	AcceptFriendRequest struct {
		Email        string
		EmailRequest string
	}

	DeleteFriendRequest struct {
		Email       string
		FriendEmail string
	}
)

func (f Friendship) MarshalJSON() ([]byte, error) {
	type alias Friendship

	data := struct {
		alias
		DateConnected string `json:"date_connected"`
	}{
		alias: alias(f),
	}
	if !f.DateConnected.IsZero() {
		data.DateConnected = f.DateConnected.Format("January 02, 2006")
	}

	return json.Marshal(data)
}

func (s *Service) SearchFriends(ctx context.Context, req SearchFriendsRequest) (*SearchFriendsResponse, error) {
	res, err := s.storage.SearchUsers(ctx, req)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return res, nil
}

func (s *Service) ListFriend(ctx context.Context, email string) (*ListFriendsResponse, error) {
	friendships, err := s.storage.ListFriends(ctx, email)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	res := new(ListFriendsResponse)
	for _, f := range friendships {
		res.Friends = append(res.Friends, Friendship{
			FriendEmail:   f.FriendEmail,
			Relationship:  f.Relationship,
			DateConnected: f.DateConnected,
		})
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

func (s *Service) AcceptFriendRequest(ctx context.Context, req AcceptFriendRequest) error {
	if strings.EqualFold(req.Email, req.EmailRequest) {
		return gterr.New(gterr.InvalidArgument, "2 email must be different")
	}

	f, err := s.storage.GetFriendship(ctx, req.EmailRequest, req.Email)
	if storage.IsErrNotFound(err) {
		msg := fmt.Sprintf("the friend request from %s to %s is not exist", req.EmailRequest, req.Email)
		return gterr.New(gterr.FailedPrecondition, msg, err)
	}

	if !f.DateConnected.IsZero() {
		msg := fmt.Sprintf("%s already accept the request from %s", req.Email, req.EmailRequest)
		return gterr.New(gterr.AlreadyExists, msg)
	}

	f.DateConnected = time.Now()
	if err := s.storage.UpdateFriendship(ctx, f); err != nil {
		return gterr.New(gterr.Internal, "", err)
	}

	return nil
}

func (s *Service) CancelFriendRequest(ctx context.Context, req DeleteFriendRequest) error {
	if err := s.storage.DeleteFriendRequest(ctx, req.Email, req.FriendEmail); err != nil {
		return gterr.New(gterr.Internal, "failed to delete friend request", err)
	}
	return nil
}

func (s *Service) RejectFriendRequest(ctx context.Context, req DeleteFriendRequest) error {
	if err := s.storage.DeleteFriendRequest(ctx, req.FriendEmail, req.Email); err != nil {
		return gterr.New(gterr.Internal, "failed to delete friend request", err)
	}
	return nil
}
