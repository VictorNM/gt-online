package profile

import (
	"context"
	"errors"
	"time"

	"github.com/victornm/gtonline/internal/gterr"
)

const dateLayout = "02/01/2006"

type (
	Service struct {
		storage Storage
	}

	Storage interface {
		GetProfile(ctx context.Context, email string) (*Profile, error)
		UpdateProfile(ctx context.Context, req UpdateProfileRequest) (err error)
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

	Employer struct {
		EmployerName string `json:"employer_name" db:"employer_name"`
	}
)

type (
	Date struct {
		time.Time
	}

	GetProfileRequest struct {
		Email string `json:"email"`
	}

	UpdateProfileRequest struct {
		Email        string       `json:"email"`
		Sex          string       `json:"sex"`
		Birthdate    Date         `json:"birthdate"`
		CurrentCity  string       `json:"current_city"`
		Hometown     string       `json:"hometown"`
		Interests    []string     `json:"interests"`
		Education    []Attend     `json:"education"`
		Professional []Employment `json:"professional"`
	}

	Attend struct {
		School        string `json:"school"`
		YearGraduated int    `json:"year_graduated"`
	}

	Employment struct {
		Employer string `json:"employer"`
		JobTitle string `json:"job_title"`
	}

	Profile struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		UpdateProfileRequest
	}
)

func (d Date) MarshalJSON() ([]byte, error) {
	if y := d.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Date.MarshalJSON: year outside of range [0,9999]")
	}

	b := make([]byte, 0, len(dateLayout)+2)
	b = append(b, '"')
	b = d.AppendFormat(b, dateLayout)
	b = append(b, '"')
	return b, nil
}

func (d *Date) UnmarshalJSON(bytes []byte) error {
	// Ignore null, like in the main JSON package.
	if string(bytes) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	d.Time, err = time.Parse(`"`+dateLayout+`"`, string(bytes))
	return err
}

func (s *Service) GetProfile(ctx context.Context, req GetProfileRequest) (*Profile, error) {
	p, err := s.storage.GetProfile(ctx, req.Email)
	if errors.Is(err, gterr.ErrNotFound) {
		return nil, gterr.New(gterr.NotFound, "", err)
	}

	return p, nil
}

func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
	err := s.storage.UpdateProfile(ctx, req)
	if errors.Is(err, gterr.ErrNotFound) {
		return nil, gterr.New(gterr.NotFound, "", err)
	}

	if errors.Is(err, gterr.ErrInvalidArgument) {
		return nil, gterr.New(gterr.InvalidArgument, "", err)
	}

	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	p, err := s.storage.GetProfile(ctx, req.Email)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return p, nil
}

type ListSchoolsResponse struct {
	Schools []School `json:"schools"`
}

func (s *Service) ListSchools(ctx context.Context) (*ListSchoolsResponse, error) {
	schools, err := s.storage.ListSchools(ctx)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &ListSchoolsResponse{Schools: schools}, nil
}

type ListEmployerResponse struct {
	Employers []Employer `json:"employers"`
}

func (s *Service) ListEmployers(ctx context.Context) (*ListEmployerResponse, error) {
	employers, err := s.storage.ListEmployers(ctx)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &ListEmployerResponse{Employers: employers}, nil
}
