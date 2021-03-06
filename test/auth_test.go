package test

import (
	"errors"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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
				assert.Equal(t, gterr.InvalidArgument, gterr.ErrorCode(e.Code))
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
		assert.Equal(t, gterr.AlreadyExists, gterr.ErrorCode(e.Code))
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
			assert.Equal(t, gterr.InvalidArgument, gterr.ErrorCode(e.Code))
		})
	}
}

// TestRegisterLogin test feature user login after register successfully
func TestRegisterLogin(t *testing.T) {
	api := makeAPI()

	// Given: a user register successfully
	email, password := mustRegister(t, api)

	// When: login with the registered email, password
	res, err := api.Login(t, LoginRequest{
		Email:    email,
		Password: password,
	})

	// Then: should not error, and return non-empty AccessToken and TokenType
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken, "access_token should not empty")
	assert.NotEmpty(t, res.TokenType, "token_type should not empty")
}

func TestAuthenticate(t *testing.T) {
	api := makeAPI()
	_, err := api.ListSchools(t)
	e := mustAPIErr(t, err)
	assert.Equal(t, http.StatusUnauthorized, e.HTTPStatus)
	assert.Equal(t, gterr.Unauthenticated, gterr.ErrorCode(e.Code))
}

type (
	register struct {
		req RegisterRequest
		res *RegisterResponse
	}

	listSchools struct {
		res *ListSchoolsResponse
	}

	listEmployers struct {
		res *ListEmployersResponse
	}

	updateProfile struct {
		req UpdateProfileRequest
		res *Profile
	}
)

func TestRegisterGetProfile(t *testing.T) {
	var (
		api      = makeAPI()
		register register
	)

	// Step 1: register new user
	{
		// Given: User with new email and correct information
		req := aValidRegisterRequest()

		// When: Register using this information
		res, err := api.Register(t, req)

		// Then: Should not error
		require.NoError(t, err)
		register.req, register.res = req, res
	}

	// Use token from step 1 for next steps
	api.WithToken(Token{AccessToken: register.res.AccessToken, TokenType: register.res.TokenType})

	// Step 2: user get profile without EditProfile
	{
		res, err := api.GetProfile(t)
		require.NoError(t, err)
		require.Equal(t, register.req.Email, res.Email)
		require.Equal(t, register.req.FirstName, res.FirstName)
		require.Equal(t, register.req.LastName, res.LastName)
	}
}

// TestRegisterEditProfile test feature new user register, then navigate to EditProfile and submit form.
func TestRegisterEditProfile(t *testing.T) {
	var (
		api = makeAPI()

		register      register
		listSchools   listSchools
		listEmployers listEmployers
		updateProfile updateProfile
	)

	// Step 1: register new user
	{
		// Given: User with new email and correct information
		req := aValidRegisterRequest()

		// When: Register using this information
		res, err := api.Register(t, req)

		// Then: Should not error
		require.NoError(t, err)
		register.req, register.res = req, res
	}

	// Use token from step 1 for next steps
	api.WithToken(Token{AccessToken: register.res.AccessToken, TokenType: register.res.TokenType})

	// Step 2.1: list available schools for user to EditProfile
	{
		// When: Call list schools
		res, err := api.ListSchools(t)

		// Then: Should not error, and return non-empty schools list
		require.NoError(t, err)
		require.NotEmpty(t, res.Schools)

		listSchools.res = res
	}

	// Step 2.2: list available employers for user to EditProfile
	{
		// When: Call list schools
		res, err := api.ListEmployers(t)

		// Then: Should not error, and return non-empty schools list
		require.NoError(t, err)
		require.NotEmpty(t, res.Employers)

		listEmployers.res = res
	}

	// Step 3: user fill and submit EditProfile form
	{
		req := UpdateProfileRequest{
			Sex:         "M",
			Birthdate:   "29/05/1970",
			CurrentCity: "New York",
			Hometown:    "New York",
			Interests:   []string{"Technology"},
			Education: []Attend{{
				School:        listSchools.res.Schools[0].SchoolName,
				YearGraduated: 1980,
			}},
			Professional: []Employment{{
				Employer: listEmployers.res.Employers[0].EmployerName,
				JobTitle: "CEO",
			}},
		}
		res, err := api.UpdateProfile(t, req)

		// Then: Should not error, and return an updated profile
		require.NoError(t, err, "update profile failed")
		require.Equal(t, res, &Profile{
			Email:        register.req.Email,
			FirstName:    register.req.FirstName,
			LastName:     register.req.LastName,
			Sex:          req.Sex,
			Birthdate:    req.Birthdate,
			CurrentCity:  req.CurrentCity,
			Hometown:     req.Hometown,
			Interests:    req.Interests,
			Education:    req.Education,
			Professional: req.Professional,
		}, "the returned profile after updated is not match expectation")

		updateProfile.req, updateProfile.res = req, res
	}

	// Step 4: user query the profile again
	{
		res, err := api.GetProfile(t)
		// Then: Should not error, and return an updated profile
		require.NoError(t, err, "update profile failed")
		require.Equal(t, updateProfile.res, res, "the returned profile is not match with previous response of UpdateProfile")
	}
}

func TestValidateEditProfile(t *testing.T) {
	api := makeRegisteredAPI(t, aValidRegisterRequest())

	tests := map[string]struct {
		alter func(req *UpdateProfileRequest)
		valid bool
	}{
		"sex = 'M' should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Sex = "M"
			},
			valid: true,
		},

		"sex = 'F' should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Sex = "F"
			},
			valid: true,
		},

		"sex = '' should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Sex = ""
			},
			valid: true,
		},

		"sex != '', 'M', 'F' should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Sex = "K"
			},
			valid: false,
		},

		"null interests should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Interests = nil
			},
			valid: true,
		},

		"empty slice interests should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Interests = []string{}
			},
			valid: true,
		},

		"multi non-empty interest value should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Interests = []string{"Soccer", "Books"}
			},
			valid: true,
		},

		"empty interest value should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Interests = []string{""}
			},
			valid: false,
		},

		"duplicate interest value should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Interests = []string{"Books", "Books"}
			},
			valid: false,
		},

		"empty year_graduate should valid": {
			alter: func(req *UpdateProfileRequest) {
				req.Education = []Attend{
					{
						School:        "University of Oxford",
						YearGraduated: 0,
					},
				}
			},
			valid: true,
		},

		"negative year_graduate should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Education = []Attend{
					{
						School:        "University of Oxford",
						YearGraduated: -1,
					},
				}
			},
			valid: false,
		},

		"empty school should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Education = []Attend{
					{
						School: "",
					},
				}
			},
			valid: false,
		},

		"empty employer should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Professional = []Employment{
					{
						Employer: "",
						JobTitle: "CTO",
					},
				}
			},
			valid: false,
		},

		"empty job_title should invalid": {
			alter: func(req *UpdateProfileRequest) {
				req.Professional = []Employment{
					{
						Employer: "Microsoft",
						JobTitle: "",
					},
				}
			},
			valid: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := UpdateProfileRequest{}
			test.alter(&req)
			_, err := api.UpdateProfile(t, req)
			if test.valid {
				require.NoError(t, err)
				return
			}

			require.Error(t, err, "should return an error")
			e := mustAPIErr(t, err)
			require.EqualValues(t, gterr.InvalidArgument, gterr.ErrorCode(e.Code))
		})
	}
}

func TestAPI_SearchUsers(t *testing.T) {
	type profile struct {
		Email     string
		FirstName string
		LastName  string
		Hometown  string
	}

	// Given: a list of users exist in the system
	for _, req := range []profile{
		{
			Email:     faker.Email(),
			FirstName: "Tony",
			LastName:  "Stark",
			Hometown:  "New York",
		},
		{
			Email:     faker.Email(),
			FirstName: "Steve",
			LastName:  "Rogers",
			Hometown:  "New York",
		},
		{
			Email:     faker.Email(),
			FirstName: "Bruce",
			LastName:  "Banner",
			Hometown:  "Ohio",
		},
		{
			Email:     faker.Email(),
			FirstName: "Bruce",
			LastName:  "Wayne",
			Hometown:  "Gotham",
		},
		{
			Email:     faker.Email(),
			FirstName: "Clark",
			LastName:  "Kent",
			Hometown:  "Metropolis",
		},
	} {
		pass := faker.Password()
		createUser(t,
			RegisterRequest{
				Email:                req.Email,
				Password:             pass,
				PasswordConfirmation: pass,
				FirstName:            req.FirstName,
				LastName:             req.LastName,
			},
			UpdateProfileRequest{
				Hometown: req.Hometown,
			})
	}

	// When: the current user search for friends
	api := makeRegisteredAPI(t, aValidRegisterRequest())
	res, err := api.ListUsers(t, ListUsersRequest{
		Email:    "",
		Name:     "Tony",
		Hometown: "Metropolis",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Users)
	assert.Equal(t, res.Count, len(res.Users))
}

func TestFriendship_RequestAndAcceptFriend(t *testing.T) {
	// Given: 2 users exist in the system
	user, friend := aValidRegisterRequest(), aValidRegisterRequest()
	userAPI, friendAPI := makeRegisteredAPI(t, user), makeRegisteredAPI(t, friend)

	// Step 1: user send a friend request
	{
		// When: user send a friend request
		err := userAPI.CreateFriendRequest(t, CreateFriendRequest{
			FriendEmail:  friend.Email,
			Relationship: "Co-worker",
		})

		// Then: it should be success
		require.NoError(t, err)

		// When: user list pending requests
		res, err := userAPI.ListFriendRequests(t)

		// Then: the new request should be in the response
		require.NoError(t, err)
		require.Contains(t, res.RequestTo, FriendRequest{
			Email:        friend.Email,
			Relationship: "Co-worker",
		})
	}

	// Step 2: friend see pending request and accept it
	{
		// When: friend list pending requests
		res, err := friendAPI.ListFriendRequests(t)

		// Then: should see request from user
		require.NoError(t, err)
		require.Contains(t, res.RequestFrom, FriendRequest{
			Email:        user.Email,
			Relationship: "Co-worker",
		})

		// When: friend accept the request
		err = friendAPI.AcceptFriendRequest(t, AcceptFriendRequest{
			FriendEmail: res.RequestFrom[0].Email,
		})

		// Then: should be success
		require.NoError(t, err)

		// When: friend list pending requests again
		res, err = friendAPI.ListFriendRequests(t)

		// Then: should not see request from user
		require.NoError(t, err)
		require.NotContains(t, res.RequestFrom, FriendRequest{
			Email:        user.Email,
			Relationship: "Co-worker",
		})
	}

	// Step 3: user see friend in the friend list
	{
		// When: user list pending requests again
		res, err := userAPI.ListFriendRequests(t)

		// Then: the accepted request should be not in the list
		require.NoError(t, err)
		require.NotContains(t, res.RequestTo, FriendRequest{
			Email: friend.Email,
		})

		// When: user list friends
		friends, err := userAPI.ListFriends(t)

		// Then: should see the new friend in the list
		require.NoError(t, err)
		require.Contains(t, friends.Friends, Friendship{
			FriendEmail:   friend.Email,
			Relationship:  "Co-worker",
			DateConnected: time.Now().Format("January 02, 2006"),
		})
	}
}

func TestFriendship_RequestAndCancelFriend(t *testing.T) {
	// Given: 2 users exist in the system
	user, friend := aValidRegisterRequest(), aValidRegisterRequest()
	userAPI, _ := makeRegisteredAPI(t, user), makeRegisteredAPI(t, friend)

	// When: user send a friend request
	err := userAPI.CreateFriendRequest(t, CreateFriendRequest{
		FriendEmail:  friend.Email,
		Relationship: "Co-worker",
	})

	// Then: it should be success
	require.NoError(t, err)

	// When: user list pending requests
	res, err := userAPI.ListFriendRequests(t)

	// Then: the new request should be in the response
	require.NoError(t, err)
	require.Contains(t, res.RequestTo, FriendRequest{
		Email:        friend.Email,
		Relationship: "Co-worker",
	})

	// When: user cancel the friend request
	err = userAPI.CancelFriendRequest(t, DeleteFriendRequest{
		FriendEmail: friend.Email,
	})

	// Then: it should be success
	require.NoError(t, err)

	// When: user list pending requests again
	res, err = userAPI.ListFriendRequests(t)

	// Then: the new request should be in the response
	require.NoError(t, err)
	require.NotContains(t, res.RequestTo, FriendRequest{
		Email:        friend.Email,
		Relationship: "Co-worker",
	})
}

func TestFriendship_RequestAndRejectFriend(t *testing.T) {
	// Given: 2 users exist in the system
	user, friend := aValidRegisterRequest(), aValidRegisterRequest()
	userAPI, friendAPI := makeRegisteredAPI(t, user), makeRegisteredAPI(t, friend)

	// When: user send a friend request
	err := userAPI.CreateFriendRequest(t, CreateFriendRequest{
		FriendEmail:  friend.Email,
		Relationship: "Co-worker",
	})

	// Then: it should be success
	require.NoError(t, err)

	// When: friend list pending requests
	res, err := friendAPI.ListFriendRequests(t)

	// Then: the new request should be in the response
	require.NoError(t, err)
	require.Contains(t, res.RequestFrom, FriendRequest{
		Email:        user.Email,
		Relationship: "Co-worker",
	})

	// When: friend reject the friend request
	err = friendAPI.RejectFriendRequest(t, DeleteFriendRequest{
		FriendEmail: res.RequestFrom[0].Email,
	})

	// Then: it should be success
	require.NoError(t, err)

	// When: friend list pending requests again
	res, err = friendAPI.ListFriendRequests(t)

	// Then: the new request should be in the response
	require.NoError(t, err)
	require.NotContains(t, res.RequestFrom, FriendRequest{
		Email:        user.Email,
		Relationship: "Co-worker",
	})
}

func aValidRegisterRequest() RegisterRequest {
	req := RegisterRequest{
		Email:     faker.Email(),
		Password:  faker.Password(),
		FirstName: faker.FirstName(),
		LastName:  faker.LastName(),
	}
	req.PasswordConfirmation = req.Password
	return req
}

func mustRegister(t *testing.T, api *API) (email string, password string) {
	req := aValidRegisterRequest()
	_, err := api.Register(t, req)
	require.NoError(t, err, "register failed")
	return req.Email, req.Password
}

func makeRegisteredAPI(t *testing.T, req RegisterRequest) *API {
	api := makeAPI()
	res, err := api.Register(t, req)
	require.NoError(t, err, "register failed")
	api.WithToken(Token{
		AccessToken: res.AccessToken,
		TokenType:   res.TokenType,
	})
	return api
}

func createUser(t *testing.T, reg RegisterRequest, up UpdateProfileRequest) {
	api := makeRegisteredAPI(t, reg)
	_, err := api.UpdateProfile(t, up)
	require.NoError(t, err, "update profile failed")
}

func makeAPI() *API {
	return &API{addr: addr}
}

func mustAPIErr(t *testing.T, err error) *APIError {
	t.Helper()

	e := new(APIError)
	if errors.As(err, &e) {
		return e
	}
	t.Fatalf("The provided error is not an %T: %v", e, err)
	return nil
}

var (
	env  string
	addr string
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
	addr = "http://localhost:8080" // define in docker-compose.yaml
	if env == "local" {
		srv := httptest.NewServer(server.New(testConfig()))
		defer srv.Close()
		addr = srv.URL
	}

	return m.Run()
}

func testConfig() server.Config {
	c := server.Config{}
	// App Config
	c.App.Addr = ":8080"
	c.Auth.Secret = "JznqcOJCAEc1aq7Zulm83OtQt7md2gOK"

	// DB config
	c.DB.Addr = "localhost:3306"
	c.DB.User = "root"
	c.DB.Pass = "root"
	c.DB.Name = "gt-online"
	return c
}
