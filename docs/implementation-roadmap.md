# 実装ロードマップ

## 📋 全体戦略

```
Phase 1 (MVP)    → Phase 2 (成長期)  → Phase 3 (交渉準備) → Phase 4 (公式化)
   0-3ヶ月            3-12ヶ月             6-12ヶ月             12ヶ月〜
```

**段階的アプローチ**:
1. ユーザー投稿型で法的リスクを最小化
2. 実績とユーザー基盤を構築
3. データを武器に事務所と交渉
4. 公式パートナーとして成長

---

## Phase 1: MVP（0-3ヶ月）

### 現在の状況
- ✅ DDD構造での基本CRUD実装完了
- ✅ MongoDB接続・基本インフラ完成
- 🚧 法的保護機能を実装中

### 目標
- **技術検証**: DDD構造の動作確認
- **法的基盤**: プロバイダ責任制限法に基づく設計
- **最小限の機能**: 投稿・削除・モデレーション

---

### Week 1-2: 法的保護機能の実装

#### Task 1: 削除申請機能（3日）

**実装内容**:
```go
// internal/domain/removal/
├── removal_request.go     // エンティティ
├── repository.go          // リポジトリIF
└── service.go             // ドメインサービス

// internal/application/removal/
├── command.go             // コマンドDTO
├── query.go               // クエリDTO
└── service.go             // アプリケーションサービス

// internal/infrastructure/persistence/mongodb/
└── removal_repository.go  // リポジトリ実装

// internal/interface/handlers/
└── removal_handler.go     // HTTPハンドラー
```

**エンドポイント**:
```
POST   /api/v1/removal-requests      # 削除申請
GET    /api/v1/removal-requests      # 申請一覧（管理者）
GET    /api/v1/removal-requests/:id  # 申請詳細
PUT    /api/v1/removal-requests/:id  # ステータス更新
```

#### Task 2: 利用規約・プライバシーポリシー（2日）

**実装内容**:
```go
// internal/interface/handlers/
└── legal_handler.go

// エンドポイント
GET /api/v1/legal/terms              # 利用規約
GET /api/v1/legal/privacy            # プライバシーポリシー
GET /api/v1/legal/posting-guidelines # 投稿ガイドライン
```

**静的ファイル**:
```
docs/legal/
├── terms.md                  # 利用規約（日本語）
├── privacy.md                # プライバシーポリシー
└── posting-guidelines.md     # 投稿ガイドライン
```

#### Task 3: モデレーション機能（3日）

**ドメイン拡張**:
```go
// internal/domain/idol/idol.go に追加
type ModerationStatus string

const (
    StatusPending  ModerationStatus = "pending"   // 承認待ち
    StatusApproved ModerationStatus = "approved"  // 承認済み
    StatusRejected ModerationStatus = "rejected"  // 却下
    StatusFlagged  ModerationStatus = "flagged"   // 要確認
)

type Idol struct {
    // 既存フィールド
    moderationStatus ModerationStatus
    createdBy        string  // ユーザーID（将来のため）
    lastEditedBy     string
    flags            int     // 通報カウント
}
```

**エンドポイント**:
```
POST   /api/v1/idols/:id/flag        # 通報
GET    /api/v1/moderation/pending    # 承認待ち一覧（管理者）
PUT    /api/v1/moderation/:id/approve # 承認
PUT    /api/v1/moderation/:id/reject  # 却下
```

---

### Week 3-4: データ品質向上

#### Task 4: バリデーション強化（2日）

**値オブジェクトの強化**:
```go
// internal/domain/idol/value_object.go に追加

// ImageURL のバリデーション強化
func NewImageURL(value string) (ImageURL, error) {
    if value == "" {
        return ImageURL{}, nil // 空は許可
    }

    // 外部URLのみ許可（直接アップロード禁止）
    if !isValidExternalURL(value) {
        return ImageURL{}, errors.New("外部URLのみ許可されています")
    }

    // 公式サイトorSNSのURLを推奨
    if !isOfficialSource(value) {
        // 警告は出すが許可（ログに記録）
        logWarning("非公式ソースのURL: %s", value)
    }

    return ImageURL{value: value}, nil
}

func isOfficialSource(url string) bool {
    officialDomains := []string{
        "twitter.com",
        "instagram.com",
        "facebook.com",
        "youtube.com",
        // 事務所公式サイトのドメインを追加
    }
    // ドメインチェックロジック
}
```

#### Task 5: 編集履歴機能（3日）

**新ドメイン**:
```go
// internal/domain/history/
├── edit_history.go        // エンティティ
└── repository.go          // リポジトリIF

type EditHistory struct {
    ID          HistoryID
    IdolID      idol.IdolID
    Version     int
    EditedBy    string
    EditedAt    time.Time
    Changes     []Change
    Reason      string
}

type Change struct {
    Field    string
    OldValue string
    NewValue string
}
```

**エンドポイント**:
```
GET /api/v1/idols/:id/history        # 編集履歴一覧
GET /api/v1/idols/:id/history/:version # 特定バージョン
POST /api/v1/idols/:id/revert/:version # 巻き戻し
```

#### Task 6: 検索機能の実装（3日）

**検索用リポジトリ拡張**:
```go
// internal/domain/idol/repository.go に追加

type SearchCriteria struct {
    Name         string
    Group        string
    Nationality  string
    MinAge       *int
    MaxAge       *int
    Status       []ModerationStatus
    SortBy       string // "name", "created_at", "updated_at"
    SortOrder    string // "asc", "desc"
    Limit        int
    Offset       int
}

SearchIdols(ctx context.Context, criteria SearchCriteria) ([]*Idol, int, error)
```

**エンドポイント**:
```
GET /api/v1/idols/search?name=山田&group=グループA&sort=name
```

---

### Week 5-6: テスト・ドキュメント整備

#### Task 7: テストの実装（4日）

```bash
# ドメイン層のテスト
internal/domain/idol/*_test.go
internal/domain/removal/*_test.go
internal/domain/history/*_test.go

# アプリケーション層のテスト
internal/application/idol/*_test.go
internal/application/removal/*_test.go

# インフラ層のテスト
internal/infrastructure/persistence/mongodb/*_test.go

# ハンドラーのテスト
internal/interface/handlers/*_test.go
```

**目標カバレッジ**: 70%以上

#### Task 8: API ドキュメント（2日）

**OpenAPI仕様書の作成**:
```yaml
# docs/openapi.yaml
openapi: 3.0.0
info:
  title: Idol API
  version: 1.0.0
  description: ユーザー投稿型アイドル情報データベース

paths:
  /api/v1/idols:
    get:
      summary: アイドル一覧取得
      parameters:
        - name: name
          in: query
          schema:
            type: string
    post:
      summary: アイドル作成
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateIdolRequest'
```

#### Task 9: デプロイ準備（2日）

```bash
# Dockerfile作成
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o idol-api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/idol-api .
CMD ["./idol-api"]
```

**docker-compose更新**:
```yaml
services:
  app:
    build: .
    ports:
      - "8081:8081"
    environment:
      - MONGODB_URI=mongodb://mongo:27017
    depends_on:
      - mongo

  mongo:
    image: mongo:7
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
```

---

### Phase 1 完了基準

```yaml
機能:
  ✅ 基本CRUD操作
  ✅ 削除申請フォーム
  ✅ モデレーション機能
  ✅ 編集履歴
  ✅ 検索機能

法的対応:
  ✅ 利用規約
  ✅ プライバシーポリシー
  ✅ 投稿ガイドライン
  ✅ 24時間以内削除対応フロー

品質:
  ✅ テストカバレッジ70%以上
  ✅ API ドキュメント完備
  ✅ Docker化完了
```

---

## Phase 2: 成長期（3-12ヶ月）

### 目標
- **ユーザー獲得**: MAU 1万人以上
- **データ充実**: アイドル登録数 500名以上
- **品質向上**: 自動モデレーション実装

### 主要機能

#### 1. ユーザー認証・権限管理（2週間）
```go
// internal/domain/user/
├── user.go           // ユーザーエンティティ
├── role.go           // 役割（admin, moderator, user）
└── repository.go

// 認証機能
- JWT トークン
- リフレッシュトークン
- OAuth（Google, Twitter）
```

#### 2. 通報機能（1週間）
```go
// internal/domain/report/
└── report.go

type ReportType string
const (
    TypeCopyright   ReportType = "copyright"
    TypeDefamation  ReportType = "defamation"
    TypeFalseInfo   ReportType = "false_info"
    TypePrivacy     ReportType = "privacy"
)
```

#### 3. 自動モデレーション（2週間）
```go
// internal/infrastructure/moderation/
├── profanity_filter.go    // NGワードフィルター
├── url_validator.go       // URL妥当性チェック
└── ai_moderator.go        // AI モデレーション（将来）
```

#### 4. 分析ダッシュボード（2週間）
```go
// internal/application/analytics/
└── service.go

// 提供データ
- アイドル別アクセス数
- 事務所別ランキング
- ユーザー属性分析
- 検索キーワードランキング
```

#### 5. API公開（1週間）
```go
// APIキー管理
// レート制限
// ドキュメント（Swagger UI）
```

---

## Phase 3: 交渉準備期（6-12ヶ月）

### 目標
- **実績構築**: 事務所交渉に必要なデータ収集
- **提案資料作成**: パートナーシップ提案書

### KPI収集機能

```go
// internal/application/kpi/
└── service.go

type AgencyKPI struct {
    AgencyName     string
    IdolCount      int
    TotalViews     int
    MonthlyViews   int
    UserDemographics map[string]interface{}
    TopIdols       []IdolRanking
}
```

### 提案書作成サポート
- 自動レポート生成
- グラフ・チャート作成
- PDFエクスポート

---

## Phase 4: パートナーシップ（12ヶ月〜）

### 公式データ統合

```go
// internal/domain/idol/idol.go に追加
type DataSource string

const (
    SourceUserContributed DataSource = "user_contributed"
    SourceOfficialVerified DataSource = "official_verified"
)

type Idol struct {
    // 既存フィールド
    dataSource    DataSource
    verifiedAt    *time.Time
    verifiedBy    string  // 事務所名
    agencyID      string  // 事務所ID
}
```

### 事務所管理機能

```go
// internal/domain/agency/
├── agency.go         // 事務所エンティティ
├── contract.go       // 契約情報
└── repository.go

// 機能
- 事務所アカウント
- 所属アイドル一括管理
- データ更新権限
- アクセス分析レポート
```

---

## 📊 マイルストーン

| Phase | 期間 | 主要成果物 | KPI |
|-------|------|-----------|-----|
| Phase 1 | 0-3ヶ月 | MVP公開 | MAU 100人 |
| Phase 2 | 3-12ヶ月 | ユーザー基盤 | MAU 1万人, アイドル 500名 |
| Phase 3 | 6-12ヶ月 | 交渉準備完了 | 提案書作成, 1社以上アプローチ |
| Phase 4 | 12ヶ月〜 | 公式化 | パートナー事務所 5社以上 |

---

## ⚠️ リスクと対策

### 技術リスク
- **MongoDB スケーラビリティ**: 早期にシャーディング設計
- **API レート制限**: Redis でキャッシュ・レート制限

### ビジネスリスク
- **法的クレーム**: 弁護士との顧問契約
- **競合出現**: 差別化ポイント（DDD品質、法的コンプライアンス）
- **事務所交渉失敗**: 複数社並行アプローチ

---

## 次のステップ

**今すぐ始めること**:
1. ✅ Phase 1 Week 1-2 のタスクを開始
2. 削除申請機能の実装
3. 利用規約・プライバシーポリシーの作成
4. モデレーション機能の追加
