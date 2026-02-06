# Idol API - ユーザー投稿型アイドル情報データベース

## 📋 プロジェクト概要

**ユーザー投稿型のアイドル情報プラットフォーム**を段階的に構築し、最終的に事務所公式データと連携するサービスです。

### ビジョン
- Phase 1-2: ユーザー投稿型プラットフォームとして基盤構築
- Phase 3-4: 事務所との公式パートナーシップ締結
- 最終目標: アイドル情報の信頼できる公式データソースとなる

### 現在のステータス
- ✅ DDD（ドメイン駆動設計）アーキテクチャで実装完了
- ✅ アイドル管理機能（基本CRUD操作）
- ✅ グループ管理機能
- ✅ 削除申請機能
- 🚧 利用規約・プライバシーポリシーの作成
- 🚧 モデレーション機能

---

## 🎯 段階的ロードマップ

### Phase 1: MVP（0-3ヶ月）- **現在ここ**
**目標**: 技術検証 + 法的リスク最小化

実装内容:
- ✅ 基本CRUD操作（DDD構造）
- ✅ アイドル管理機能
- ✅ グループ管理機能
- ✅ 削除申請機能
- 🚧 利用規約・プライバシーポリシー
- 🚧 モデレーション機能

法的保護:
- プロバイダ責任制限法に基づく設計
- 24時間以内の削除対応体制
- 画像は外部リンクのみ（直接ホスティングしない）

### Phase 2: 成長期（3-12ヶ月）
**目標**: ユーザー基盤確立 + データ品質向上

実装予定:
- ユーザー認証・権限管理
- 編集履歴・バージョン管理
- 通報機能
- 自動モデレーション
- 分析ダッシュボード

### Phase 3: 交渉準備期（6-12ヶ月）
**目標**: 事務所交渉のための実績構築

KPI収集:
- MAU（月間アクティブユーザー）
- 事務所別アクセス数
- ファン層分析データ
- 公式サイトへの送客実績

### Phase 4: パートナーシップ（12ヶ月〜）
**目標**: 公式データ獲得

戦略:
- 小規模事務所から開始
- 無料/有料プランの提供
- 段階的に大手事務所へ展開

---

## 🏗 アーキテクチャ

### DDD（ドメイン駆動設計）構造

```
internal/
├── config/              # 設定
├── domain/              # ドメイン層
│   ├── agency/
│   ├── event/
│   ├── group/
│   ├── idol/
│   ├── removal/
│   └── tag/
├── application/         # アプリケーション層
│   ├── agency/
│   ├── event/
│   ├── group/
│   ├── idol/
│   ├── removal/
│   └── tag/
├── usecase/             # ユースケース層
│   ├── agency/
│   ├── event/
│   ├── group/
│   ├── idol/
│   ├── removal/
│   └── tag/
├── infrastructure/      # インフラ層
│   ├── database/
│   │   └── mongodb.go         # DB接続
│   └── persistence/
│       └── mongodb/
│           ├── agency_repository.go
│           ├── event_repository.go
│           ├── group_repository.go
│           ├── idol_repository.go
│           ├── removal_repository.go
│           └── tag_repository.go
└── interface/           # プレゼンテーション層
    ├── handlers/
    │   ├── agency_handler.go
    │   ├── event_handler.go
    │   ├── group_handler.go
    │   ├── idol_handler.go
    │   ├── removal_handler.go
    │   ├── tag_handler.go
    │   └── term_handler.go
    └── middleware/
        ├── error.go
        ├── logger.go
        ├── ratelimit.go
        └── security.go
```

### 技術スタック
- **言語**: Go 1.24.4
- **Webフレームワーク**: Gin
- **データベース**: MongoDB v2
- **アーキテクチャ**: DDD（ドメイン駆動設計）

---

## 🚀 セットアップ

### 前提条件
- Go 1.24.4以上
- Docker & Docker Compose（MongoDB用）

### 1. リポジトリのクローン
```bash
git clone <repository-url>
cd idol-api
```

### 2. 環境変数の設定
```bash
cp .env.example .env.local
```

`.env.local` の内容:
```env
MONGODB_URI=mongodb://admin:password@localhost:27017/?authSource=admin
MONGODB_DATABASE=idol_database
SERVER_PORT=8081
GIN_MODE=debug
```

### 3. MongoDBの起動
```bash
docker-compose up -d
```

### 4. アプリケーションの起動
```bash
go run cmd/api/main.go
```

サーバーが http://localhost:8081 で起動します。

---

## 📡 API エンドポイント

### ヘルスチェック
```bash
GET /health
```

### アイドル管理
```bash
POST   /api/v1/idols      # アイドル作成
GET    /api/v1/idols      # アイドル一覧取得
GET    /api/v1/idols/:id  # アイドル詳細取得
PUT    /api/v1/idols/:id  # アイドル更新
DELETE /api/v1/idols/:id  # アイドル削除
```

### グループ管理
```bash
POST   /api/v1/groups      # グループ作成
GET    /api/v1/groups      # グループ一覧取得
GET    /api/v1/groups/:id  # グループ詳細取得
PUT    /api/v1/groups/:id  # グループ更新
DELETE /api/v1/groups/:id  # グループ削除
```

### 削除申請
```bash
POST   /api/v1/removal-requests           # 削除申請作成
GET    /api/v1/removal-requests           # 削除申請一覧取得
GET    /api/v1/removal-requests/pending   # 未処理の削除申請取得
GET    /api/v1/removal-requests/:id       # 削除申請詳細取得
PUT    /api/v1/removal-requests/:id       # 削除申請ステータス更新
```

---

## 📝 リクエスト例

### アイドル作成
```bash
curl -X POST http://localhost:8081/api/v1/idols \
  -H "Content-Type: application/json" \
  -d '{
    "name": "山田花子",
    "group": "Sample Group",
    "birthdate": "2000-05-15",
    "nationality": "日本",
    "image_url": "https://example.com/image.jpg"
  }'
```

### グループ作成
```bash
curl -X POST http://localhost:8081/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sample Group",
    "formation_date": "2015-04-01"
  }'
```

### 削除申請作成
```bash
curl -X POST http://localhost:8081/api/v1/removal-requests \
  -H "Content-Type: application/json" \
  -d '{
    "target_type": "idol",
    "target_id": "507f1f77bcf86cd799439011",
    "reason": "本人確認のため削除を希望します",
    "requester_email": "contact@example.com"
  }'
```

---

## ⚖️ 法的対応

### プロバイダ責任制限法に基づく設計
- **ユーザー投稿型**: 運営者はプラットフォーム提供者
- **迅速な削除対応**: 申請から24時間以内
- **透明性**: 利用規約・プライバシーポリシーの明示

### 禁止コンテンツ
- 著作権・肖像権侵害
- プライバシー侵害
- 虚偽情報・誹謗中傷
- 違法コンテンツ

### 削除申請の流れ
1. ユーザーが削除申請を提出
2. 申請が `pending` ステータスで保存
3. 管理者が申請を確認
4. 承認（`approved`）または却下（`rejected`）の判断
5. 承認された場合、該当コンテンツを削除

---

## 🧪 開発

### テスト実行
```bash
go test ./...
```

### ビルド
```bash
go build -o idol-api cmd/api/main.go
```

### コード整形
```bash
go fmt ./...
```

### 依存関係の更新
```bash
go mod tidy
```

---

## 📄 ライセンス

**Phase 1-2**: プロプライエタリ（開発中）
**Phase 3以降**: ユーザー投稿コンテンツは CC BY-SA 4.0 予定

---

## 🤝 貢献

現在は初期開発フェーズのため、外部貢献は受け付けていません。
Phase 2以降でコミュニティ貢献を開始予定です。

---

## 📞 お問い合わせ

- **削除申請**: `/api/v1/removal-requests` エンドポイントを使用
- **技術サポート**: [準備中]
- **事務所提携**: [準備中]
