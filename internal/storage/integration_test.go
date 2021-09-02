//go:build integration
// +build integration

package storage_test

import (
	"context"
	"os"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"

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
