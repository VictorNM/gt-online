package friend_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victornm/gtonline/internal/friend"
	"github.com/victornm/gtonline/internal/gterr"
	"github.com/victornm/gtonline/internal/storage/memory"
)

func TestService_CreateFriend(t *testing.T) {
	users := []memory.User{
		{
			Email: "foo@mock.com",
		},
		{
			Email: "bar@mock.com",
		},
	}

	t.Run("create friend with same email should failed", func(t *testing.T) {
		mock := memory.NewStorage()
		mock.InsertUsers(users)

		s := makeService(t, mock)
		err := s.CreateFriend(context.TODO(), friend.CreateFriendRequest{
			Email:       "foo@mock.com",
			FriendEmail: "foo@mock.com",
		})

		assert.Equal(t, gterr.InvalidArgument, gterr.Code(err))
	})

	t.Run("create friend with email that not an user should failed", func(t *testing.T) {
		mock := memory.NewStorage()
		mock.InsertUsers(users)

		s := makeService(t, mock)
		err := s.CreateFriend(context.TODO(), friend.CreateFriendRequest{
			Email:       "foo@mock.com",
			FriendEmail: "tony@stark.com",
		})

		assert.Equal(t, gterr.NotFound, gterr.Code(err))
	})

	t.Run("valid creation", func(t *testing.T) {
		mock := memory.NewStorage()
		mock.InsertUsers(users)

		s := makeService(t, mock)
		err := s.CreateFriend(context.TODO(), friend.CreateFriendRequest{
			Email:       "foo@mock.com",
			FriendEmail: "bar@mock.com",
		})

		assert.NoError(t, err)

		res, err := s.ListFriendRequests(context.TODO(), "foo@mock.com")
		require.NoError(t, err)
		require.Equal(t, "bar@mock.com", res.RequestTo[0].Email)

		res, err = s.ListFriendRequests(context.TODO(), "bar@mock.com")
		require.NoError(t, err)
		require.Equal(t, "foo@mock.com", res.RequestFrom[0].Email)
	})

	t.Run("double creation should success", func(t *testing.T) {
		mock := memory.NewStorage()
		mock.InsertUsers(users)

		s := makeService(t, mock)
		err := s.CreateFriend(context.TODO(), friend.CreateFriendRequest{
			Email:       "foo@mock.com",
			FriendEmail: "bar@mock.com",
		})
		require.NoError(t, err)

		err = s.CreateFriend(context.TODO(), friend.CreateFriendRequest{
			Email:        "foo@mock.com",
			FriendEmail:  "bar@mock.com",
			Relationship: "Co-worker",
		})
		require.NoError(t, err)

		res, err := s.ListFriendRequests(context.TODO(), "foo@mock.com")
		require.NoError(t, err)
		require.Equal(t, "Co-worker", res.RequestTo[0].Relationship)
	})

	t.Run("request already accepted should failed", func(t *testing.T) {
		mock := memory.NewStorage()
		mock.InsertUsers(users)
		err := mock.InsertFriendship(context.TODO(), &friend.Friendship{
			Email:         "foo@mock.com",
			FriendEmail:   "bar@mock.com",
			Relationship:  "",
			DateConnected: time.Now(),
		})
		require.NoError(t, err)

		s := makeService(t, mock)
		err = s.CreateFriend(context.TODO(), friend.CreateFriendRequest{
			Email:       "foo@mock.com",
			FriendEmail: "bar@mock.com",
		})
		require.Equal(t, gterr.AlreadyExists, gterr.Code(err))
	})
}

func makeService(_ *testing.T, s friend.Storage) *friend.Service {
	return friend.NewService(s)
}
