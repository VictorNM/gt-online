package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victornm/gtonline/internal/config"
)

type Config struct {
	Mode string
	App  struct {
		Name string
		Port int64
	}
	Cache struct {
		Enable        bool
		ExpiredMinute int64 `mapstructure:"expired_minute"`
	}
}

var (
	defaultConfig = func() Config {
		var c Config
		c.Mode = "local"
		c.App.Name = "gt-online"
		c.App.Port = 8000
		c.Cache.Enable = false
		c.Cache.ExpiredMinute = 0
		return c
	}

	configFile = "./testdata/full.yaml"
	// yamlFullConfig return the config associate with the configFile
	yamlFullConfig = func() Config {
		var c Config
		c.Mode = "sandbox"
		c.App.Name = "gt-online"
		c.App.Port = 8080
		c.Cache.Enable = true
		c.Cache.ExpiredMinute = 10
		return c
	}
)

func TestLoadDefaultConfig(t *testing.T) {
	c := defaultConfig()
	err := config.Load("./abc.yaml", &c)
	require.NoError(t, err)

	assert.Equal(t, defaultConfig(), c)
}

func TestLoadFullFromYaml(t *testing.T) {
	var c Config
	err := config.Load("./testdata/full.yaml", &c)
	require.NoError(t, err)

	assert.Equal(t, yamlFullConfig(), c)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("APP_NAME", "refund")

	var c Config
	err := config.Load(configFile, &c)
	require.NoError(t, err)

	want := yamlFullConfig()
	want.App.Name = "refund"
	assert.Equal(t, want, c)
}

func TestLoadConfigHasUnderscoreFromEnv(t *testing.T) {
	t.Setenv("CACHE_EXPIRED_MINUTE", "15")

	var c Config
	err := config.Load(configFile, &c)
	require.NoError(t, err)

	want := yamlFullConfig()
	want.Cache.ExpiredMinute = 15
	assert.Equal(t, want, c)
}

func TestMissingConfigFileShouldUseDefaultAndEnv(t *testing.T) {
	t.Setenv("MODE", "prod")

	c := defaultConfig()
	err := config.Load("./testdata/abc.yaml", &c)
	require.NoError(t, err)

	want := defaultConfig()
	want.Mode = "prod"
	assert.Equal(t, want, c)
}
