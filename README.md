# Idol API

アイドル情報を蓄積・提供するデータベース API。複数のサービスが共通基盤として利用することを想定した API-as-a-Platform。

## アーキテクチャ

Clean Architecture（5 層構造）で実装。

```
internal/
├── domain/              # ドメイン層（外部依存ゼロ）
├── application/         # アプリケーション層（ユースケースのオーケストレーション）
├── usecase/             # UseCase 層（Input Port 定義・DTO 変換）
├── infrastructure/      # インフラ層（MongoDB 実装・アダプター）
│   ├── persistence/mongodb/
│   └── adapters/
└── interface/           # インターフェース層（HTTP ハンドラー・ミドルウェア）
    ├── handlers/
    └── middleware/
```

**依存方向**: `interface → usecase → application → domain ← infrastructure`

詳細: [`docs/clean-architecture/`](docs/clean-architecture/)

## 技術スタック

- **言語**: Go 1.24+
- **フレームワーク**: Gin
- **データベース**: MongoDB (go.mongodb.org/mongo-driver/v2)
- **API ドキュメント**: Swaggo / Swagger UI

## セットアップ

### 前提条件

- Go 1.24 以上
- Docker & Docker Compose

### 起動手順

```bash
# MongoDB 起動
docker-compose up -d

# 環境変数設定
cp .env.example .env.local
# .env.local を編集

# 起動
go run cmd/api/main.go
```

`.env.local` の設定例:

```env
MONGODB_URI=mongodb://admin:password@localhost:27017/?authSource=admin
MONGODB_DATABASE=idol_database
SERVER_PORT=8081
GIN_MODE=debug
WRITE_API_KEY=your-write-api-key
ADMIN_API_KEY=your-admin-api-key
```

## API ドキュメント

API 仕様は Swaggo を唯一の正本として管理します。エンドポイント、認証、リクエスト/レスポンス、クエリ仕様は Swagger UI または生成物を参照してください。

- Swagger UI: `http://localhost:8081/swagger/index.html` (`GIN_MODE=debug/test` のみ)
- OpenAPI YAML: `docs/swagger.yaml`
- OpenAPI JSON: `docs/swagger.json`

生成を更新する場合:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
```

## 開発

```bash
# テスト実行
go test ./...

# ビルド
go build -o idol-api cmd/api/main.go

# コード整形
go fmt ./...

# 依存関係整理
go mod tidy
```

新機能追加の手順は [`docs/DEVELOPMENT_GUIDE.md`](docs/DEVELOPMENT_GUIDE.md) を参照。

## ドキュメント

| ドキュメント | 内容 |
|------------|------|
| [Clean Architecture](docs/clean-architecture/) | レイヤー定義・境界ルール・ADR |
| [Development Guide](docs/DEVELOPMENT_GUIDE.md) | 開発手順・規約 |
| [Docker Guide](docs/docker-guide.md) | Docker 環境構築 |
| [Deployment](docs/deployment.md) | デプロイ手順 |
| [Legal Guidelines](docs/legal-guidelines.md) | 法的対応方針 |
