// @title           Idol API
// @version         1.0
// @description     包括的アイドル情報API - アイドル、グループ、事務所、イベント情報を提供
// @termsOfService  https://github.com/kuro48/idol-api/blob/main/backend/static/terms/terms_of_service.md

// @contact.name   Idol API Maintainers
// @contact.url    https://github.com/kuro48/idol-api/issues

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      api.example.com
// @BasePath  /api/v1

// @schemes https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer API key. Example: "Bearer ik_live_..."

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/config"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
	infraAuth "github.com/kuro48/idol-api/internal/infrastructure/auth"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/shared/logger"

	_ "github.com/kuro48/idol-api/docs" // Swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// ロガーの初期化（最初に実行）
	logger.Setup(slog.LevelInfo)

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定読み込みエラー", "error", err)
		os.Exit(1)
	}

	// MongoDBに接続
	db, err := database.Connect(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		slog.Error("データベース接続エラー", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ginモード設定
	gin.SetMode(cfg.GinMode)

	// DDD構造での初期化（詳細は wire.go を参照）
	ctx := context.Background()
	deps := buildApp(ctx, cfg, db)

	// Ginルーターのセットアップ（デフォルトミドルウェアなし）
	router := gin.New()

	// 信頼するプロキシを設定（空の場合はプロキシを信頼しない）
	if cfg.TrustedProxies != "" {
		trustedProxies := strings.Split(cfg.TrustedProxies, ",")
		for i, p := range trustedProxies {
			trustedProxies[i] = strings.TrimSpace(p)
		}
		if err := router.SetTrustedProxies(trustedProxies); err != nil {
			slog.Warn("信頼プロキシ設定エラー", "error", err)
		}
	} else {
		// プロキシを信頼しない（RemoteAddr を直接使用）
		if err := router.SetTrustedProxies(nil); err != nil {
			slog.Warn("信頼プロキシ設定エラー", "error", err)
		}
	}

	// ミドルウェア設定（順序重要）
	router.Use(gin.Recovery())                                         // パニック回復
	router.Use(middleware.Logger())                                    // 構造化ログ
	router.Use(middleware.ErrorHandler())                              // エラーハンドリング
	router.Use(middleware.AuditContext())                              // 監査コンテキスト（作成者・ソース追跡）
	router.Use(middleware.UsageTrackerMiddleware(deps.analyticsAppService)) // API利用トラッキング
	router.Use(middleware.RequestBodyLimit(5 << 20))                   // 5 MiB

	// CORS設定（CORS_ALLOWED_ORIGINS 環境変数で制御）
	corsOrigins := parseCORSOrigins(cfg.CORSAllowedOrigins, cfg.GinMode)
	corsConfig := cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-ID-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// セキュリティヘッダー設定
	router.Use(middleware.SecurityHeaders())

	// レート制限設定（RATE_LIMIT_RPS / RATE_LIMIT_BURST で調整可能）
	// 注意: インメモリ実装のため水平スケール時は値を 1/レプリカ数 に下げるか、ロードバランサー側でも制限すること
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	router.Use(rateLimiter.Limit())
	publicMutationLimiter := middleware.NewRateLimiter(cfg.PublicMutationRateLimitRPS, cfg.PublicMutationRateLimitBurst)

	// ヘルスチェックエンドポイント
	// liveness: プロセスが生きているかのみ確認（依存先チェックなし）
	router.GET("/health/live", deps.healthHandler.Live)
	// readiness: MongoDB疎通確認（依存先が利用可能か確認）
	router.GET("/health/ready", deps.healthHandler.Ready)
	// 後方互換のため /health も維持
	router.GET("/health", deps.healthHandler.Health)

	// Swagger UI（本番環境では無効化）
	if cfg.GinMode != gin.ReleaseMode {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Legacy static frontend shells are development-only. Production should serve
	// the React app from the frontend build/hosting path to avoid exposing stale auth UI.
	if cfg.GinMode != gin.ReleaseMode {
		router.Static("/assets", "./static/web/assets")
		router.GET("/app", func(c *gin.Context) {
			c.File("./static/web/app.html")
		})
		router.GET("/admin", func(c *gin.Context) {
			c.File("./static/web/admin.html")
		})
	}

	// idol-auth 認証の初期化（IDOL_AUTH_URL が設定されている場合のみ有効）
	var oidcVerifier domainAuth.TokenVerifier
	var identityVerifier domainAuth.IdentityVerifier
	if cfg.IdolAuthURL != "" {
		v, err := infraAuth.NewIntrospectionVerifier(cfg.IdolAuthURL)
		if err != nil {
			slog.Error("idol-auth 初期化失敗", "error", err)
			os.Exit(1)
		}
		oidcVerifier = v
		slog.Info("idol-auth 認証が有効です", "url", cfg.IdolAuthURL)
	} else {
		slog.Warn("idol-auth 認証は無効です（IDOL_AUTH_URL 未設定）。write/admin エンドポイントは 503 を返します")
	}
	if cfg.IdolAuthIssuerURL != "" {
		v, err := infraAuth.NewIDTokenVerifier(cfg.IdolAuthIssuerURL, cfg.IdolAuthClientID)
		if err != nil {
			slog.Error("idol-auth ID token検証初期化失敗", "error", err)
			os.Exit(1)
		}
		identityVerifier = v
		slog.Info("idol-auth ID token検証が有効です", "issuer", cfg.IdolAuthIssuerURL)
	} else {
		slog.Warn("idol-auth ID token検証は無効です（IDOL_AUTH_ISSUER_URL 未設定）。ユーザー申請エンドポイントは 503 を返します")
	}

	writeAuth := middleware.OIDCWriteAuth(oidcVerifier)
	adminAuth := middleware.OIDCAdminAuth(oidcVerifier)
	userAuth := middleware.OIDCUserAuth(oidcVerifier, identityVerifier)

	// 認証ミドルウェアと publicMutationLimiter は OIDC 設定後に deps.routes へセット
	deps.routes.writeAuth = writeAuth
	deps.routes.adminAuth = adminAuth
	deps.routes.userAuth = userAuth
	deps.routes.publicMutationLimiter = publicMutationLimiter

	v1 := router.Group("/api/v1")
	registerAPIRoutes(v1, deps.routes)

	// サーバー起動（グレースフルシャットダウン対応）
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// バックグラウンドワーカー用キャンセル可能コンテキスト
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	// 失敗した Webhook 配信を 5 分ごとにリトライ
	deps.webhookAppService.StartRetryWorker(workerCtx, 5*time.Minute)

	slog.Info("サーバーを起動します", "address", addr, "architecture", "DDD")
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("サーバー起動エラー", "error", err)
			os.Exit(1)
		}
	}()

	// SIGTERM/SIGINT を待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("シャットダウン開始...")

	// バックグラウンドワーカーを停止
	workerCancel()

	// インフライトの非同期処理が完了するまで待機
	deps.webhookAppService.Shutdown()
	deps.jobAppService.Shutdown()

	// HTTP サーバーを 30 秒以内にシャットダウン
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTPシャットダウンエラー", "error", err)
	}
	slog.Info("サーバーを正常に停止しました")
}

// indexEnsurer は EnsureIndexes 呼び出しのメタデータと関数をまとめる
type indexEnsurer struct {
	name       string
	collection string
	fn         func(ctx context.Context) error
}

// ensureIndexes は各リポジトリのインデックスを作成する。失敗しても続行する。
func ensureIndexes(ctx context.Context, ensurers []indexEnsurer) {
	for _, e := range ensurers {
		if err := e.fn(ctx); err != nil {
			slog.Warn(e.name+"インデックス作成失敗（続行）", "error", err, "collection", e.collection)
		} else {
			slog.Info(e.name+"インデックス作成完了", "collection", e.collection)
		}
	}
}

func parseCORSOrigins(raw string, ginMode string) []string {
	if strings.TrimSpace(raw) == "" && ginMode != gin.ReleaseMode {
		return []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080"}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
