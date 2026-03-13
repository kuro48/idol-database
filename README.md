# Idol API

アイドル情報を蓄積・提供するデータベース API。複数のサービスが共通基盤として利用することを想定した API-as-a-Platform。

## アーキテクチャ

Clean Architecture（5 層構造）で実装。

```
internal/
├── domain/              # ドメイン層（外部依存ゼロ）
├── application/         # アプリケーション層（ユースケースのオーケストレーション）
├── usecase/             # UseCase 層（Input Port 定義・DTO 変換）
├── infrastructure/      # インフラ層（MongoDB 実装）
│   └── persistence/mongodb/
└── interface/           # インターフェース層（HTTP ハンドラー・ミドルウェア）
    ├── handlers/
    └── middleware/
cmd/api/
├── adapters/            # アダプター（Composition Root）
└── main.go
```

**依存方向**: `interface → usecase → application → domain ← infrastructure`

詳細: [`docs/clean-architecture/`](docs/clean-architecture/)

## 技術スタック

- **言語**: Go 1.24+
- **フレームワーク**: Gin
- **データベース**: MongoDB (go.mongodb.org/mongo-driver/v2)
- **ドキュメント**: Swagger UI (`/swagger/index.html`)

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

## API エンドポイント

Swagger UI: `http://localhost:8081/swagger/index.html`

> 注意: `GIN_MODE=release`（本番環境）では Swagger UI は無効になります。

### 認証

| スコープ | ヘッダー | 用途 |
|---------|---------|------|
| 読み取り | なし | GET 系エンドポイント |
| 書き込み | `Authorization: Bearer <WRITE_API_KEY>` | 作成・更新・削除 |
| 管理者 | `Authorization: Bearer <ADMIN_API_KEY>` | 管理エンドポイント |

### ヘルスチェック

```
GET /health
GET /health/live
GET /health/ready
```

### アイドル

```
GET    /api/v1/idols                  # 検索・一覧（ページネーション付き）
GET    /api/v1/idols/:id              # 詳細
POST   /api/v1/idols                  # 作成（write 権限）
PUT    /api/v1/idols/:id              # 更新（write 権限）
DELETE /api/v1/idols/:id              # 削除（write 権限）
PUT    /api/v1/idols/:id/social-links # SNS リンク更新（write 権限）
POST   /api/v1/idols/bulk             # 一括作成（write 権限）
```

アイドルは `name`・`aliases`（別名/旧名）・`birthdate`・`agency_id`・`tag_ids`・`social_links`・`external_ids` を持つ。別名でも検索可能。

### グループ

```
GET    /api/v1/groups      # 一覧（ページネーション付き）
GET    /api/v1/groups/:id  # 詳細
POST   /api/v1/groups      # 作成（write 権限）
PUT    /api/v1/groups/:id  # 更新（write 権限）
DELETE /api/v1/groups/:id  # 削除（write 権限）
```

### 事務所

```
GET    /api/v1/agencies      # 一覧（ページネーション付き）
GET    /api/v1/agencies/:id  # 詳細
POST   /api/v1/agencies      # 作成（write 権限）
PUT    /api/v1/agencies/:id  # 更新（write 権限）
DELETE /api/v1/agencies/:id  # 削除（write 権限）
```

### タグ

```
GET    /api/v1/tags      # 検索・一覧
GET    /api/v1/tags/:id  # 詳細
POST   /api/v1/tags      # 作成（write 権限）
PUT    /api/v1/tags/:id  # 更新（write 権限）
DELETE /api/v1/tags/:id  # 削除（write 権限）
```

### イベント

```
GET    /api/v1/events                              # 検索・一覧
GET    /api/v1/events/upcoming                     # 今後のイベント
GET    /api/v1/events/:id                          # 詳細
POST   /api/v1/events                              # 作成（write 権限）
PUT    /api/v1/events/:id                          # 更新（write 権限）
DELETE /api/v1/events/:id                          # 削除（write 権限）
POST   /api/v1/events/:id/performers               # パフォーマー追加（write 権限）
DELETE /api/v1/events/:id/performers/:performer_id # パフォーマー削除（write 権限）
```

### 削除申請

```
POST /api/v1/removal-requests        # 申請作成
GET  /api/v1/removal-requests/:id    # 詳細
GET  /api/v1/removal-requests        # 一覧（admin）
GET  /api/v1/removal-requests/pending # 未処理一覧（admin）
PUT  /api/v1/removal-requests/:id    # ステータス更新（admin）
```

### Webhook

```
POST   /api/v1/admin/webhooks                      # サブスクリプション作成（admin）
GET    /api/v1/admin/webhooks                      # 一覧（admin）
GET    /api/v1/admin/webhooks/:id                  # 詳細（admin）
DELETE /api/v1/admin/webhooks/:id                  # 削除（admin）
POST   /api/v1/webhooks/receive/:subscription_id   # Webhook 受信
```

### エクスポート（admin）

```
GET /api/v1/admin/export/idols  # 全アイドルデータをエクスポート（JSON/JSONL、レート制限あり）
GET /api/v1/admin/export/logs   # エクスポート実行履歴
```

### 非同期ジョブ（admin）

```
POST /api/v1/admin/jobs/bulk-import  # バルクインポートジョブをキュー投入
GET  /api/v1/admin/jobs/:id          # ジョブ状態・結果取得
POST /api/v1/admin/jobs/:id/retry    # 失敗ジョブの再実行
```

### API 利用分析（admin）

```
GET /api/v1/admin/analytics/usage  # APIキー単位の利用統計（?days=7、最大 90 日）
```

### 利用規約

```
GET /api/v1/terms/service  # 利用規約
GET /api/v1/terms/privacy  # プライバシーポリシー
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

# Swagger docs 再生成
swag init -g cmd/api/main.go -o docs
```

新機能追加の手順は [`docs/DEVELOPMENT_GUIDE.md`](docs/DEVELOPMENT_GUIDE.md) を参照。

## ドキュメント

| ドキュメント | 内容 |
|------------|------|
| [Clean Architecture](docs/clean-architecture/) | レイヤー定義・境界ルール・ADR |
| [Development Guide](docs/DEVELOPMENT_GUIDE.md) | 開発手順・規約 |
| [Search API Spec](docs/search-api-spec.md) | 検索パラメーター仕様 |
| [Error Codes](docs/error-codes.md) | エラーコード一覧 |
| [API Versioning](docs/api-versioning-policy.md) | バージョニング方針 |
| [Removal Flow](docs/removal-request-flow.md) | 削除申請フロー |
| [Docker Guide](docs/docker-guide.md) | Docker 環境構築 |
| [Deployment](docs/deployment.md) | デプロイ手順 |
| [Legal Guidelines](docs/legal-guidelines.md) | 法的対応方針 |
