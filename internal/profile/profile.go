package profile

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
		ListSchools(ctx context.Context) ([]School, error)
		ListEmployers(ctx context.Context) ([]Employer, error)
	}
)

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

type (
	School struct {
		SchoolName string `json:"school_name" db:"school_name"`
		Type       string `json:"type" db:"type"`
	}

	ListSchoolsResponse struct {
		Schools []School `json:"schools"`
	}

	Employer struct {
		EmployerName string `json:"employer_name" db:"employer_name"`
	}

	ListEmployerResponse struct {
		Employers []Employer `json:"employers"`
	}
)

func (s *Service) ListSchools(ctx context.Context) (*ListSchoolsResponse, error) {
	schools, err := s.storage.ListSchools(ctx)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &ListSchoolsResponse{Schools: schools}, nil
}

func (s *Service) ListEmployers(ctx context.Context) (*ListEmployerResponse, error) {
	employers, err := s.storage.ListEmployers(ctx)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &ListEmployerResponse{Employers: employers}, nil
}
