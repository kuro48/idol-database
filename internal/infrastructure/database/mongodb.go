package database

import (
	"context"
	"fmt"
	"log/slog"
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

	slog.Info("MongoDBに接続を試みています", "database", dbName)

	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetMaxPoolSize(200).
		SetMinPoolSize(10)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("MongoDB接続エラー: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("MongoDB Pingエラー: %w", err)
	}

	slog.Info("MongoDBに接続しました", "database", dbName)

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

// Ping はMongoDBへの疎通確認を行う
func (m *MongoDB) Ping(ctx context.Context) error {
    pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return m.Client.Ping(pingCtx, readpref.Primary())
}