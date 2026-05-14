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
IDOL_AUTH_URL=https://auth.example.com
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
| `TRUSTED_PROXIES` | No | — | 信頼プロキシ CIDR（カンマ区切り、空=プロキシ信頼なし） |
| `IDOL_AUTH_URL` | release 必須 | — | idol-auth の公開 URL（例: `https://auth.example.com`） |
| `RATE_LIMIT_RPS` | No | `10` | グローバルレート制限（リクエスト/秒） |
| `RATE_LIMIT_BURST` | No | `20` | グローバルレート制限バースト許容数 |
| `PUBLIC_MUTATION_RATE_LIMIT_RPS` | No | `0.2` | 公開 POST 系追加レート制限（リクエスト/秒） |
| `PUBLIC_MUTATION_RATE_LIMIT_BURST` | No | `3` | 公開 POST 系バースト許容数 |
| `WEBHOOK_TIMEOUT_SECONDS` | No | `10` | Webhook HTTP クライアントタイムアウト（秒） |
| `SMTP_HOST` | No | — | SMTP ホスト（空の場合はメール通知無効） |
| `SMTP_PORT` | No | `587` | SMTP ポート |
| `SMTP_USERNAME` | No | — | SMTP ユーザー名 |
| `SMTP_PASSWORD` | No | — | SMTP パスワード |
| `SMTP_FROM` | No | — | 送信元メールアドレス |
| `SMTP_FROM_NAME` | No | `Idol API` | 送信元表示名 |
| `STRIPE_SECRET_KEY` | No | — | Stripe シークレットキー（空の場合は決済無効） |
| `STRIPE_WEBHOOK_SECRET` | Stripe 有効時必須 | — | Stripe Webhook 署名シークレット |
| `STRIPE_KEY_SEED_SECRET` | Stripe 有効時必須 | — | APIキー生成シークレット |
| `STRIPE_PRICE_DEVELOPER` | Stripe 有効時必須 | — | Developer プランの Stripe Price ID |
| `STRIPE_PRICE_BUSINESS` | Stripe 有効時必須 | — | Business プランの Stripe Price ID |
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

本番公開時は `GIN_MODE=release` を設定し、`IDOL_AUTH_URL` に idol-auth の公開 URL を指定してください。write/admin 操作は idol-auth が発行したアクセストークン（`admin` ロール必須）でのみ許可されます。

公開してよい未認証エンドポイントは、ヘルスチェック、利用規約/プライバシーポリシー、投稿・削除申請の作成、外部Webhook受信に限定します。アイドル、グループ、事務所、イベント、リリース、タグのread APIは匿名アクセスを許可しません。

MongoDBインデックスは起動時に各リポジトリの `EnsureIndexes` で作成されます。本番データ投入前に `/health/ready` が200を返すこと、ログに各インデックス作成完了が出ていることを確認してください。

スモーク確認:

```bash
BASE_URL=https://api.example.com ./backend/scripts/smoke-read-auth.sh
BASE_URL=https://api.example.com API_KEY=ik_live_xxx ./backend/scripts/smoke-read-auth.sh
```
