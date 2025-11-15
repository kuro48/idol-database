package main

import (
    "fmt"
    "log"

    "github.com/gin-gonic/gin"
    "github.com/kuro48/idol-api/internal/config"
    "github.com/kuro48/idol-api/internal/infrastructure/database"
    "github.com/kuro48/idol-api/internal/infrastructure/repository"
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

    // リポジトリとハンドラーの初期化
    idolRepo := repository.NewIdolRepository(db.Database)
    idolHandler := handlers.NewIdolHandler(idolRepo)

    // Ginルーターのセットアップ
    router := gin.Default()

    // ヘルスチェックエンドポイント
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "message": "Idol API is running",
        })
    })

	v1 := router.Group("/api/v1")
	{
		idols := v1.Group("/idols")
		{
			idols.POST("", idolHandler.CreateIdol)      // 新規作成
            idols.GET("", idolHandler.GetIdols)         // 一覧取得
            idols.GET("/:id", idolHandler.GetIdol)      // 詳細取得
            idols.PUT("/:id", idolHandler.UpdateIdol)   // 更新
            idols.DELETE("/:id", idolHandler.DeleteIdol) // 削除
		}
	}
	
	// サーバー起動
    addr := fmt.Sprintf(":%s", cfg.ServerPort)
    fmt.Printf("🚀 サーバーを起動します: http://localhost%s\n", addr)
    if err := router.Run(addr); err != nil {
        log.Fatal("サーバー起動エラー:", err)
    }
}
