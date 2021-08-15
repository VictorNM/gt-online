package test

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/server"
)

func TestRegister_Validation(t *testing.T) {
	validReq := func() RegisterRequest {
		req := RegisterRequest{
			Email:                faker.Email(),
			Password:             "Abc@123_xyZ",
			PasswordConfirmation: "Abc@123_xyZ",
			FirstName:            "Hoa",
			LastName:             "Binh",
		}
		return req
	}

	t.Run("validate request format", func(t *testing.T) {
		tests := map[string]struct {
			alterReq func(req *RegisterRequest)
		}{
			"invalid email": {
				alterReq: func(req *RegisterRequest) {
					req.Email = "invalid"
				},
			},

			"password != confirmation password": {
				alterReq: func(req *RegisterRequest) {
					req.Password = "a-password-123"
					req.PasswordConfirmation = "another-password-456"
				},
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				api := makeAPI()
				req := validReq()
				test.alterReq(&req)

				_, err := api.Register(t, req)
				e := mustAPIErr(t, err)
				assert.Equal(t, http.StatusBadRequest, e.HTTPStatus)
				assert.Equal(t, gterr.InvalidArgument, e.Code)
			})
		}
	})

	t.Run("duplicate email can't register", func(t *testing.T) {
		api := makeAPI()
		req := validReq()

		res, err := api.Register(t, req)
		require.NoError(t, err)
		assert.Equal(t, req.Email, res.Email)

		_, err = api.Register(t, req)
		e := mustAPIErr(t, err)
		assert.Equal(t, http.StatusConflict, e.HTTPStatus)
		assert.Equal(t, gterr.AlreadyExists, e.Code)
	})
}

func TestLogin_Validation(t *testing.T) {
	validReq := func() LoginRequest {
		req := LoginRequest{
			Email:    faker.Email(),
			Password: "Abc@123_xyZ",
		}
		return req
	}

	tests := map[string]struct {
		alterReq func(req *LoginRequest)
	}{
		"invalid email": {
			alterReq: func(req *LoginRequest) {
				req.Email = "invalid"
			},
		},

		"empty password": {
			alterReq: func(req *LoginRequest) {
				req.Password = ""
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			api := makeAPI()
			req := validReq()
			test.alterReq(&req)

			_, err := api.Login(t, req)
			e := mustAPIErr(t, err)
			assert.Equal(t, http.StatusBadRequest, e.HTTPStatus)
			assert.Equal(t, gterr.InvalidArgument, e.Code)
		})
	}
}

func TestRegisterLogin(t *testing.T) {
	api := makeAPI()
	req := RegisterRequest{
		Email:                faker.Email(),
		Password:             "Abc@123_xyZ",
		PasswordConfirmation: "Abc@123_xyZ",
		FirstName:            "Hoa",
		LastName:             "Binh",
	}
	res, err := api.Register(t, req)
	require.NoError(t, err)
	assert.Equal(t, req.Email, res.Email)

	res2, err := api.Login(t, LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res2.AccessToken, "access_token should not empty")
	assert.True(t, strings.EqualFold("bearer", res2.TokenType), "token_type should be 'bearer'")
}

func makeAPI() *API {
	return api
}

func mustAPIErr(t *testing.T, err error) *APIError {
	t.Helper()

	e := new(APIError)
	if errors.As(err, &e) {
		return e
	}
	t.Fatalf("The provided error is not an %T", e)
	return nil
}

var (
	env string
	api *API
)

func init() {
	flag.StringVar(&env, "env", "local", "the environment to call API: local or docker")
}

func TestMain(m *testing.M) {
	flag.Parse()
	gin.SetMode(gin.TestMode)

	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	var addr = "http://localhost:8080"
	if env == "local" {
		cfg := server.DefaultConfig()
		cfg.DB.Addr = "localhost:3306"
		s := server.New(cfg)
		srv := httptest.NewServer(s)
		defer srv.Close()
		addr = srv.URL
	}

	api = &API{addr: addr}
	return m.Run()
}
