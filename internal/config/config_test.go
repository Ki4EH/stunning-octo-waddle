package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("successfully loads config with defaults", func(t *testing.T) {
		t.Setenv("DATABASE_USER", "testuser")
		t.Setenv("DATABASE_PASSWORD", "testpass")
		t.Setenv("DATABASE_NAME", "testdb")
		t.Setenv("DATABASE_HOST", "localhost")

		cfg, err := LoadConfig()
		require.NoError(t, err, "should not return error when all required vars are set")

		assert.Equal(t, "5433", cfg.DatabasePort, "should use default DATABASE_PORT")
		assert.Equal(t, "testuser", cfg.DatabaseUser)
		assert.Equal(t, "testpass", cfg.DatabasePassword)
		assert.Equal(t, "testdb", cfg.DatabaseName)
		assert.Equal(t, "localhost", cfg.DatabaseHost)
		assert.Equal(t, "8080", cfg.ServerPort, "should use default SERVER_PORT")
	})

	t.Run("successfully overrides defaults", func(t *testing.T) {
		t.Setenv("DATABASE_USER", "user")
		t.Setenv("DATABASE_PASSWORD", "pass")
		t.Setenv("DATABASE_NAME", "db")
		t.Setenv("DATABASE_HOST", "dbhost")
		t.Setenv("DATABASE_PORT", "3306")
		t.Setenv("SERVER_PORT", "3000")
		t.Setenv("ENVIRONMENT", "staging")

		cfg, err := LoadConfig()
		require.NoError(t, err)

		assert.Equal(t, "3306", cfg.DatabasePort, "should override default DATABASE_PORT")
		assert.Equal(t, "3000", cfg.ServerPort, "should override default SERVER_PORT")
		assert.Equal(t, "staging", cfg.Environment, "should set optional field")
	})
}
