package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Server      Server      `yaml:"server"`
	Database    Database    `yaml:"database"`
	Logging     Logging     `yaml:"logging"`
	Tracing     Tracing     `yaml:"tracing"`
	AWS         AWS         `yaml:"aws"`
	Application Application `yaml:"application"`
}

func TestLoadConfigFile(t *testing.T) {
	t.Run("load valid file from env var", func(t *testing.T) {
		os.Setenv("CONFIG_FILE", filepath.Join("testdata", "valid.yaml"))
		defer os.Unsetenv("CONFIG_FILE")

		var cfg testConfig
		err := Load(&cfg)
		require.NoError(t, err)

		assertValidServerConfig(t, cfg)
		assertValidDatabaseConfig(t, cfg)
	})

	t.Run("load valid file from given path", func(t *testing.T) {
		var cfg testConfig
		err := LoadFromFile(filepath.Join("testdata", "valid.yaml"), &cfg)
		require.NoError(t, err)

		assertValidServerConfig(t, cfg)
		assertValidDatabaseConfig(t, cfg)
	})
}

func assertValidServerConfig(t *testing.T, cfg testConfig) {
	assert.Equal(t, "testhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Contains(t, cfg.Server.AllowedOrigins, "http://localhost")
}

func assertValidDatabaseConfig(t *testing.T, cfg testConfig) {
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "dbhost", cfg.Database.Host)
}
