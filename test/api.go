package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const (
	pathRegister  = "/auth/register"
	pathLogin     = "/auth/login"
	pathSchools   = "/schools"
	pathEmployers = "/employers"
)

type (
	API struct {
		addr string

		token Token
	}

	RegisterRequest struct {
		Email                string `json:"email"`
		Password             string `json:"password"`
		PasswordConfirmation string `json:"password_confirmation"`
		FirstName            string `json:"first_name"`
		LastName             string `json:"last_name"`
	}

	RegisterResponse struct {
		Email string `json:"email"`
		Token
	}

	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	LoginResponse struct {
		Token
	}

	ListSchoolsResponse struct {
		Schools []struct {
			SchoolName string `json:"school_name"`
			Type       string `json:"type"`
		} `json:"schools"`
	}

	ListEmployersResponse struct {
		Employers []struct {
			EmployerName string `json:"employer_name"`
		} `json:"employers"`
	}
)

func (api *API) Register(t *testing.T, req RegisterRequest) (*RegisterResponse, error) {
	res := new(RegisterResponse)
	if err := api.send(t, http.MethodPost, pathRegister, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) Login(t *testing.T, req LoginRequest) (*LoginResponse, error) {
	res := new(LoginResponse)
	if err := api.send(t, http.MethodPost, pathLogin, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) ListSchools(t *testing.T) (*ListSchoolsResponse, error) {
	res := new(ListSchoolsResponse)
	if err := api.send(t, http.MethodGet, pathSchools, nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) ListEmployers(t *testing.T) (*ListEmployersResponse, error) {
	res := new(ListEmployersResponse)
	if err := api.send(t, http.MethodGet, pathEmployers, nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (api *API) WithToken(token Token) *API {
	api.token = token
	return api
}

func (api *API) send(t *testing.T, method string, path string, in interface{}, out interface{}) error {
	t.Helper()

	var reader io.Reader

	if in != nil {
		data, _ := json.Marshal(in)
		reader = bytes.NewReader(data)
	}

	u := api.addr + path

	r, err := http.NewRequest(method, u, reader)
	if err != nil {
		t.Fatalf("Building HTTP request: %v", err)
	}
	r.Header.Set("Content-Type", "application/json")

	if tk := api.token.TokenType + " " + api.token.AccessToken; tk != "" {
		r.Header.Set("Authorization", tk)
	}

	w, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("Sendind HTTP request: %v", err)
	}
	defer w.Body.Close()

	b, _ := io.ReadAll(w.Body)

	if w.StatusCode < 200 || w.StatusCode > 299 {
		e := new(APIError)
		if err := json.Unmarshal(b, e); err != nil {
			t.Fatalf("Unmarshal APIError: %v, HTTP status=%d, body=%s ", err, w.StatusCode, string(b))
		}
		e.HTTPStatus = w.StatusCode
		return e
	}

	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("Unmarshal body: %v, body=%s ", err, string(b))
	}
	return nil
}

type APIError struct {
	HTTPStatus int
	Code       string
	Message    string
}

func (err *APIError) Error() string {
	return fmt.Sprintf("HTTP status=%d, code=%s, message=%s", err.HTTPStatus, err.Code, err.Message)
}
