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
	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	appTag "github.com/kuro48/idol-api/internal/application/tag"
	"github.com/kuro48/idol-api/internal/config"
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

	// ユースケース層
	idolUsecase := usecaseIdol.NewUsecase(idolAppService, agencyAppService)
	removalUsecase := usecaseRemoval.NewUsecase(removalAppService, idolAppService, groupAppService)
	groupUsecase := usecaseGroup.NewUsecase(groupAppService)
	agencyUsecase := usecaseAgency.NewUsecase(agencyAppService)
	eventUsecase := usecaseEvent.NewUsecase(eventAppService)
	tagUsecase := usecaseTag.NewUsecase(tagAppService)

	// プレゼンテーション層: ハンドラー
	idolHandler := handlers.NewIdolHandler(idolUsecase)
	removalHandler := handlers.NewRemovalHandler(removalUsecase)
	groupHandler := handlers.NewGroupHandler(groupUsecase)
	agencyHandler := handlers.NewAgencyHandler(agencyUsecase)
	eventHandler := handlers.NewEventHandler(eventUsecase)
	tagHandler := handlers.NewTagHandler(tagUsecase)
	termHandler := handlers.NewTermHandler("./static")

	// Ginルーターのセットアップ（デフォルトミドルウェアなし）
	router := gin.New()

	// ミドルウェア設定（順序重要）
	router.Use(gin.Recovery())                   // パニック回復
	router.Use(middleware.Logger())              // 構造化ログ
	router.Use(middleware.ErrorHandler())        // エラーハンドリング

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
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Idol API is running with DDD architecture",
		})
	})

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		idols := v1.Group("/idols")
		{
			idols.POST("", idolHandler.CreateIdol)                     // 新規作成
			idols.GET("", idolHandler.ListIdols)                       // 一覧取得
			idols.GET("/:id", idolHandler.GetIdol)                     // 詳細取得
			idols.PUT("/:id", idolHandler.UpdateIdol)                  // 更新
			idols.DELETE("/:id", idolHandler.DeleteIdol)               // 削除
			idols.PUT("/:id/social-links", idolHandler.UpdateSocialLinks) // SNSリンク更新
		}

		removalRequests := v1.Group("/removal-requests")
		{
			removalRequests.POST("", removalHandler.CreateRemovalRequest)              // 削除申請作成
			removalRequests.GET("", removalHandler.ListAllRemovalRequests)             // 全削除申請取得（管理者用）
			removalRequests.GET("/pending", removalHandler.ListPendingRemovalRequests) // 保留中取得（管理者用）
			removalRequests.GET("/:id", removalHandler.GetRemovalRequest)              // 削除申請詳細取得
			removalRequests.PUT("/:id", removalHandler.UpdateStatus)                   // ステータス更新（管理者用）
		}

		groups := v1.Group("/groups")
		{
			groups.POST("", groupHandler.CreateGroup)
			groups.GET("", groupHandler.ListGroup)
			groups.GET("/:id", groupHandler.GetGroup)
			groups.PUT("/:id", groupHandler.UpdateGroup)
			groups.DELETE("/:id", groupHandler.DeleteGroup)
		}

		agencies := v1.Group("/agencies")
		{
			agencies.POST("", agencyHandler.CreateAgency)
			agencies.GET("", agencyHandler.ListAgencies)
			agencies.GET("/:id", agencyHandler.GetAgency)
			agencies.PUT("/:id", agencyHandler.UpdateAgency)
			agencies.DELETE("/:id", agencyHandler.DeleteAgency)
		}

		terms := v1.Group("/terms")
		{
			terms.GET("/service", termHandler.ShowTermsOfService)
			terms.GET("/privacy", termHandler.ShowPrivacyPolicy)
		}

		events := v1.Group("/events")
		{
			events.POST("", eventHandler.CreateEvent)                               // イベント作成
			events.GET("", eventHandler.ListEvents)                                 // イベント一覧取得（検索機能付き）
			events.GET("/upcoming", eventHandler.GetUpcomingEvents)                 // 今後のイベント取得
			events.GET("/:id", eventHandler.GetEvent)                               // イベント詳細取得
			events.PUT("/:id", eventHandler.UpdateEvent)                            // イベント更新
			events.DELETE("/:id", eventHandler.DeleteEvent)                         // イベント削除
			events.POST("/:id/performers", eventHandler.AddPerformer)               // パフォーマー追加
			events.DELETE("/:id/performers/:performer_id", eventHandler.RemovePerformer) // パフォーマー削除
		}

		tags := v1.Group("/tags")
		{
			tags.POST("", tagHandler.CreateTag)       // タグ作成
			tags.GET("", tagHandler.ListTags)         // タグ一覧取得
			tags.GET("/:id", tagHandler.GetTag)       // タグ詳細取得
			tags.PUT("/:id", tagHandler.UpdateTag)    // タグ更新
			tags.DELETE("/:id", tagHandler.DeleteTag) // タグ削除
		}
	}

	// サーバー起動
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 サーバーを起動します (DDD architecture): http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}
