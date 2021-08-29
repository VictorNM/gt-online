package user

import (
	"context"
	"time"

	"github.com/victornm/gtonline/internal/gterr"
)

type (
	UpdateProfileRequest struct {
		Email       string    `json:"email"`
		Sex         string    `json:"sex"`
		Birthdate   time.Time `json:"birthdate"`
		CurrentCity string    `json:"current_city"`
		Hometown    string    `json:"hometown"`
		Interests   []string  `json:"interests"`
		Education   []struct {
			School        string `json:"school"`
			YearGraduated int    `json:"year_graduated"`
		} `json:"education"`
		Professional []struct {
			Employer string `json:"employer"`
			JobTitle string `json:"job_title"`
		} `json:"professional"`
	}

	Profile struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		UpdateProfileRequest
	}
)

type (
	Service struct {
		storage Storage
	}

	Storage interface {
		UpdateProfile(ctx context.Context, req UpdateProfileRequest) error
	}
)

func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
	err := s.storage.UpdateProfile(ctx, req)
	if err == gterr.ErrNotFound {
		return nil, gterr.New(gterr.NotFound, "User not exist", err)
	}
	return s.GetProfile(ctx, req.Email)
}

func (s *Service) GetProfile(ctx context.Context, email string) (*Profile, error) {
	return nil, nil
}
