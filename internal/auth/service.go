package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"github.com/victornm/gtonline/internal/gterr"
)

type (
	Service struct {
		storage Storage
		secret  []byte
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
)

func NewService(storage Storage, secret []byte) *Service {
	return &Service{
		storage: storage,
		secret:  secret,
	}
}

type (
	RegisterRequest struct {
		Email                string `json:"email" binding:"email,required"`
		Password             string `json:"password" binding:"required"`
		PasswordConfirmation string `json:"password_confirmation" binding:"eqfield=Password"`
		FirstName            string `json:"first_name" binding:"required"`
		LastName             string `json:"last_name" binding:"required"`
	}

	RegisterResponse struct {
		Email string `json:"email"`
		Token
	}
)

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	_, err := s.storage.FindUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, gterr.New(gterr.AlreadyExists, fmt.Sprintf("Email %s already registered.", req.Email))
	}

	if err != nil && err != gterr.ErrNotFound {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	hashed, err := hash(req.Password)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	u := User{
		Email:          req.Email,
		HashedPassword: hashed,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
	}
	err = s.storage.CreateUser(ctx, u)

	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	token, err := genToken(u, s.secret)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &RegisterResponse{
		Email: req.Email,
		Token: newBearerToken(token),
	}, nil
}

type (
	LoginRequest struct {
		Email    string `json:"email" binding:"email,required"`
		Password string `json:"password" binding:"required"`
	}

	// LoginResponse follow the convention described here: https://www.oauth.com/oauth2-servers/access-tokens/access-token-response/
	LoginResponse struct {
		Token
	}
)

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	u, err := s.storage.FindUserByEmail(ctx, req.Email)
	if errors.Is(err, gterr.ErrNotFound) {
		return nil, gterr.New(gterr.Unauthenticated, "Email or password do not matched.", err)
	}

	match, err := compareHash(u.HashedPassword, req.Password)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	if !match {
		return nil, gterr.New(gterr.Unauthenticated, "Email or password do not matched.", err)
	}

	token, err := genToken(*u, s.secret)
	if err != nil {
		return nil, gterr.New(gterr.Internal, "", err)
	}

	return &LoginResponse{
		Token: newBearerToken(token),
	}, nil
}

type (
	Token struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
)

func newBearerToken(token string) Token {
	return Token{
		AccessToken: token,
		TokenType:   "Bearer",
	}
}

func (s *Service) Authenticate(_ context.Context, req Token) (*UserAuthDTO, error) {
	if !strings.EqualFold(req.TokenType, "bearer") {
		return nil, gterr.New(gterr.Unauthenticated, "Invalid access token", fmt.Errorf("token type not supported: %v", req.TokenType))
	}

	u, valid, err := parseToken(req.AccessToken, s.secret)
	if err != nil {
		return nil, gterr.New(gterr.Unauthenticated, "Invalid access token", err)
	}

	if !valid {
		return nil, gterr.New(gterr.Unauthenticated, "Invalid access token")
	}

	return u, nil
}

func hash(pass string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func compareHash(hashed, pass string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pass))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func genToken(u User, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "gt-online/auth",
		},
		UserAuthDTO: &UserAuthDTO{Email: u.Email},
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %v", err)
	}

	return tokenString, nil
}

func parseToken(tokenString string, secret []byte) (*UserAuthDTO, bool, error) {
	var claims jwtClaims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (i interface{}, err error) {
		return secret, nil
	})

	if err != nil {
		return nil, false, fmt.Errorf("parse token: %v", err)
	}

	if !token.Valid {
		return nil, false, nil
	}

	return claims.UserAuthDTO, true, nil
}

type jwtClaims struct {
	jwt.StandardClaims
	*UserAuthDTO
}

type UserAuthDTO struct {
	Email string `json:"email"`
}
