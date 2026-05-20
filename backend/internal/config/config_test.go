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
		t.Setenv("IDOL_AUTH_URL", "https://auth.example.com")
		t.Setenv("IDOL_AUTH_ISSUER_URL", "https://issuer.example.com")
		t.Setenv("IDOL_AUTH_CLIENT_ID", "idol-db-frontend")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "mongodb://test:test@localhost:27017", cfg.MongoDBURI)
		assert.Equal(t, "test_database", cfg.MongoDBDatabase)
		assert.Equal(t, "9000", cfg.ServerPort)
		assert.Equal(t, "release", cfg.GinMode)
		assert.Equal(t, "https://auth.example.com", cfg.IdolAuthURL)
		assert.Equal(t, "https://issuer.example.com", cfg.IdolAuthIssuerURL)
		assert.Equal(t, "idol-db-frontend", cfg.IdolAuthClientID)
		assert.Equal(t, "https://app.example.com", cfg.CORSAllowedOrigins)
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
		assert.Equal(t, "", cfg.CORSAllowedOrigins)
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

	t.Run("release mode without IDOL_AUTH_URL returns error", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("IDOL_AUTH_URL", "")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		var valErr *ValidationError
		assert.ErrorAs(t, err, &valErr)
		assert.Equal(t, "IDOL_AUTH_URL", valErr.Field)
	})

	t.Run("release mode without IDOL_AUTH_ISSUER_URL returns error", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("IDOL_AUTH_URL", "https://auth.example.com")
		t.Setenv("IDOL_AUTH_ISSUER_URL", "")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		var valErr *ValidationError
		assert.ErrorAs(t, err, &valErr)
		assert.Equal(t, "IDOL_AUTH_ISSUER_URL", valErr.Field)
	})

	t.Run("release mode with IDOL_AUTH_URL and issuer succeeds", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("IDOL_AUTH_URL", "https://auth.example.com")
		t.Setenv("IDOL_AUTH_ISSUER_URL", "https://issuer.example.com")
		t.Setenv("IDOL_AUTH_CLIENT_ID", "idol-db-frontend")
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "https://auth.example.com", cfg.IdolAuthURL)
		assert.Equal(t, "https://issuer.example.com", cfg.IdolAuthIssuerURL)
	})

	t.Run("release mode without IDOL_AUTH_CLIENT_ID returns error", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("IDOL_AUTH_URL", "https://auth.example.com")
		t.Setenv("IDOL_AUTH_ISSUER_URL", "https://issuer.example.com")
		t.Setenv("IDOL_AUTH_CLIENT_ID", "")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		var valErr *ValidationError
		assert.ErrorAs(t, err, &valErr)
		assert.Equal(t, "IDOL_AUTH_CLIENT_ID", valErr.Field)
	})

	t.Run("release mode without CORS_ALLOWED_ORIGINS returns error", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("IDOL_AUTH_URL", "https://auth.example.com")
		t.Setenv("IDOL_AUTH_ISSUER_URL", "https://issuer.example.com")
		t.Setenv("IDOL_AUTH_CLIENT_ID", "idol-db-frontend")
		t.Setenv("CORS_ALLOWED_ORIGINS", "")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		var valErr *ValidationError
		assert.ErrorAs(t, err, &valErr)
		assert.Equal(t, "CORS_ALLOWED_ORIGINS", valErr.Field)
	})

	t.Run("release mode rejects insecure CORS origin", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://test:test@localhost:27017")
		t.Setenv("MONGODB_DATABASE", "test_database")
		t.Setenv("SERVER_PORT", "8081")
		t.Setenv("GIN_MODE", "release")
		t.Setenv("IDOL_AUTH_URL", "https://auth.example.com")
		t.Setenv("IDOL_AUTH_ISSUER_URL", "https://issuer.example.com")
		t.Setenv("IDOL_AUTH_CLIENT_ID", "idol-db-frontend")
		t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")

		cfg, err := Load()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		var valErr *ValidationError
		assert.ErrorAs(t, err, &valErr)
		assert.Equal(t, "CORS_ALLOWED_ORIGINS", valErr.Field)
	})

	t.Run("public mutation rate limit defaults are restrictive", func(t *testing.T) {
		t.Setenv("PUBLIC_MUTATION_RATE_LIMIT_RPS", "")
		t.Setenv("PUBLIC_MUTATION_RATE_LIMIT_BURST", "")

		cfg, err := Load()

		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, 0.2, cfg.PublicMutationRateLimitRPS)
		assert.Equal(t, 3, cfg.PublicMutationRateLimitBurst)
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
