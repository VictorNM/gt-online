//go:build integration
// +build integration

package storage_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victornm/gtonline/internal/auth"
	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/profile"
	"github.com/victornm/gtonline/internal/server"
	"github.com/victornm/gtonline/internal/storage"
)

func TestStorage_ListSchools(t *testing.T) {
	s := makeStorage(t)

	schools, err := s.ListSchools(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, schools)
}

func TestStorage_ListEmployers(t *testing.T) {
	s := makeStorage(t)

	employers, err := s.ListEmployers(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, employers)
}

func TestUpdateProfile(t *testing.T) {
	s := makeStorage(t)

	ctx := context.Background()
	email := "foo@bar.com"

	// Create new user
	err := s.CreateUser(ctx, auth.User{
		Email:          email,
		HashedPassword: "123",
		FirstName:      "foo",
		LastName:       "bar",
	})
	require.NoError(t, err, "create user failed")
	t.Cleanup(func() {
		if err := s.DeleteUser(ctx, email); err != nil {
			t.Errorf("delete user failed: %v", err)
		}
	})

	// Get profile after create user
	p, err := s.GetProfile(ctx, email)
	require.NoError(t, err)
	require.Equal(t, "foo", p.FirstName)

	// Update profile 1st time:
	req := profile.UpdateProfileRequest{
		Email:        email,
		Sex:          "M",
		Birthdate:    time.Now(),
		CurrentCity:  "FooCity",
		Hometown:     "BarCity",
		Interests:    []string{"Books"},
		Education:    []profile.Attend{{School: "University of Oxford", YearGraduated: 2021}},
		Professional: []profile.Employment{{Employer: "Microsoft", JobTitle: "CEO"}},
	}
	err = s.UpdateProfile(ctx, req)
	require.NoError(t, err)
	p, err = s.GetProfile(ctx, email)
	require.NoError(t, err)
	require.Equal(t, "M", p.Sex)
	require.Equal(t, []string{"Books"}, p.Interests)
	require.Equal(t, []profile.Attend{{School: "University of Oxford", YearGraduated: 2021}}, p.Education)
	require.Equal(t, []profile.Employment{{Employer: "Microsoft", JobTitle: "CEO"}}, p.Professional)

	// Update profile 2nd time:
	req.Sex = "F"
	req.Interests = []string{"Technology"}
	req.Education = nil
	req.Professional = append(req.Professional, profile.Employment{Employer: "Apple", JobTitle: "CTO"})
	err = s.UpdateProfile(ctx, req)
	require.NoError(t, err)
	p, err = s.GetProfile(ctx, email)
	require.NoError(t, err)
	require.Equal(t, "F", p.Sex)
	require.Equal(t, []string{"Technology"}, p.Interests)
	require.Empty(t, p.Education)
	require.ElementsMatch(t, []profile.Employment{
		{Employer: "Microsoft", JobTitle: "CEO"},
		{Employer: "Apple", JobTitle: "CTO"},
	}, p.Professional)
}

func TestUpdateProfileInvalidEmployer(t *testing.T) {
	s := makeStorage(t)

	ctx := context.Background()
	email := "foo@bar.com"

	// Create new user
	err := s.CreateUser(ctx, auth.User{
		Email:          email,
		HashedPassword: "123",
		FirstName:      "foo",
		LastName:       "bar",
	})
	require.NoError(t, err, "create user failed")
	t.Cleanup(func() {
		if err := s.DeleteUser(ctx, email); err != nil {
			t.Errorf("delete user failed: %v", err)
		}
	})

	// Update profile 1st time:
	req := profile.UpdateProfileRequest{
		Email:        email,
		Sex:          "M",
		Birthdate:    time.Now(),
		CurrentCity:  "FooCity",
		Hometown:     "BarCity",
		Professional: []profile.Employment{{Employer: "Tiki", JobTitle: "CEO"}},
	}
	err = s.UpdateProfile(ctx, req)
	require.Error(t, err)

	assert.True(t, errors.Is(err, gterr.ErrInvalidArgument), err)
}

func makeStorage(t *testing.T) *storage.Storage {
	once.Do(func() {
		var err error
		cfg := server.DefaultConfig().DB
		cfg.Addr = "localhost:3306"
		s, err = storage.New(cfg)
		require.NoError(t, err)
		require.NoError(t, s.Ping())
	})
	return s
}

var (
	once sync.Once
	s    *storage.Storage
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	defer func() {
		if s != nil {
			_ = s.Close()
		}
	}()
	return m.Run()
}

func TestCast(t *testing.T) {
	var i interface{}

	j, ok := i.(string)
	t.Log(ok)
	t.Log(j)
}
