package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("with all environment variables", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "9000")
		t.Setenv("GIN_MODE", "release")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://test:test@localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "test_database", cfg.MongoDBDatabase)
		assert.Equal(t, "9000", cfg.ServerPort)
		assert.Equal(t, "release", cfg.GinMode)
	})

	t.Run("default values are applied when env vars are empty", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "")
		t.Setenv("MONGODB_DATABASE", "")
		t.Setenv("SERVER_PORT", "")
		t.Setenv("GIN_MODE", "")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "idol_database", cfg.MongoDBDatabase)
		assert.Equal(t, "8081", cfg.ServerPort)
		assert.Equal(t, "debug", cfg.GinMode)
	})

	t.Run("partial environment variables use defaults for missing", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://custom:custom@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "")
		t.Setenv("SERVER_PORT", "3000")
		t.Setenv("GIN_MODE", "")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://custom:custom@localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "idol_database", cfg.MongoDBDatabase)
		assert.Equal(t, "3000", cfg.ServerPort)
		assert.Equal(t, "debug", cfg.GinMode)
	})

	t.Run("invalid SERVER_PORT returns error", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "invalid")
		t.Setenv("GIN_MODE", "debug")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("invalid GIN_MODE returns error", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "invalid")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("existing environment variable", func(t *testing.T) {
		key := "TEST_ENV_VAR"
		value := "test_value"
		t.Setenv(key, value)

		result := getEnv(key, "default_value")

		assert.Equal(t, value, result)
	})

	t.Run("non-existing environment variable returns default", func(t *testing.T) {
		key := "NON_EXISTING_VAR"
		defaultValue := "default_value"
		t.Setenv(key, "")

		result := getEnv(key, defaultValue)

		assert.Equal(t, defaultValue, result)
	})

	t.Run("empty environment variable returns default", func(t *testing.T) {
		key := "EMPTY_VAR"
		defaultValue := "default_value"
		t.Setenv(key, "")

		result := getEnv(key, defaultValue)

		assert.Equal(t, defaultValue, result)
	})
}
