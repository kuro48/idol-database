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
	"log"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appEvent "github.com/kuro48/idol-api/internal/application/event"
	appExport "github.com/kuro48/idol-api/internal/application/export"
	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	appTag "github.com/kuro48/idol-api/internal/application/tag"
	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/config"
	"github.com/kuro48/idol-api/internal/infrastructure/adapters"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/infrastructure/persistence/mongodb"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	usecaseAgency "github.com/kuro48/idol-api/internal/usecase/agency"
	usecaseEvent "github.com/kuro48/idol-api/internal/usecase/event"
	usecaseGroup "github.com/kuro48/idol-api/internal/usecase/group"
	usecaseIdol "github.com/kuro48/idol-api/internal/usecase/idol"
	usecaseRemoval "github.com/kuro48/idol-api/internal/usecase/removal"
	usecaseTag "github.com/kuro48/idol-api/internal/usecase/tag"

	_ "github.com/kuro48/idol-api/docs" // Swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("設定読み込みエラー:", err)
	}

	// MongoDBに接続
	db, err := database.Connect(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		log.Fatal("データベース接続エラー:", err)
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

	// MongoDBインデックスの作成
	ctx := context.Background()
	if err := idolRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  Idolインデックス作成エラー（続行します）: %v", err)
	} else {
		log.Println("✅ Idol MongoDBインデックスを作成しました")
	}
	if err := eventRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  Eventインデックス作成エラー（続行します）: %v", err)
	} else {
		log.Println("✅ Event MongoDBインデックスを作成しました")
	}
	if err := tagRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  Tagインデックス作成エラー（続行します）: %v", err)
	} else {
		log.Println("✅ Tag MongoDBインデックスを作成しました")
	}
	if err := groupRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  Groupインデックス作成エラー（続行します）: %v", err)
	} else {
		log.Println("✅ Group MongoDBインデックスを作成しました")
	}
	if err := agencyRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  Agencyインデックス作成エラー（続行します）: %v", err)
	} else {
		log.Println("✅ Agency MongoDBインデックスを作成しました")
	}

	// アプリケーション層: アプリケーションサービス
	idolAppService := appIdol.NewApplicationService(idolRepo)
	removalAppService := appRemoval.NewApplicationService(removalRepo)
	groupAppService := appGroup.NewApplicationService(groupRepo)
	agencyAppService := appAgency.NewApplicationService(agencyRepo)
	eventAppService := appEvent.NewApplicationService(eventRepo)
	tagAppService := appTag.NewApplicationService(tagRepo)
	webhookAppService := appWebhook.NewApplicationService(webhookSubRepo, webhookDelRepo)
	exportAppService := appExport.NewApplicationService(exportLogRepo, idolAppService)

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

	// ユースケース層
	idolUsecase := usecaseIdol.NewUsecase(idolAppPort, agencyAppPortForIdol)
	removalUsecase := usecaseRemoval.NewUsecase(removalAppPort, removalIdolPort, removalGroupPort)
	groupUsecase := usecaseGroup.NewUsecase(groupAppPort)
	agencyUsecase := usecaseAgency.NewUsecase(agencyAppPort)
	eventUsecase := usecaseEvent.NewUsecase(eventAppPort)
	tagUsecase := usecaseTag.NewUsecase(tagAppPort)

	// プレゼンテーション層: ハンドラー
	idolHandler := handlers.NewIdolHandler(idolUsecase)
	removalHandler := handlers.NewRemovalHandler(removalUsecase)
	groupHandler := handlers.NewGroupHandler(groupUsecase)
	agencyHandler := handlers.NewAgencyHandler(agencyUsecase)
	eventHandler := handlers.NewEventHandler(eventUsecase)
	tagHandler := handlers.NewTagHandler(tagUsecase)
	termHandler := handlers.NewTermHandler("./static")
	webhookHandler := handlers.NewWebhookHandler(webhookAppService)
	exportHandler := handlers.NewExportHandler(exportAppService)

	// Ginルーターのセットアップ（デフォルトミドルウェアなし）
	router := gin.New()

	// 信頼するプロキシを設定（空の場合はプロキシを信頼しない）
	if cfg.TrustedProxies != "" {
		trustedProxies := strings.Split(cfg.TrustedProxies, ",")
		for i, p := range trustedProxies {
			trustedProxies[i] = strings.TrimSpace(p)
		}
		if err := router.SetTrustedProxies(trustedProxies); err != nil {
			log.Printf("⚠️  信頼プロキシ設定エラー: %v", err)
		}
	} else {
		// プロキシを信頼しない（RemoteAddr を直接使用）
		if err := router.SetTrustedProxies(nil); err != nil {
			log.Printf("⚠️  信頼プロキシ設定エラー: %v", err)
		}
	}

	// ミドルウェア設定（順序重要）
	router.Use(gin.Recovery())                   // パニック回復
	router.Use(middleware.Logger())              // 構造化ログ
	router.Use(middleware.ErrorHandler())        // エラーハンドリング
	router.Use(middleware.AuditContext())        // 監査コンテキスト（作成者・ソース追跡）

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

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

	// サーバー起動
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 サーバーを起動します (DDD architecture): http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}
