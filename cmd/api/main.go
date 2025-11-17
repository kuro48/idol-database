package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/application/group"
	"github.com/kuro48/idol-api/internal/application/idol"
	"github.com/kuro48/idol-api/internal/application/removal"
	"github.com/kuro48/idol-api/internal/config"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/infrastructure/persistence/mongodb"
	"github.com/kuro48/idol-api/internal/interface/handlers"
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

	// アプリケーション層: アプリケーションサービス
	idolAppService := idol.NewApplicationService(idolRepo)
	removalAppService := removal.NewApplicationService(removalRepo, idolRepo)
	groupAppService := group.NewApplicationService(groupRepo)

	// プレゼンテーション層: ハンドラー
	idolHandler := handlers.NewIdolHandler(idolAppService)
	removalHandler := handlers.NewRemovalHandler(removalAppService)
	groupHandler := handlers.NewGroupHandler(groupAppService)

	// Ginルーターのセットアップ
	router := gin.Default()

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
	}

	// サーバー起動
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("🚀 サーバーを起動します (DDD architecture): http://localhost%s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}
