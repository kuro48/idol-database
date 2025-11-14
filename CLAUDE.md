# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

アイドル情報を管理するREST APIサーバー。Go 1.24.4、Gin Webフレームワーク、MongoDB v2を使用。

## 開発コマンド

### ビルドと実行
```bash
# アプリケーションの実行
go run main.go

# ビルド
go build -o idol-api main.go

# 依存関係の更新
go mod tidy
```

### Dockerを使用した開発
```bash
# MongoDBコンテナの起動
docker-compose up -d

# MongoDBコンテナの停止
docker-compose down

# ボリューム含めて削除
docker-compose down -v
```

## アーキテクチャ

### 現在の構造
- **main.go**: エントリーポイント。MongoDBへの接続とGinサーバーの起動を行う
- **cmd/api/**: APIサーバーの実行可能ファイル用（将来的に使用予定）
- **internal/**: 内部パッケージ用（将来的に使用予定）

### MongoDB接続
- MongoDB Atlas（クラウド）を使用
- 接続文字列はmain.go:19に直接記述（**⚠️ 本番環境では環境変数化が必要**）
- MongoDB v2 driver使用（`go.mongodb.org/mongo-driver/v2`）

### サーバー設定
- デフォルトポート: 8081
- フレームワーク: Gin（`github.com/gin-gonic/gin`）

## 重要な注意事項

### セキュリティ
1. **main.go:19の認証情報**: MongoDB接続文字列にユーザー名とパスワードがハードコードされています。環境変数に移行する必要があります
2. Dockerの`MONGO_INITDB_ROOT_USERNAME/PASSWORD`は開発環境用の設定

### プロジェクト段階
このプロジェクトは初期段階で、以下が未実装です：
- ハンドラー層の分離
- リポジトリパターン
- ビジネスロジック層
- テストコード
- 環境変数管理
- エラーハンドリング戦略

将来的には`cmd/api`と`internal`ディレクトリを使用した標準的なGo Project Layoutへの移行が想定されます。
