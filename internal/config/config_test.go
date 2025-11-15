package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("with environment variables", func(t *testing.T) {
		// 環境変数を設定
		os.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		os.Setenv("MONGODB_DATABASE", "test_database")
		os.Setenv("SERVER_PORT", "9000")
		os.Setenv("GIN_MODE", "release")
		defer func() {
			os.Unsetenv("MONGODB_URI")
			os.Unsetenv("MONGODB_DATABASE")
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("GIN_MODE")
		}()

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://test:test@localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "test_database", cfg.MongoDBDatabase)
		assert.Equal(t, "9000", cfg.ServerPort)
		assert.Equal(t, "release", cfg.GinMode)
	})

	t.Run("with default values", func(t *testing.T) {
		// 環境変数をクリア
		os.Unsetenv("MONGODB_URI")
		os.Unsetenv("MONGODB_DATABASE")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("GIN_MODE")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "idol_database", cfg.MongoDBDatabase)
		assert.Equal(t, "8081", cfg.ServerPort)
		assert.Equal(t, "debug", cfg.GinMode)
	})

	t.Run("partial environment variables", func(t *testing.T) {
		// 一部の環境変数のみ設定
		os.Setenv("MONGODB_URI", "mongodb://custom:custom@localhost:27017")
		os.Setenv("SERVER_PORT", "3000")
		defer func() {
			os.Unsetenv("MONGODB_URI")
			os.Unsetenv("SERVER_PORT")
		}()

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://custom:custom@localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "idol_database", cfg.MongoDBDatabase) // デフォルト値
		assert.Equal(t, "3000", cfg.ServerPort)
		assert.Equal(t, "debug", cfg.GinMode) // デフォルト値
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("existing environment variable", func(t *testing.T) {
		key := "TEST_ENV_VAR"
		value := "test_value"
		os.Setenv(key, value)
		defer os.Unsetenv(key)

		result := getEnv(key, "default_value")

		assert.Equal(t, value, result)
	})

	t.Run("non-existing environment variable", func(t *testing.T) {
		key := "NON_EXISTING_VAR"
		defaultValue := "default_value"
		os.Unsetenv(key)

		result := getEnv(key, defaultValue)

		assert.Equal(t, defaultValue, result)
	})

	t.Run("empty environment variable", func(t *testing.T) {
		key := "EMPTY_VAR"
		defaultValue := "default_value"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		result := getEnv(key, defaultValue)

		assert.Equal(t, defaultValue, result)
	})
}
