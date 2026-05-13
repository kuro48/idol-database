package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDBURI         string
	MongoDBDatabase    string
	ServerPort         string
	GinMode            string
	CORSAllowedOrigins string        // カンマ区切り。空の場合はデフォルト値を使用
	TrustedProxies     string        // カンマ区切りの信頼プロキシIPレンジ（空の場合はプロキシ信頼なし）
	WebhookTimeout     time.Duration // WebhookHTTPクライアントのタイムアウト（WEBHOOK_TIMEOUT_SECONDS で変更可能、デフォルト: 10秒）
	RateLimitRPS       float64       // 1秒あたりのリクエスト数上限（RATE_LIMIT_RPS、デフォルト: 10）
	RateLimitBurst     int           // バースト許容数（RATE_LIMIT_BURST、デフォルト: 20）
	// 公開POST系（投稿・削除申請・外部Webhook受信）に追加適用する低レート制限
	PublicMutationRateLimitRPS   float64 // PUBLIC_MUTATION_RATE_LIMIT_RPS、デフォルト: 0.2
	PublicMutationRateLimitBurst int     // PUBLIC_MUTATION_RATE_LIMIT_BURST、デフォルト: 3
	// idol-auth 認証設定（空の場合は write/admin エンドポイントが 503 を返す）
	IdolAuthURL string // idol-auth の公開 URL（IDOL_AUTH_URL、例: https://auth.example.com）
	// SMTP メール通知設定（SMTP_HOST が空の場合はメール通知を無効化）
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string // 送信元メールアドレス
	SMTPFromName string // 送信元表示名
	// Stripe 決済設定（STRIPE_SECRET_KEY が空の場合は決済機能を無効化）
	StripeSecretKey      string // sk_live_... または sk_test_...
	StripeWebhookSecret  string // whsec_...（Webhook署名検証用）
	StripeKeySeedSecret  string // 決済完了時のAPIキー決定生成用シークレット
	StripePriceDeveloper string // Developer プランの Stripe Price ID
	StripePriceBusiness  string // Business プランの Stripe Price ID
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

	webhookTimeoutSec, err := strconv.Atoi(getEnv("WEBHOOK_TIMEOUT_SECONDS", "10"))
	if err != nil || webhookTimeoutSec <= 0 {
		webhookTimeoutSec = 10
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil || smtpPort <= 0 {
		smtpPort = 587
	}

	rateLimitRPS, err := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "10"), 64)
	if err != nil || rateLimitRPS <= 0 {
		rateLimitRPS = 10
	}

	rateLimitBurst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	if err != nil || rateLimitBurst <= 0 {
		rateLimitBurst = 20
	}

	publicMutationRateLimitRPS, err := strconv.ParseFloat(getEnv("PUBLIC_MUTATION_RATE_LIMIT_RPS", "0.2"), 64)
	if err != nil || publicMutationRateLimitRPS <= 0 {
		publicMutationRateLimitRPS = 0.2
	}

	publicMutationRateLimitBurst, err := strconv.Atoi(getEnv("PUBLIC_MUTATION_RATE_LIMIT_BURST", "3"))
	if err != nil || publicMutationRateLimitBurst <= 0 {
		publicMutationRateLimitBurst = 3
	}

	cfg := &Config{
		MongoDBURI:                   getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase:              getEnv("MONGODB_DATABASE", "idol_database"),
		ServerPort:                   getEnv("SERVER_PORT", "8081"),
		GinMode:                      getEnv("GIN_MODE", "debug"),
		CORSAllowedOrigins:           getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080"),
		TrustedProxies:               getEnv("TRUSTED_PROXIES", ""),
		WebhookTimeout:               time.Duration(webhookTimeoutSec) * time.Second,
		RateLimitRPS:                 rateLimitRPS,
		RateLimitBurst:               rateLimitBurst,
		PublicMutationRateLimitRPS:   publicMutationRateLimitRPS,
		PublicMutationRateLimitBurst: publicMutationRateLimitBurst,
		IdolAuthURL:                  getEnv("IDOL_AUTH_URL", ""),
		SMTPHost:                     getEnv("SMTP_HOST", ""),
		SMTPPort:                     smtpPort,
		SMTPUsername:                 getEnv("SMTP_USERNAME", ""),
		SMTPPassword:                 getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                     getEnv("SMTP_FROM", ""),
		SMTPFromName:                 getEnv("SMTP_FROM_NAME", "Idol API"),
		StripeSecretKey:              getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:          getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeKeySeedSecret:          getEnv("STRIPE_KEY_SEED_SECRET", ""),
		StripePriceDeveloper:         getEnv("STRIPE_PRICE_DEVELOPER", ""),
		StripePriceBusiness:          getEnv("STRIPE_PRICE_BUSINESS", ""),
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

	// 本番モードでは IdolAuthURL を必須とする
	if c.GinMode == "release" && c.IdolAuthURL == "" {
		return &ValidationError{
			Field:   "IDOL_AUTH_URL",
			Message: "本番環境では IDOL_AUTH_URL の設定が必須です",
		}
	}

	// Stripe を有効にする場合は必須項目を揃える
	if c.StripeSecretKey != "" {
		if c.StripeWebhookSecret == "" {
			return &ValidationError{Field: "STRIPE_WEBHOOK_SECRET", Message: "Stripe決済を有効化する場合は Webhook secret が必須です"}
		}
		if c.StripeKeySeedSecret == "" {
			return &ValidationError{Field: "STRIPE_KEY_SEED_SECRET", Message: "Stripe決済を有効化する場合は APIキー生成シークレットが必須です"}
		}
		if c.StripePriceDeveloper == "" {
			return &ValidationError{Field: "STRIPE_PRICE_DEVELOPER", Message: "Developer プランの Stripe Price ID は必須です"}
		}
		if c.StripePriceBusiness == "" {
			return &ValidationError{Field: "STRIPE_PRICE_BUSINESS", Message: "Business プランの Stripe Price ID は必須です"}
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
