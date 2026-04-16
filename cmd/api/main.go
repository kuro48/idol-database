// @title           Idol API
// @version         1.0
// @description     包括的アイドル情報API - アイドル、グループ、事務所、イベント情報を提供
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@idol-api.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8081
// @BasePath  /api/v1

// @schemes http https

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
	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appAnalytics "github.com/kuro48/idol-api/internal/application/analytics"
	appEvent "github.com/kuro48/idol-api/internal/application/event"
	appExport "github.com/kuro48/idol-api/internal/application/export"
	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appJob "github.com/kuro48/idol-api/internal/application/job"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	appSubmission "github.com/kuro48/idol-api/internal/application/submission"
	appTag "github.com/kuro48/idol-api/internal/application/tag"
	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/config"
	"github.com/kuro48/idol-api/cmd/api/adapters"
	"github.com/kuro48/idol-api/internal/infrastructure/email"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/infrastructure/persistence/mongodb"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/shared/logger"
	usecaseAgency "github.com/kuro48/idol-api/internal/usecase/agency"
	usecaseEvent "github.com/kuro48/idol-api/internal/usecase/event"
	usecaseGroup "github.com/kuro48/idol-api/internal/usecase/group"
	usecaseIdol "github.com/kuro48/idol-api/internal/usecase/idol"
	usecaseRemoval "github.com/kuro48/idol-api/internal/usecase/removal"
	usecaseSubmission "github.com/kuro48/idol-api/internal/usecase/submission"
	usecaseTag "github.com/kuro48/idol-api/internal/usecase/tag"

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

	// MongoDBインデックスの作成
	ctx := context.Background()
	if err := idolRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Idolインデックス作成失敗（続行）", "error", err, "collection", "idols")
	} else {
		slog.Info("Idolインデックス作成完了", "collection", "idols")
	}
	if err := eventRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Eventインデックス作成失敗（続行）", "error", err, "collection", "events")
	} else {
		slog.Info("Eventインデックス作成完了", "collection", "events")
	}
	if err := tagRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Tagインデックス作成失敗（続行）", "error", err, "collection", "tags")
	} else {
		slog.Info("Tagインデックス作成完了", "collection", "tags")
	}
	if err := groupRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Groupインデックス作成失敗（続行）", "error", err, "collection", "groups")
	} else {
		slog.Info("Groupインデックス作成完了", "collection", "groups")
	}
	if err := agencyRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Agencyインデックス作成失敗（続行）", "error", err, "collection", "agencies")
	} else {
		slog.Info("Agencyインデックス作成完了", "collection", "agencies")
	}
	if err := analyticsRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Analyticsインデックス作成失敗（続行）", "error", err, "collection", "api_usage_logs")
	} else {
		slog.Info("Analyticsインデックス作成完了", "collection", "api_usage_logs")
	}
	if err := jobRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Jobインデックス作成失敗（続行）", "error", err, "collection", "async_jobs")
	} else {
		slog.Info("Jobインデックス作成完了", "collection", "async_jobs")
	}
	if err := removalRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Removalインデックス作成失敗（続行）", "error", err, "collection", "removal_requests")
	} else {
		slog.Info("Removalインデックス作成完了", "collection", "removal_requests")
	}
	if err := submissionRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("Submissionインデックス作成失敗（続行）", "error", err, "collection", "submissions")
	} else {
		slog.Info("Submissionインデックス作成完了", "collection", "submissions")
	}
	if err := webhookSubRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("WebhookSubインデックス作成失敗（続行）", "error", err, "collection", "webhook_subscriptions")
	} else {
		slog.Info("WebhookSubインデックス作成完了", "collection", "webhook_subscriptions")
	}
	if err := webhookDelRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("WebhookDelインデックス作成失敗（続行）", "error", err, "collection", "webhook_delivery_logs")
	} else {
		slog.Info("WebhookDelインデックス作成完了", "collection", "webhook_delivery_logs")
	}
	if err := exportLogRepo.EnsureIndexes(ctx); err != nil {
		slog.Warn("ExportLogインデックス作成失敗（続行）", "error", err, "collection", "export_logs")
	} else {
		slog.Info("ExportLogインデックス作成完了", "collection", "export_logs")
	}

	// アプリケーション層: アプリケーションサービス
	analyticsAppService := appAnalytics.NewApplicationService(analyticsRepo)
	jobAppService := appJob.NewApplicationService(jobRepo)
	idolAppService := appIdol.NewApplicationService(idolRepo)
	removalAppService := appRemoval.NewApplicationService(removalRepo)
	groupAppService := appGroup.NewApplicationService(groupRepo)
	agencyAppService := appAgency.NewApplicationService(agencyRepo)
	eventAppService := appEvent.NewApplicationService(eventRepo)
	tagAppService := appTag.NewApplicationService(tagRepo)
	webhookAppService := appWebhook.NewApplicationService(webhookSubRepo, webhookDelRepo)
	exportAppService := appExport.NewApplicationService(exportLogRepo, idolAppService)
	submissionAppService := appSubmission.NewApplicationService(submissionRepo)

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

	// メール通知の初期化（SMTP_HOST が設定されている場合のみ有効化）
	var emailNotifier usecaseSubmission.EmailNotifier
	if cfg.SMTPHost != "" {
		emailNotifier = email.NewSMTPNotifier(email.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
		slog.Info("メール通知が有効です", "smtp_host", cfg.SMTPHost, "smtp_port", cfg.SMTPPort)
	} else {
		slog.Info("メール通知は無効です（SMTP_HOST 未設定）")
	}

	// ユースケース層
	idolUsecase := usecaseIdol.NewUsecase(idolAppPort, agencyAppPortForIdol)
	removalUsecase := usecaseRemoval.NewUsecase(removalAppPort, removalIdolPort, removalGroupPort)
	groupUsecase := usecaseGroup.NewUsecase(groupAppPort)
	agencyUsecase := usecaseAgency.NewUsecase(agencyAppPort)
	eventUsecase := usecaseEvent.NewUsecase(eventAppPort)
	tagUsecase := usecaseTag.NewUsecase(tagAppPort)
	submissionUsecase := usecaseSubmission.NewUsecase(submissionAppPort, emailNotifier)

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
	router.Use(gin.Recovery())                   // パニック回復
	router.Use(middleware.Logger())              // 構造化ログ
	router.Use(middleware.ErrorHandler())        // エラーハンドリング
	router.Use(middleware.AuditContext())        // 監査コンテキスト（作成者・ソース追跡）
	router.Use(middleware.UsageTrackerMiddleware(analyticsAppService)) // API利用トラッキング

	// CORS設定（CORS_ALLOWED_ORIGINS 環境変数で制御）
	corsOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
	corsConfig := cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
	router.Use(cors.New(corsConfig))

	// セキュリティヘッダー設定
	router.Use(middleware.SecurityHeaders())

	// レート制限設定（10リクエスト/秒、バースト20）
	rateLimiter := middleware.NewRateLimiter(10, 20)
	router.Use(rateLimiter.Limit())

	// ヘルスチェックエンドポイント
	// liveness: プロセスが生きているかのみ確認（依存先チェックなし）
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// readiness: MongoDB疎通確認（依存先が利用可能か確認）
	router.GET("/health/ready", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(503, gin.H{"status": "unavailable", "error": "database unreachable"})
			return
		}
		c.JSON(200, gin.H{"status": "ok"})
	})
	// 後方互換のため /health も維持
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Idol API is running with DDD architecture",
		})
	})

	// Swagger UI（本番環境では無効化）
	if cfg.GinMode != gin.ReleaseMode {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// APIキー設定
	apiKeyCfg := middleware.APIKeyConfig{
		WriteAPIKey: cfg.WriteAPIKey,
		AdminAPIKey: cfg.AdminAPIKey,
	}
	writeAuth := middleware.WriteAuth(apiKeyCfg)
	adminAuth := middleware.AdminAuth(cfg.AdminAPIKey)

	v1 := router.Group("/api/v1")
	{
		// アイドル: 読み取りは公開、書き込みは write スコープ必須
		idols := v1.Group("/idols")
		{
			idols.GET("", idolHandler.ListIdols)                                   // 一覧取得（公開）
			idols.GET("/:id", idolHandler.GetIdol)                                // 詳細取得（公開）
			idols.GET("/:id/external-ids", idolHandler.GetExternalIDs)            // 外部IDマッピング取得（公開）
		}
		idolsWrite := v1.Group("/idols", writeAuth)
		{
			idolsWrite.POST("", idolHandler.CreateIdol)                            // 新規作成
			idolsWrite.POST("/bulk", idolHandler.BulkCreateIdols)                  // バルク作成
			idolsWrite.PUT("/:id", idolHandler.UpdateIdol)                         // 更新
			idolsWrite.DELETE("/:id", idolHandler.DeleteIdol)                      // 削除
			idolsWrite.PUT("/:id/social-links", idolHandler.UpdateSocialLinks)     // SNSリンク更新
			idolsWrite.PUT("/:id/external-ids", idolHandler.UpdateExternalIDs)     // 外部IDマッピング更新
		}

		// 削除申請: 申請・参照は公開、管理は admin スコープ必須
		removalRequests := v1.Group("/removal-requests")
		{
			removalRequests.POST("", removalHandler.CreateRemovalRequest) // 削除申請作成（公開）
			removalRequests.GET("/:id", removalHandler.GetRemovalRequest) // 削除申請詳細取得（公開）
		}
		adminRemoval := v1.Group("/removal-requests", adminAuth)
		{
			adminRemoval.GET("", removalHandler.ListAllRemovalRequests)             // 全削除申請取得
			adminRemoval.GET("/pending", removalHandler.ListPendingRemovalRequests) // 保留中取得
			adminRemoval.PUT("/:id", removalHandler.UpdateStatus)                   // ステータス更新
		}
		idolsAdmin := v1.Group("/idols", adminAuth)
		{
			idolsAdmin.PUT("/:id/restore", idolHandler.RestoreIdol)                          // アイドル復元
			idolsAdmin.GET("/:id/duplicate-candidates", idolHandler.GetDuplicateCandidates)  // 重複候補取得
		}

		// API利用分析（admin スコープ必須）
		adminAnalytics := v1.Group("/admin/analytics", adminAuth)
		{
			adminAnalytics.GET("/usage", analyticsHandler.GetUsageSummary) // API利用サマリー取得
		}

		// 非同期ジョブ管理（admin スコープ必須）
		adminJobs := v1.Group("/admin/jobs", adminAuth)
		{
			adminJobs.POST("/bulk-import", jobHandler.EnqueueBulkImport) // バルクインポートジョブ作成
			adminJobs.GET("/:id", jobHandler.GetJobStatus)               // ジョブステータス取得
			adminJobs.POST("/:id/retry", jobHandler.RetryJob)            // ジョブリトライ
		}

		// Webhook管理（admin スコープ必須）
		adminWebhooks := v1.Group("/admin/webhooks", adminAuth)
		{
			adminWebhooks.POST("", webhookHandler.CreateSubscription)       // 購読作成
			adminWebhooks.GET("", webhookHandler.ListSubscriptions)         // 購読一覧
			adminWebhooks.DELETE("/:id", webhookHandler.DeleteSubscription) // 購読削除
		}

		// Webhook受信エンドポイント（公開: 外部からの受信）
		v1.POST("/webhooks/receive/:subscription_id", webhookHandler.ReceiveWebhook)

		// エクスポート（admin スコープ必須）
		adminExport := v1.Group("/admin/export", adminAuth)
		{
			adminExport.GET("/idols", exportHandler.ExportIdols)   // アイドルエクスポート
			adminExport.GET("/logs", exportHandler.ListExportLogs) // 実行履歴
		}

		// グループ: 読み取りは公開、書き込みは write スコープ必須
		groups := v1.Group("/groups")
		{
			groups.GET("", groupHandler.ListGroup)
			groups.GET("/:id", groupHandler.GetGroup)
		}
		groupsWrite := v1.Group("/groups", writeAuth)
		{
			groupsWrite.POST("", groupHandler.CreateGroup)
			groupsWrite.PUT("/:id", groupHandler.UpdateGroup)
			groupsWrite.DELETE("/:id", groupHandler.DeleteGroup)
		}

		// 事務所: 読み取りは公開、書き込みは write スコープ必須
		agencies := v1.Group("/agencies")
		{
			agencies.GET("", agencyHandler.ListAgencies)
			agencies.GET("/:id", agencyHandler.GetAgency)
		}
		agenciesWrite := v1.Group("/agencies", writeAuth)
		{
			agenciesWrite.POST("", agencyHandler.CreateAgency)
			agenciesWrite.PUT("/:id", agencyHandler.UpdateAgency)
			agenciesWrite.DELETE("/:id", agencyHandler.DeleteAgency)
		}

		terms := v1.Group("/terms")
		{
			terms.GET("/service", termHandler.ShowTermsOfService)
			terms.GET("/privacy", termHandler.ShowPrivacyPolicy)
		}

		// イベント: 読み取りは公開、書き込みは write スコープ必須
		events := v1.Group("/events")
		{
			events.GET("", eventHandler.ListEvents)            // イベント一覧取得（検索機能付き）
			events.GET("/upcoming", eventHandler.GetUpcomingEvents) // 今後のイベント取得
			events.GET("/:id", eventHandler.GetEvent)          // イベント詳細取得
		}
		eventsWrite := v1.Group("/events", writeAuth)
		{
			eventsWrite.POST("", eventHandler.CreateEvent)                               // イベント作成
			eventsWrite.PUT("/:id", eventHandler.UpdateEvent)                            // イベント更新
			eventsWrite.DELETE("/:id", eventHandler.DeleteEvent)                         // イベント削除
			eventsWrite.POST("/:id/performers", eventHandler.AddPerformer)               // パフォーマー追加
			eventsWrite.DELETE("/:id/performers/:performer_id", eventHandler.RemovePerformer) // パフォーマー削除
		}

		// 投稿審査: 作成・取得は公開、審査は admin スコープ必須
		submissions := v1.Group("/submissions")
		{
			submissions.POST("", submissionHandler.CreateSubmission)          // 投稿作成（公開）
			submissions.GET("/:id", submissionHandler.GetSubmission)          // 投稿詳細取得（公開）
			submissions.PUT("/:id/revise", submissionHandler.ReviseSubmission) // 差し戻し後の再投稿（公開）
		}
		adminSubmissions := v1.Group("/submissions", adminAuth)
		{
			adminSubmissions.GET("", submissionHandler.ListAllSubmissions)          // 全投稿一覧
			adminSubmissions.GET("/pending", submissionHandler.ListPendingSubmissions) // 審査待ち一覧
			adminSubmissions.PUT("/:id/status", submissionHandler.UpdateStatus)    // ステータス更新
		}

		// タグ: 読み取りは公開、書き込みは write スコープ必須
		tags := v1.Group("/tags")
		{
			tags.GET("", tagHandler.ListTags)      // タグ一覧取得
			tags.GET("/:id", tagHandler.GetTag)    // タグ詳細取得
		}
		tagsWrite := v1.Group("/tags", writeAuth)
		{
			tagsWrite.POST("", tagHandler.CreateTag)       // タグ作成
			tagsWrite.PUT("/:id", tagHandler.UpdateTag)    // タグ更新
			tagsWrite.DELETE("/:id", tagHandler.DeleteTag) // タグ削除
		}
	}

	// サーバー起動（グレースフルシャットダウン対応）
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

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
