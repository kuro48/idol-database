package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/group"
	"github.com/kuro48/idol-api/internal/application/idol"
	"github.com/kuro48/idol-api/internal/application/removal"
	"github.com/kuro48/idol-api/internal/config"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/infrastructure/persistence/mongodb"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
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

	// MongoDBインデックスの作成
	ctx := context.Background()
	if err := idolRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("⚠️  インデックス作成エラー（続行します）: %v", err)
	} else {
		log.Println("✅ MongoDBインデックスを作成しました")
	}

	// アプリケーション層: アプリケーションサービス
	idolAppService := idol.NewApplicationService(idolRepo)
	removalAppService := removal.NewApplicationService(removalRepo, idolRepo, groupRepo)
	groupAppService := group.NewApplicationService(groupRepo)

	// プレゼンテーション層: ハンドラー
	idolHandler := handlers.NewIdolHandler(idolAppService)
	removalHandler := handlers.NewRemovalHandler(removalAppService)
	groupHandler := handlers.NewGroupHandler(groupAppService)
	termHandler := handlers.NewTermHandler("./static")

	// Ginルーターのセットアップ（デフォルトミドルウェアなし）
	router := gin.New()

	// ミドルウェア設定（順序重要）
	router.Use(gin.Recovery())                   // パニック回復
	router.Use(middleware.Logger())              // 構造化ログ
	router.Use(middleware.ErrorHandler())        // エラーハンドリング

	// CORS設定
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
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

	v1 := router.Group("/api/v1")
	{
		idols := v1.Group("/idols")
		{
			idols.POST("", idolHandler.CreateIdol)       // 新規作成
			idols.GET("", idolHandler.ListIdols)         // 一覧取得
			idols.GET("/:id", idolHandler.GetIdol)       // 詳細取得
			idols.PUT("/:id", idolHandler.UpdateIdol)    // 更新
			idols.DELETE("/:id", idolHandler.DeleteIdol) // 削除
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

		terms := v1.Group("/terms")
		{
			terms.GET("/service", termHandler.ShowTermsOfService)
			terms.GET("/privacy", termHandler.ShowPrivacyPolicy)
		}
	}

	// サーバー起動
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 サーバーを起動します (DDD architecture): http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}
