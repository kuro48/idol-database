package database

import (
    "context"
    "fmt"
    "time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoDB はデータベース接続を保持
type MongoDB struct {
    Client   *mongo.Client
    Database *mongo.Database
}

// Connect はMongoDBに接続する
func Connect(uri, dbName string) (*MongoDB, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

	fmt.Println("🔄 MongoDBに接続を試みています...")
	fmt.Printf("📍 データベース名: %s\n", dbName)

	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

    // MongoDBに接続
    fmt.Println("⏳ クライアント接続中...")
    client, err := mongo.Connect(opts)
    if err != nil {
        return nil, fmt.Errorf("MongoDB接続エラー: %w", err)
    }

    // 接続確認（Pingを送信）
    fmt.Println("⏳ Ping送信中...")
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return nil, fmt.Errorf("MongoDB Pingエラー: %w", err)
    }

    fmt.Println("✅ MongoDBに接続しました")

    return &MongoDB{
        Client:   client,
        Database: client.Database(dbName),
    }, nil
}

// Close はデータベース接続を閉じる
func (m *MongoDB) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return m.Client.Disconnect(ctx)
}