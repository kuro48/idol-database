package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDBURI      string
	MongoDBDatabase string
	ServerPort      string
	GinMode         string
}

// ValidationError は設定バリデーションエラー
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("設定エラー [%s]: %s", e.Field, e.Message)
}

// Load は環境変数から設定を読み込み、バリデーションを行う
func Load() (*Config, error) {
	// .env.local → .env の順で読み込み
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	cfg := &Config{
		MongoDBURI:      getEnv("MONGODB_URI", ""),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", ""),
		ServerPort:      getEnv("SERVER_PORT", "8081"),
		GinMode:         getEnv("GIN_MODE", "debug"),
	}

	// バリデーション実行
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate は設定値のバリデーションを行う
func (c *Config) Validate() error {
	// MongoDB URI の必須チェック
	if c.MongoDBURI == "" {
		return &ValidationError{
			Field:   "MONGODB_URI",
			Message: "MongoDB接続URIは必須です",
		}
	}

	// MongoDB Database の必須チェック
	if c.MongoDBDatabase == "" {
		return &ValidationError{
			Field:   "MONGODB_DATABASE",
			Message: "MongoDBデータベース名は必須です",
		}
	}

	// ServerPort のバリデーション
	if c.ServerPort == "" {
		return &ValidationError{
			Field:   "SERVER_PORT",
			Message: "サーバーポートは必須です",
		}
	}
	port, err := strconv.Atoi(c.ServerPort)
	if err != nil || port < 1 || port > 65535 {
		return &ValidationError{
			Field:   "SERVER_PORT",
			Message: "サーバーポートは1-65535の数値である必要があります",
		}
	}

	// GinMode のバリデーション
	validModes := map[string]bool{
		"debug":   true,
		"release": true,
		"test":    true,
	}
	if !validModes[c.GinMode] {
		return &ValidationError{
			Field:   "GIN_MODE",
			Message: "GIN_MODEは debug, release, test のいずれかである必要があります",
		}
	}

	return nil
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
