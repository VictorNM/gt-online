package friend

import (
	"context"

	"github.com/victornm/gtonline/internal/gterr"
)

type (
	Service struct {
		storage Storage
	}

	Storage interface {
		SearchUsers(ctx context.Context, req SearchFriendsRequest) (*SearchFriendsResponse, error)
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
)

func (s *Service) SearchFriends(ctx context.Context, req SearchFriendsRequest) (*SearchFriendsResponse, error) {
	res, err := s.storage.SearchUsers(ctx, req)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return res, nil
}
