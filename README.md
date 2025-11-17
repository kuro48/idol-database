# Idol API - ユーザー投稿型アイドル情報データベース

## 📋 プロジェクト概要

**ユーザー投稿型のアイドル情報プラットフォーム**を段階的に構築し、最終的に事務所公式データと連携するサービスです。

### ビジョン
- Phase 1-2: ユーザー投稿型プラットフォームとして基盤構築
- Phase 3-4: 事務所との公式パートナーシップ締結
- 最終目標: アイドル情報の信頼できる公式データソースとなる

### 現在のステータス
- ✅ DDD（ドメイン駆動設計）アーキテクチャで実装完了
- ✅ 基本CRUD操作実装済み
- 🚧 ユーザー投稿機能・法的保護機能を実装中

---

## 🎯 段階的ロードマップ

### Phase 1: MVP（0-3ヶ月）- **現在ここ**
**目標**: 技術検証 + 法的リスク最小化

実装内容:
- ✅ 基本CRUD操作（DDD構造）
- 🚧 ユーザー投稿機能
- 🚧 削除申請フォーム
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
├── domain/              # ドメイン層
│   └── idol/
│       ├── value_object.go  # 値オブジェクト
│       ├── idol.go          # エンティティ（Aggregate Root）
│       ├── repository.go    # リポジトリインターフェース
│       └── service.go       # ドメインサービス
├── application/         # アプリケーション層
│   └── idol/
│       ├── command.go       # コマンドDTO
│       ├── query.go         # クエリDTO
│       └── service.go       # アプリケーションサービス
├── infrastructure/      # インフラ層
│   ├── database/
│   │   └── mongodb.go       # DB接続
│   └── persistence/
│       └── mongodb/
│           └── idol_repository.go  # リポジトリ実装
└── interface/           # プレゼンテーション層
    └── handlers/
        └── idol_handler_ddd.go     # HTTPハンドラー
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
POST   /api/v1/idols      # 作成
GET    /api/v1/idols      # 一覧取得
GET    /api/v1/idols/:id  # 詳細取得
PUT    /api/v1/idols/:id  # 更新
DELETE /api/v1/idols/:id  # 削除
```

### リクエスト例
```bash
# アイドル作成
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

### 削除申請
```bash
POST /api/v1/removal-requests
```

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

---

## 📚 ドキュメント

- [実装ロードマップ](docs/implementation-roadmap.md) - 詳細な実装計画
- [法的ガイドライン](docs/legal-guidelines.md) - 法的リスクと対策
- [API仕様](docs/api-specification.md) - API詳細仕様
- [アーキテクチャ](docs/architecture.md) - DDD設計詳細

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

- **削除申請**: [準備中]
- **技術サポート**: [準備中]
- **事務所提携**: [準備中]
