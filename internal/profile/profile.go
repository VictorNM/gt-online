package profile

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/storage"
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
	GetProfileRequest struct {
		Email string `json:"email"`
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
		Email        string       `json:"email"`
		FirstName    string       `json:"first_name"`
		LastName     string       `json:"last_name"`
		Sex          string       `json:"sex,omitempty"`
		Birthdate    time.Time    `json:"birthdate,omitempty"`
		CurrentCity  string       `json:"current_city,omitempty"`
		Hometown     string       `json:"hometown,omitempty"`
		Interests    []string     `json:"interests,omitempty"`
		Education    []Attend     `json:"education,omitempty"`
		Professional []Employment `json:"professional,omitempty"`
	}
)

func (r Profile) MarshalJSON() ([]byte, error) {
	type alias Profile

	data := struct {
		alias
		Birthdate string `json:"birthdate,omitempty"`
	}{
		alias: alias(r),
	}
	if !r.Birthdate.IsZero() {
		data.Birthdate = r.Birthdate.Format(dateLayout)
	}

	return json.Marshal(data)
}

func (s *Service) GetProfile(ctx context.Context, req GetProfileRequest) (*Profile, error) {
	p, err := s.storage.GetProfile(ctx, req.Email)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, gterr.New(gterr.NotFound, "", err)
	}

	return p, nil
}

type UpdateProfileRequest struct {
	Email        string       `json:"-"`
	Sex          string       `json:"sex" binding:"oneof='' 'M' 'F'"`
	Birthdate    time.Time    `json:"birthdate"`
	CurrentCity  string       `json:"current_city"`
	Hometown     string       `json:"hometown"`
	Interests    []string     `json:"interests"`
	Education    []Attend     `json:"education"`
	Professional []Employment `json:"professional"`
}

func (r *UpdateProfileRequest) UnmarshalJSON(bytes []byte) (err error) {
	if r == nil {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(r)}
	}

	type alias UpdateProfileRequest
	var data struct {
		alias
		Birthdate string `json:"birthdate,omitempty"`
	}

	defer func() {
		if err == nil {
			*r = UpdateProfileRequest(data.alias)
		}
	}()

	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}

	if data.Birthdate == "" {
		return nil
	}

	d, err := time.Parse(dateLayout, data.Birthdate)
	if err != nil {
		return &json.UnmarshalTypeError{
			Value: data.Birthdate,
			Type:  reflect.TypeOf(r.Birthdate),
		}
	}

	data.alias.Birthdate = d
	return nil
}

func (s *Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
	// Validate interests
	im := make(map[string]struct{})
	for _, i := range req.Interests {
		if i == "" {
			return nil, gterr.New(gterr.InvalidArgument, "empty interest value")
		}

		im[i] = struct{}{}
	}
	if len(req.Interests) != len(im) {
		return nil, gterr.New(gterr.InvalidArgument, "duplicate interest value")
	}

	// Validate education
	for _, a := range req.Education {
		if a.School == "" {
			return nil, gterr.New(gterr.InvalidArgument, "empty school value")
		}

		if a.YearGraduated < 0 {
			return nil, gterr.New(gterr.InvalidArgument, "negative year_graduated")
		}
	}

	// Validate professional
	for _, e := range req.Professional {
		if e.Employer == "" {
			return nil, gterr.New(gterr.InvalidArgument, "empty employer value")
		}

		if e.JobTitle == "" {
			return nil, gterr.New(gterr.InvalidArgument, "empty job_title value")
		}
	}

	err := s.storage.UpdateProfile(ctx, req)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, gterr.New(gterr.NotFound, "", err)
	}

	if errors.Is(err, storage.ErrInvalidArgument) {
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
