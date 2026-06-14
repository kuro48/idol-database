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
	"github.com/kuro48/idol-api/cmd/api/adapters"
	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appAnalytics "github.com/kuro48/idol-api/internal/application/analytics"
	appAPIKey "github.com/kuro48/idol-api/internal/application/apikey"
	appBilling "github.com/kuro48/idol-api/internal/application/billing"
	appEditHistory "github.com/kuro48/idol-api/internal/application/edithistory"
	appEvent "github.com/kuro48/idol-api/internal/application/event"
	appExport "github.com/kuro48/idol-api/internal/application/export"
	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appJob "github.com/kuro48/idol-api/internal/application/job"
	appMembership "github.com/kuro48/idol-api/internal/application/membership"
	appRelease "github.com/kuro48/idol-api/internal/application/release"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	appSubmission "github.com/kuro48/idol-api/internal/application/submission"
	appTag "github.com/kuro48/idol-api/internal/application/tag"
	appVenue "github.com/kuro48/idol-api/internal/application/venue"
	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/config"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
	"github.com/kuro48/idol-api/internal/domain/plan"
	"github.com/kuro48/idol-api/internal/infrastructure/adapters/email"
	infraAuth "github.com/kuro48/idol-api/internal/infrastructure/auth"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/infrastructure/persistence/mongodb"
	infraStripe "github.com/kuro48/idol-api/internal/infrastructure/stripe"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/shared/logger"
	usecaseAgency "github.com/kuro48/idol-api/internal/usecase/agency"
	usecaseEditHistory "github.com/kuro48/idol-api/internal/usecase/edithistory"
	usecaseEvent "github.com/kuro48/idol-api/internal/usecase/event"
	usecaseGroup "github.com/kuro48/idol-api/internal/usecase/group"
	usecaseIdol "github.com/kuro48/idol-api/internal/usecase/idol"
	usecaseMembership "github.com/kuro48/idol-api/internal/usecase/membership"
	usecaseRelease "github.com/kuro48/idol-api/internal/usecase/release"
	usecaseRemoval "github.com/kuro48/idol-api/internal/usecase/removal"
	usecaseSubmission "github.com/kuro48/idol-api/internal/usecase/submission"
	usecaseTag "github.com/kuro48/idol-api/internal/usecase/tag"
	usecaseVenue "github.com/kuro48/idol-api/internal/usecase/venue"

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

	// DDD構造での初期化
	// インフラ層: リポジトリ
	idolRepo := mongodb.NewIdolRepository(db.Database)
	removalRepo := mongodb.NewRemovalRepository(db.Database)
	groupRepo := mongodb.NewGroupRepository(db.Database)
	agencyRepo := mongodb.NewAgencyRepository(db.Database)
	eventRepo := mongodb.NewEventRepository(db.Database)
	tagRepo := mongodb.NewTagRepository(db.Database)
	webhookSubRepo := mongodb.NewWebhookSubscriptionRepository(db.Database)
	webhookDelRepo := mongodb.NewWebhookDeliveryRepository(db.Database)
	exportLogRepo := mongodb.NewExportLogRepository(db.Database)
	analyticsRepo := mongodb.NewAnalyticsRepository(db.Database)
	jobRepo := mongodb.NewJobRepository(db.Database)
	submissionRepo := mongodb.NewSubmissionRepository(db.Database)
	apikeyRepo := mongodb.NewAPIKeyRepository(db.Database)
	usageRepo := mongodb.NewUsageRepository(db.Database)
	billingRepo := mongodb.NewBillingFulfillmentRepository(db.Database)
	releaseRepo := mongodb.NewReleaseRepository(db.Database)
	editHistoryRepo := mongodb.NewEditHistoryRepository(db.Database)
	membershipRepo := mongodb.NewMembershipRepository(db.Database)
	venueRepo := mongodb.NewVenueRepository(db.Database)

	// MongoDBインデックスの作成（失敗しても続行する）
	ctx := context.Background()
	ensureIndexes(ctx, []indexEnsurer{
		{name: "Idol", collection: "idols", fn: idolRepo.EnsureIndexes},
		{name: "Event", collection: "events", fn: eventRepo.EnsureIndexes},
		{name: "Tag", collection: "tags", fn: tagRepo.EnsureIndexes},
		{name: "Group", collection: "groups", fn: groupRepo.EnsureIndexes},
		{name: "Agency", collection: "agencies", fn: agencyRepo.EnsureIndexes},
		{name: "Analytics", collection: "api_usage_logs", fn: analyticsRepo.EnsureIndexes},
		{name: "Job", collection: "async_jobs", fn: jobRepo.EnsureIndexes},
		{name: "Removal", collection: "removal_requests", fn: removalRepo.EnsureIndexes},
		{name: "Submission", collection: "submissions", fn: submissionRepo.EnsureIndexes},
		{name: "APIKey", collection: "api_keys", fn: apikeyRepo.EnsureIndexes},
		{name: "Usage", collection: "api_key_usage", fn: usageRepo.EnsureIndexes},
		{name: "WebhookSub", collection: "webhook_subscriptions", fn: webhookSubRepo.EnsureIndexes},
		{name: "WebhookDel", collection: "webhook_delivery_logs", fn: webhookDelRepo.EnsureIndexes},
		{name: "ExportLog", collection: "export_logs", fn: exportLogRepo.EnsureIndexes},
		{name: "BillingFulfillment", collection: "billing_fulfillments", fn: billingRepo.EnsureIndexes},
		{name: "Release", collection: "releases", fn: releaseRepo.EnsureIndexes},
		{name: "EditHistory", collection: "edit_history", fn: editHistoryRepo.EnsureIndexes},
		{name: "Membership", collection: "memberships", fn: membershipRepo.EnsureIndexes},
		{name: "Venue", collection: "venues", fn: venueRepo.EnsureIndexes},
	})

	// アプリケーション層: アプリケーションサービス
	analyticsAppService := appAnalytics.NewApplicationService(analyticsRepo)
	webhookAppService := appWebhook.NewApplicationServiceWithTimeout(webhookSubRepo, webhookDelRepo, cfg.WebhookTimeout)
	idolAppService := appIdol.NewApplicationService(idolRepo, webhookAppService)
	removalAppService := appRemoval.NewApplicationService(removalRepo)
	groupAppService := appGroup.NewApplicationService(groupRepo, webhookAppService)
	agencyAppService := appAgency.NewApplicationService(agencyRepo, webhookAppService)
	eventAppService := appEvent.NewApplicationService(eventRepo, webhookAppService)
	jobAppService := appJob.NewApplicationService(jobRepo, idolAppService)
	tagAppService := appTag.NewApplicationService(tagRepo)
	exportAppService := appExport.NewApplicationService(exportLogRepo, idolAppService)
	submissionAppService := appSubmission.NewApplicationService(submissionRepo)
	apikeyAppService := appAPIKey.NewApplicationService(apikeyRepo)
	releaseAppService := appRelease.NewApplicationService(releaseRepo, webhookAppService)
	editHistoryAppService := appEditHistory.NewApplicationService(editHistoryRepo)
	membershipAppService := appMembership.NewApplicationService(membershipRepo)
	venueAppService := appVenue.NewApplicationService(venueRepo)

	// 起動時に RUNNING 状態で止まっているジョブを PENDING に戻す
	if err := jobAppService.RecoverStuckJobs(ctx); err != nil {
		slog.Warn("スタックジョブのリカバリ失敗（続行）", "error", err)
	}

	// アダプター層: application サービスを usecase output port に適合させる
	idolAppPort := adapters.NewIdolAppAdapter(idolAppService)
	agencyAppPortForIdol := adapters.NewAgencyAppAdapter(agencyAppService)
	removalAppPort := adapters.NewRemovalAppAdapter(removalAppService)
	removalIdolPort := adapters.NewRemovalIdolAdapter(idolAppService)
	removalGroupPort := adapters.NewRemovalGroupAdapter(groupAppService)
	groupAppPort := adapters.NewGroupAppAdapter(groupAppService)
	agencyAppPort := adapters.NewAgencyAppAdapterForUsecase(agencyAppService)
	eventAppPort := adapters.NewEventAppAdapter(eventAppService)
	tagAppPort := adapters.NewTagAppAdapter(tagAppService)
	submissionAppPort := adapters.NewSubmissionAppAdapter(submissionAppService)
	submissionTargetPort := adapters.NewSubmissionTargetAppAdapter(idolAppService, groupAppService, agencyAppService, eventAppService)
	releaseAppPort := adapters.NewReleaseAppAdapter(releaseAppService)
	releaseIdolPort := adapters.NewIdolExistenceAdapter(idolAppService)
	releaseGroupPort := adapters.NewGroupExistenceAdapter(groupAppService)
	editHistoryAppPort := adapters.NewEditHistoryAppAdapter(editHistoryAppService)
	membershipAppPort := adapters.NewMembershipAppAdapter(membershipAppService)
	venueAppPort := adapters.NewVenueAppAdapter(venueAppService)

	// メール通知の初期化（SMTP_HOST が設定されている場合のみ有効化）
	// インターフェース型変数を使うことで SMTP 未設定時に真の nil（typed-nil ではない）になる。
	var smtpNotifier *email.SMTPNotifier
	var emailNotifier usecaseSubmission.EmailNotifier
	var removalNotifier usecaseRemoval.RemovalNotifier
	if cfg.SMTPHost != "" {
		smtpNotifier = email.NewSMTPNotifier(email.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
		emailNotifier = smtpNotifier
		removalNotifier = smtpNotifier
		slog.Info("メール通知が有効です", "smtp_host", cfg.SMTPHost, "smtp_port", cfg.SMTPPort)
	} else {
		slog.Info("メール通知は無効です（SMTP_HOST 未設定）")
	}

	// ユースケース層
	idolUsecase := usecaseIdol.NewUsecase(idolAppPort, agencyAppPortForIdol)
	removalUsecase := usecaseRemoval.NewUsecase(removalAppPort, removalIdolPort, removalGroupPort, removalNotifier, webhookAppService)
	groupUsecase := usecaseGroup.NewUsecase(groupAppPort)
	agencyUsecase := usecaseAgency.NewUsecase(agencyAppPort)
	eventUsecase := usecaseEvent.NewUsecase(eventAppPort)
	tagUsecase := usecaseTag.NewUsecase(tagAppPort)
	submissionUsecase := usecaseSubmission.NewUsecase(submissionAppPort, submissionTargetPort, emailNotifier)
	releaseUsecase := usecaseRelease.NewUsecase(releaseAppPort, releaseIdolPort, releaseGroupPort)
	editHistoryUsecase := usecaseEditHistory.NewUsecase(editHistoryAppPort)
	membershipUsecase := usecaseMembership.NewUsecase(membershipAppPort)
	venueUsecase := usecaseVenue.NewUsecase(venueAppPort)

	// プレゼンテーション層: ハンドラー
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsAppService)
	jobHandler := handlers.NewJobHandler(jobAppService)
	idolHandler := handlers.NewIdolHandler(idolUsecase)
	removalHandler := handlers.NewRemovalHandler(removalUsecase)
	groupHandler := handlers.NewGroupHandler(groupUsecase)
	agencyHandler := handlers.NewAgencyHandler(agencyUsecase)
	eventHandler := handlers.NewEventHandler(eventUsecase)
	tagHandler := handlers.NewTagHandler(tagUsecase)
	termHandler := handlers.NewTermHandler("./static")
	webhookHandler := handlers.NewWebhookHandler(adapters.NewWebhookAppAdapter(webhookAppService))
	exportHandler := handlers.NewExportHandler(exportAppService)
	submissionHandler := handlers.NewSubmissionHandler(submissionUsecase)
	releaseHandler := handlers.NewReleaseHandler(releaseUsecase)
	editHistoryHandler := handlers.NewEditHistoryHandler(editHistoryUsecase)
	membershipHandler := handlers.NewMembershipHandler(membershipUsecase)
	venueHandler := handlers.NewVenueHandler(venueUsecase)
	apikeyHandler := handlers.NewAPIKeyHandler(apikeyAppService)
	meHandler := handlers.NewMeHandler()
	healthHandler := handlers.NewHealthHandler(db)
	var billingHandler *handlers.BillingHandler
	if cfg.StripeSecretKey != "" && smtpNotifier != nil {
		billingService := appBilling.NewService(
			infraStripe.NewClient(cfg.StripeSecretKey, cfg.StripeWebhookSecret),
			billingRepo,
			apikeyAppService,
			smtpNotifier,
			appBilling.Config{
				StripeSigningSecret: cfg.StripeWebhookSecret,
				KeySeedSecret:       cfg.StripeKeySeedSecret,
				PriceIDs: map[plan.Type]string{
					plan.TypeDeveloper: cfg.StripePriceDeveloper,
					plan.TypeBusiness:  cfg.StripePriceBusiness,
				},
			},
		)
		billingHandler = handlers.NewBillingHandlerWithAllowedRedirectOrigins(billingService, parseCORSOrigins(cfg.CORSAllowedOrigins, cfg.GinMode))
		slog.Info("Stripe課金導線が有効です")
	} else {
		slog.Info("Stripe課金導線は無効です", "stripe_enabled", cfg.StripeSecretKey != "", "smtp_enabled", smtpNotifier != nil)
	}

	// プランベース認証ミドルウェア（外部開発者向けAPIキー）
	// Auth: APIキー必須。検証に成功したリクエストのみ使用量をカウントして通過させる。
	planAuth := middleware.NewPlanAuth(apikeyRepo, usageRepo)

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
	router.Use(middleware.UsageTrackerMiddleware(analyticsAppService)) // API利用トラッキング
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
	router.GET("/health/live", healthHandler.Live)
	// readiness: MongoDB疎通確認（依存先が利用可能か確認）
	router.GET("/health/ready", healthHandler.Ready)
	// 後方互換のため /health も維持
	router.GET("/health", healthHandler.Health)

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

	v1 := router.Group("/api/v1")
	registerAPIRoutes(v1, routeDeps{
		writeAuth:             writeAuth,
		adminAuth:             adminAuth,
		userAuth:              userAuth,
		publicMutationLimiter: publicMutationLimiter,
		planAuth:              planAuth,
		idolHandler:           idolHandler,
		removalHandler:        removalHandler,
		groupHandler:          groupHandler,
		agencyHandler:         agencyHandler,
		eventHandler:          eventHandler,
		tagHandler:            tagHandler,
		webhookHandler:        webhookHandler,
		exportHandler:         exportHandler,
		submissionHandler:     submissionHandler,
		releaseHandler:        releaseHandler,
		editHistoryHandler:    editHistoryHandler,
		membershipHandler:     membershipHandler,
		venueHandler:          venueHandler,
		apikeyHandler:         apikeyHandler,
		meHandler:             meHandler,
		analyticsHandler:      analyticsHandler,
		jobHandler:            jobHandler,
		termHandler:           termHandler,
		billingHandler:        billingHandler,
	})

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
	webhookAppService.StartRetryWorker(workerCtx, 5*time.Minute)

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
	webhookAppService.Shutdown()
	jobAppService.Shutdown()

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
