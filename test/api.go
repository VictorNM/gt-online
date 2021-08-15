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
	pathRegister = "/auth/register"
	pathLogin    = "/auth/login"
)

type (
	API struct {
		addr string
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
	}

	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	LoginResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
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

func (api *API) send(t *testing.T, method string, path string, in interface{}, out interface{}) error {
	t.Helper()

	data, _ := json.Marshal(in)
	u := api.addr + path

	r, err := http.NewRequest(method, u, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Building HTTP request: %v", err)
	}
	r.Header.Set("Content-Type", "application/json")

	t.Logf("Sending request: %s %s %s", r.Method, r.URL, string(data))
	w, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatalf("Sendind HTTP request: %v", err)
	}
	defer w.Body.Close()

	b, _ := io.ReadAll(w.Body)

	t.Logf("Received response: HTTP status=%d body=%s", w.StatusCode, string(b))

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
