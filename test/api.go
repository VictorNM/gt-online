package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/google/go-querystring/query"
)

const (
	pathRegister  = "/auth/register"
	pathLogin     = "/auth/login"
	pathSchools   = "/schools"
	pathEmployers = "/employers"
	pathProfile   = "/users/profile"
	pathUsers     = "/users"
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

	UpdateProfileRequest struct {
		Sex          string       `json:"sex"`
		Birthdate    string       `json:"birthdate"`
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
		Email        string       `json:"email"`
		FirstName    string       `json:"first_name"`
		LastName     string       `json:"last_name"`
		Sex          string       `json:"sex"`
		Birthdate    string       `json:"birthdate"`
		CurrentCity  string       `json:"current_city"`
		Hometown     string       `json:"hometown"`
		Interests    []string     `json:"interests"`
		Education    []Attend     `json:"education"`
		Professional []Employment `json:"professional"`
	}

	ListUsersRequest struct {
		Email    string `url:"email,omitempty"`
		Name     string `url:"name,omitempty"`
		Hometown string `url:"hometown,omitempty"`
	}

	ListUsersResponse struct {
		Count int `json:"count"`
		Users []struct {
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Hometown  string `json:"hometown"`
		}
	}

	CreateFriendRequest struct {
		FriendEmail  string `json:"-"`
		Relationship string `json:"relationship"`
	}

	AcceptFriendRequest struct {
		FriendEmail string `json:"-"`
	}

	ListFriendRequestsResponse struct {
		RequestTo   []FriendRequest `json:"request_to"`
		RequestFrom []FriendRequest `json:"request_from"`
	}

	FriendRequest struct {
		Email        string `json:"email"`
		Relationship string `json:"relationship"`
	}

	ListFriendsResponse struct {
		Friends []Friendship
	}

	Friendship struct {
		FriendEmail   string `json:"friend_email"`
		Relationship  string `json:"relationship"`
		DateConnected string `json:"date_connected"`
	}

	DeleteFriendRequest struct {
		FriendEmail string
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

func (api *API) GetProfile(t *testing.T) (*Profile, error) {
	res := new(Profile)
	if err := api.send(t, http.MethodGet, pathProfile, nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) UpdateProfile(t *testing.T, req UpdateProfileRequest) (*Profile, error) {
	res := new(Profile)
	if err := api.send(t, http.MethodPut, pathProfile, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) ListUsers(t *testing.T, req ListUsersRequest) (*ListUsersResponse, error) {
	res := new(ListUsersResponse)
	if err := api.get(t, pathUsers, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) CreateFriendRequest(t *testing.T, req CreateFriendRequest) error {
	path := fmt.Sprintf("/friends/requests/%s", req.FriendEmail)
	return api.send(t, http.MethodPut, path, req, nil)
}

func (api *API) AcceptFriendRequest(t *testing.T, req AcceptFriendRequest) error {
	path := fmt.Sprintf("/friends/%s", req.FriendEmail)
	return api.send(t, http.MethodPut, path, req, nil)
}

func (api *API) ListFriendRequests(t *testing.T) (*ListFriendRequestsResponse, error) {
	res := new(ListFriendRequestsResponse)
	if err := api.get(t, "/friends/requests", nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) ListFriends(t *testing.T) (*ListFriendsResponse, error) {
	res := new(ListFriendsResponse)
	if err := api.get(t, "/friends", nil, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (api *API) CancelFriendRequest(t *testing.T, req DeleteFriendRequest) error {
	path := fmt.Sprintf("/friends/requests/%s", req.FriendEmail)
	return api.send(t, http.MethodDelete, path, req, nil)
}

func (api *API) RejectFriendRequest(t *testing.T, req DeleteFriendRequest) error {
	path := fmt.Sprintf("/friends/requests/%s?action=reject", req.FriendEmail)
	return api.send(t, http.MethodDelete, path, req, nil)
}

func (api *API) get(t *testing.T, path string, req interface{}, res interface{}) error {
	v, _ := query.Values(req)
	if q := v.Encode(); q != "" {
		path = path + "?" + q
	}
	return api.send(t, http.MethodGet, path, nil, res)
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
		t.Fatalf("Sending HTTP request: %v", err)
	}
	defer w.Body.Close()

	b, _ := io.ReadAll(w.Body)
	log.Println(string(b))

	if w.StatusCode < 200 || w.StatusCode > 299 {
		e := new(APIError)
		if err := json.Unmarshal(b, e); err != nil {
			t.Fatalf("Unmarshal APIError: %v, HTTP status=%d, body=%s ", err, w.StatusCode, string(b))
		}
		e.HTTPStatus = w.StatusCode
		return e
	}

	if out == nil {
		return nil
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
