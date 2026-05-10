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

## 技術スタック

- **言語**: Go 1.26.3+
- **フレームワーク**: Gin
- **データベース**: MongoDB (go.mongodb.org/mongo-driver/v2)
- **API ドキュメント**: Swaggo / Swagger UI

## セットアップ

### 前提条件

- Go 1.26.3 以上
- Docker & Docker Compose

### 起動手順

```bash
# MongoDB 起動
docker-compose up -d

# 環境変数設定
cp .env.example .env.local
# .env.local を編集

# 起動（backend ディレクトリで実行）
cd backend && go run cmd/api/main.go
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

### 環境変数リファレンス

<!-- AUTO-GENERATED from .env.example -->
| 変数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `MONGODB_URI` | Yes | — | MongoDB 接続 URI |
| `MONGODB_DATABASE` | Yes | — | 使用するデータベース名 |
| `MONGO_USERNAME` | Docker のみ | — | Docker Compose 用 MongoDB 認証ユーザー名 |
| `MONGO_PASSWORD` | Docker のみ | — | Docker Compose 用 MongoDB 認証パスワード |
| `SERVER_PORT` | No | `8081` | HTTP サーバーのリッスンポート |
| `GIN_MODE` | No | `release` | Gin 実行モード（`debug` / `release`） |
| `CORS_ALLOWED_ORIGINS` | No | — | 許可する CORS オリジン（カンマ区切り） |
| `WRITE_API_KEY` | Yes | — | write スコープ APIキー（POST/PUT/DELETE 用） |
| `ADMIN_API_KEY` | Yes | — | admin スコープ APIキー（32 文字以上推奨） |
| `OIDC_ISSUER` | No | — | OIDC Issuer URL（空の場合は APIキー認証のみ） |
| `OIDC_AUDIENCE` | No | — | OIDC リソースサーバー識別子 |
| `RATE_LIMIT_RPS` | No | `10` | グローバルレート制限（リクエスト/秒） |
| `RATE_LIMIT_BURST` | No | `20` | グローバルレート制限バースト許容数 |
| `PUBLIC_MUTATION_RATE_LIMIT_RPS` | No | `0.2` | 公開 POST 系追加レート制限（リクエスト/秒） |
| `PUBLIC_MUTATION_RATE_LIMIT_BURST` | No | `3` | 公開 POST 系バースト許容数 |
<!-- END AUTO-GENERATED -->

## API ドキュメント

API 仕様は Swaggo を唯一の正本として管理します。エンドポイント、認証、リクエスト/レスポンス、クエリ仕様は Swagger UI または生成物を参照してください。

- Swagger UI: `http://localhost:8081/swagger/index.html` (`GIN_MODE=debug/test` のみ)
- OpenAPI YAML: `backend/docs/swagger.yaml`
- OpenAPI JSON: `backend/docs/swagger.json`

## フロントエンド

APIサーバーは依存なしの静的フロントエンドも配信します。

- Data console: `http://localhost:8081/app`
- Admin dashboard: `http://localhost:8081/admin`

Data console では各一覧、情報投稿、削除申請、writeキーによるJSON登録ができます。Admin dashboard ではAPIキー発行、利用状況、投稿審査、削除申請、ジョブ確認ができます。

生成を更新する場合:

```bash
cd backend && go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs
```

## 開発

以下のコマンドはすべて `backend/` ディレクトリで実行してください。

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

API 仕様以外の補助資料は削除し、運用上参照する HTTP 仕様は Swagger に集約しています。

## リリース前チェック

本番公開時は `GIN_MODE=release` を設定し、`ADMIN_API_KEY` は32文字以上のランダム値にしてください。`WRITE_API_KEY` はコンテンツ更新用、外部read APIは `Authorization: Bearer <API key>` で発行済みAPIキーを必須にしています。free APIキーの自己発行ルートは公開しないため、APIキーは管理APIから発行してください。

公開してよい未認証エンドポイントは、ヘルスチェック、利用規約/プライバシーポリシー、投稿・削除申請の作成、外部Webhook受信に限定します。アイドル、グループ、事務所、イベント、リリース、タグのread APIは匿名アクセスを許可しません。

MongoDBインデックスは起動時に各リポジトリの `EnsureIndexes` で作成されます。本番データ投入前に `/health/ready` が200を返すこと、ログに各インデックス作成完了が出ていることを確認してください。

スモーク確認:

```bash
BASE_URL=https://api.example.com ./backend/scripts/smoke-read-auth.sh
BASE_URL=https://api.example.com API_KEY=ik_live_xxx ./backend/scripts/smoke-read-auth.sh
```
