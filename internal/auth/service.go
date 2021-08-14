package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/victornm/gtonline/internal/gterr"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type (
	Service struct {
		storage Storage
	}

	Storage interface {
		FindUserByEmail(ctx context.Context, email string) (*User, error)
		CreateUser(ctx context.Context, u User) error
	}

	User struct {
		Email          string `db:"email"`
		HashedPassword string `db:"password"`
		FirstName      string `db:"first_name"`
		LastName       string `db:"last_name"`
	}

	RegisterRequest struct {
		Email                string `json:"email" binding:"email"`
		Password             string `json:"password" binding:"required"`
		PasswordConfirmation string `json:"password_confirmation" binding:"eqfield=Password"`
		FirstName            string `json:"first_name" binding:"required"`
		LastName             string `json:"last_name" binding:"required"`
	}

	RegisterResponse struct {
		Email string `json:"email"`
	}
)

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	_, err := s.storage.FindUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, gterr.New(gterr.AlreadyExists, fmt.Sprintf("Email %s already registered.", req.Email))
	}

	if err != nil && err != ErrNotFound {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	err = s.storage.CreateUser(ctx, User{
		Email:          req.Email,
		HashedPassword: string(hashed),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
	})

	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &RegisterResponse{
		Email: req.Email,
	}, nil
}
