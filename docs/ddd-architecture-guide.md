# DDDアーキテクチャ実装ガイド

このドキュメントは、idol-apiプロジェクトで採用しているDDD（ドメイン駆動設計）アーキテクチャの各ファイルの役割と実装パターンを説明します。

## 📚 目次

1. [DDD 4層構造の概要](#ddd-4層構造の概要)
2. [ドメイン層（Domain Layer）](#1-ドメイン層domain-layer)
3. [アプリケーション層（Application Layer）](#2-アプリケーション層application-layer)
4. [インフラ層（Infrastructure Layer）](#3-インフラ層infrastructure-layer)
5. [プレゼンテーション層（Interface Layer）](#4-プレゼンテーション層interfacepresentation-layer)
6. [実装フロー](#実装フロー全体像)
7. [重要な原則](#実装時の重要なポイント)

---

## DDD 4層構造の概要

```
internal/
├── domain/              # ドメイン層（ビジネスロジックの中核）
│   └── [bounded_context]/
│       ├── entity.go           # エンティティ
│       ├── value_object.go     # 値オブジェクト
│       ├── repository.go       # リポジトリIF
│       ├── service.go          # ドメインサービス
│       └── error.go            # ドメインエラー
│
├── usecase/             # ユースケース層（入力/出力・ユースケース実行）
│   └── [use_case]/
│       ├── command.go          # コマンド（入力）
│       ├── query.go            # クエリ/DTO（出力）
│       └── service.go          # ユースケース
│
├── application/         # アプリケーション層（ドメイン操作のオーケストレーション）
│   └── [use_case]/
│       └── service.go          # アプリケーションサービス
│
├── infrastructure/      # インフラ層（技術的詳細）
│   ├── database/
│   │   └── mongodb.go          # DB接続管理
│   └── persistence/
│       └── mongodb/
│           └── xxx_repository.go  # リポジトリ実装
│
└── interface/          # プレゼンテーション層（外部とのやり取り）
    └── handlers/
        └── xxx_handler.go      # HTTPハンドラー
```

---

## 1. ドメイン層（Domain Layer）

**責務**: ビジネスロジックとビジネスルールの実装。技術的詳細から完全に独立。

### 1-1. エンティティ（Entity）

**ファイル名**: `internal/domain/removal/removal_request.go`

**役割**:
- ビジネス上の「もの」を表現
- 一意のIDを持つ
- ライフサイクル全体で同一性が保たれる
- ビジネスルールをメソッドとして実装

**実装パターン**:

```go
// RemovalRequest は削除申請のエンティティ（Aggregate Root）
type RemovalRequest struct {
    // フィールドは小文字（外部から直接変更不可）
    id          RemovalID
    idolID      idol.IdolID
    requester   Requester
    reason      RemovalReason
    status      RemovalStatus
    createdAt   time.Time
    updatedAt   time.Time
}

// コンストラクタ: 新規作成
func NewRemovalRequest(
    idolID idol.IdolID,
    requester Requester,
    reason RemovalReason,
    // ...
) *RemovalRequest {
    now := time.Now()
    return &RemovalRequest{
        idolID:    idolID,
        requester: requester,
        status:    StatusPending, // 初期状態
        createdAt: now,
        updatedAt: now,
    }
}

// 再構築: 永続化データからの復元
func Reconstruct(
    id RemovalID,
    idolID idol.IdolID,
    // ... 全フィールド
) *RemovalRequest {
    return &RemovalRequest{
        id:     id,
        idolID: idolID,
        // ...
    }
}

// Getter: 外部からの読み取り
func (r *RemovalRequest) ID() RemovalID {
    return r.id
}

// ビジネスロジック: メソッドとして実装
func (r *RemovalRequest) Approve() error {
    // ビジネスルール: 保留中のみ承認可能
    if r.status != StatusPending {
        return NewDomainError("承認できるのは保留中の申請のみです")
    }
    r.status = StatusApproved
    r.updatedAt = time.Now()
    return nil
}

func (r *RemovalRequest) Reject() error {
    if r.status != StatusPending {
        return NewDomainError("却下できるのは保留中の申請のみです")
    }
    r.status = StatusRejected
    r.updatedAt = time.Now()
    return nil
}
```

**重要なポイント**:
- フィールドは小文字で定義（カプセル化）
- Getter/Setterで外部アクセスを制御
- ビジネスルールはメソッドで実装
- 技術的詳細（DB、HTTPなど）への依存なし

---

### 1-2. 値オブジェクト（Value Object）

**ファイル名**: `internal/domain/removal/value_object.go`

**役割**:
- 概念を表現する不変オブジェクト
- 同一性ではなく「値」で比較される
- バリデーションロジックを内包

**実装パターン**:

```go
// RemovalReason は削除理由の値オブジェクト
type RemovalReason struct {
    value string
}

// コンストラクタでバリデーション
func NewRemovalReason(value string) (RemovalReason, error) {
    // 必須チェック
    if value == "" {
        return RemovalReason{}, errors.New("削除理由は必須です")
    }

    // 長さチェック
    if len(value) < 10 {
        return RemovalReason{}, errors.New("削除理由は10文字以上で入力してください")
    }

    if len(value) > 1000 {
        return RemovalReason{}, errors.New("削除理由は1000文字以内で入力してください")
    }

    return RemovalReason{value: value}, nil
}

// Getterのみ（不変）
func (r RemovalReason) Value() string {
    return r.value
}
```

**列挙型の値オブジェクト**:

```go
// RemovalStatus は削除申請のステータス
type RemovalStatus string

const (
    StatusPending  RemovalStatus = "pending"
    StatusApproved RemovalStatus = "approved"
    StatusRejected RemovalStatus = "rejected"
)

func NewRemovalStatus(status string) (RemovalStatus, error) {
    rs := RemovalStatus(status)
    switch rs {
    case StatusPending, StatusApproved, StatusRejected:
        return rs, nil
    default:
        return "", errors.New("無効なステータスです")
    }
}
```

**重要なポイント**:
- コンストラクタ（`NewXxx`）で必ずバリデーション
- 一度作成したら変更不可（不変性）
- プリミティブ型をラップして意味を持たせる

---

### 1-3. リポジトリインターフェース

**ファイル名**: `internal/domain/removal/repository.go`

**役割**:
- データ永続化の抽象インターフェース
- 具体的な実装（MongoDB、PostgreSQLなど）は隠蔽

**実装パターン**:

```go
// Repository は削除申請リポジトリのインターフェース
type Repository interface {
    // 基本CRUD
    Save(ctx context.Context, request *RemovalRequest) error
    FindByID(ctx context.Context, id RemovalID) (*RemovalRequest, error)
    FindAll(ctx context.Context) ([]*RemovalRequest, error)
    Update(ctx context.Context, request *RemovalRequest) error
    Delete(ctx context.Context, id RemovalID) error

    // カスタムクエリ
    FindPending(ctx context.Context) ([]*RemovalRequest, error)
}
```

**重要なポイント**:
- ドメイン層は「何ができるか」だけ定義
- 「どう実装するか」はインフラ層で実装
- 依存性逆転の原則（DIP）の実現

---

### 1-4. ドメインエラー

**ファイル名**: `internal/domain/removal/error.go`

**役割**:
- ドメイン固有のエラーを定義

**実装パターン**:

```go
// DomainError はドメイン層のエラー
type DomainError struct {
    message string
}

func NewDomainError(message string) *DomainError {
    return &DomainError{message: message}
}

func (e *DomainError) Error() string {
    return e.message
}
```

---

## 2. ユースケース層（Usecase Layer）

**責務**: ユースケース（業務フロー）の実行。入力/出力の変換とアプリケーションサービスの呼び出しを担当する。

### 2-1. コマンド（Command）

**ファイル名**: `internal/usecase/removal/command.go`

**役割**:
- 外部からの入力データを表現
- HTTPリクエスト → コマンド変換

**実装パターン**:

```go
// CreateRemovalRequestCommand は削除申請作成コマンド
type CreateRemovalRequestCommand struct {
    IdolID      string `json:"idol_id" binding:"required"`
    Requester   string `json:"requester" binding:"required"`
    Reason      string `json:"reason" binding:"required"`
    ContactInfo string `json:"contact_info" binding:"required,email"`
    Evidence    string `json:"evidence"`
    Description string `json:"description" binding:"required"`
}

// UpdateStatusCommand はステータス更新コマンド
type UpdateStatusCommand struct {
    ID     string `json:"id" binding:"required"`
    Status string `json:"status" binding:"required,oneof=approved rejected"`
}
```

**重要なポイント**:
- `binding`タグでGinのバリデーション利用
- ドメインオブジェクトへの変換前の生データ
- プリミティブ型（string, intなど）を使用

---

### 2-2. クエリ/DTO（Data Transfer Object）

**ファイル名**: `internal/usecase/removal/query.go`

**役割**:
- 外部への出力データを表現
- ドメインオブジェクト → DTO変換

**実装パターン**:

```go
// RemovalRequestDTO は削除申請のデータ転送オブジェクト
type RemovalRequestDTO struct {
    ID          string    `json:"id"`
    IdolID      string    `json:"idol_id"`
    Requester   string    `json:"requester"`
    Reason      string    `json:"reason"`
    ContactInfo string    `json:"contact_info"`
    Evidence    string    `json:"evidence,omitempty"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**重要なポイント**:
- ドメインモデルの公開用表現
- 値オブジェクトの`Value()`を展開
- JSON化やAPI応答に適した形式

---

### 2-3. アプリケーションサービス（Application Service）

**ファイル名**: `internal/application/removal/service.go`

**役割**:
- ドメインオブジェクトを組み合わせた処理のオーケストレーション
- トランザクション管理
- ドメインオブジェクトの組み立て

**実装パターン**:

```go
type ApplicationService struct {
    removalRepo removal.Repository
    idolRepo    idol.Repository
}

func NewApplicationService(
    removalRepo removal.Repository,
    idolRepo idol.Repository,
) *ApplicationService {
    return &ApplicationService{
        removalRepo: removalRepo,
        idolRepo:    idolRepo,
    }
}

// CreateRemovalRequest は新しい削除申請を作成する
func (s *ApplicationService) CreateRemovalRequest(
    ctx context.Context,
    cmd CreateRemovalRequestCommand,
) (*RemovalRequestDTO, error) {
    // 1. バリデーション & ドメインオブジェクト作成
    idolID, err := idol.NewIdolID(cmd.IdolID)
    if err != nil {
        return nil, fmt.Errorf("無効なアイドルIDです: %w", err)
    }

    // 2. 関連データの存在確認
    _, err = s.idolRepo.FindByID(ctx, idolID)
    if err != nil {
        return nil, fmt.Errorf("指定されたアイドルが見つかりません: %w", err)
    }

    // 3. 値オブジェクトの組み立て
    requester, err := removal.NewRequester(cmd.Requester)
    if err != nil {
        return nil, fmt.Errorf("無効な申請者タイプです: %w", err)
    }

    reason, err := removal.NewRemovalReason(cmd.Reason)
    if err != nil {
        return nil, fmt.Errorf("削除理由が無効です: %w", err)
    }

    // ... 他の値オブジェクトも作成

    // 4. エンティティ作成
    request := removal.NewRemovalRequest(
        idolID,
        requester,
        reason,
        contactInfo,
        evidence,
        description,
    )

    // 5. 永続化
    if err := s.removalRepo.Save(ctx, request); err != nil {
        return nil, fmt.Errorf("削除申請の保存に失敗しました: %w", err)
    }

    // 6. DTOに変換して返却
    return toDTO(request), nil
}

// toDTO はエンティティをDTOに変換する
func toDTO(request *removal.RemovalRequest) *RemovalRequestDTO {
    return &RemovalRequestDTO{
        ID:          request.ID().Value(),
        IdolID:      request.IdolID().Value(),
        Requester:   string(request.Requester().Type()),
        Reason:      request.Reason().Value(),
        ContactInfo: request.ContactInfo().Value(),
        Evidence:    request.Evidence().Value(),
        Description: request.Description().Value(),
        Status:      string(request.Status()),
        CreatedAt:   request.CreatedAt(),
        UpdatedAt:   request.UpdatedAt(),
    }
}
```

**実行フロー**:
```
外部入力 → バリデーション → ドメインモデル構築 → 永続化 → DTO変換 → 外部出力
```

**重要なポイント**:
- ビジネスロジックはドメイン層に委譲
- 自身はフロー制御とトランザクション管理のみ
- エラーハンドリングとラップ

---

## 3. インフラ層（Infrastructure Layer）

**責務**: 技術的な詳細実装。DB、外部API、ファイルシステムなど。

### 3-1. リポジトリ実装

**ファイル名**: `internal/infrastructure/persistence/mongodb/removal_repository.go`

**役割**:
- ドメイン層のリポジトリインターフェースを実装
- ドメインモデル ↔ DB構造の変換

**実装パターン**:

```go
type RemovalRepository struct {
    collection *mongo.Collection
}

func NewRemovalRepository(db *mongo.Database) *RemovalRepository {
    return &RemovalRepository{
        collection: db.Collection("removal_requests"),
    }
}

// removalDocument はMongoDBに保存するドキュメント構造
type removalDocument struct {
    ID          bson.ObjectID `bson:"_id,omitempty"`
    IdolID      string        `bson:"idol_id"`
    Requester   string        `bson:"requester"`
    Reason      string        `bson:"reason"`
    ContactInfo string        `bson:"contact_info"`
    Evidence    string        `bson:"evidence,omitempty"`
    Description string        `bson:"description"`
    Status      string        `bson:"status"`
    CreatedAt   time.Time     `bson:"created_at"`
    UpdatedAt   time.Time     `bson:"updated_at"`
}

// toRemovalDocument: ドメインモデル → MongoDB構造
func toRemovalDocument(r *removal.RemovalRequest) *removalDocument {
    var objectID bson.ObjectID
    if r.ID().Value() != "" {
        objectID, _ = bson.ObjectIDFromHex(r.ID().Value())
    }

    return &removalDocument{
        ID:          objectID,
        IdolID:      r.IdolID().Value(),
        Requester:   string(r.Requester().Type()),
        Reason:      r.Reason().Value(),
        ContactInfo: r.ContactInfo().Value(),
        Evidence:    r.Evidence().Value(),
        Description: r.Description().Value(),
        Status:      string(r.Status()),
        CreatedAt:   r.CreatedAt(),
        UpdatedAt:   r.UpdatedAt(),
    }
}

// toRemovalDomain: MongoDB構造 → ドメインモデル
func toRemovalDomain(doc *removalDocument) (*removal.RemovalRequest, error) {
    id, err := removal.NewRemovalID(doc.ID.Hex())
    if err != nil {
        return nil, err
    }

    idolID, err := idol.NewIdolID(doc.IdolID)
    if err != nil {
        return nil, err
    }

    requester, err := removal.NewRequester(doc.Requester)
    if err != nil {
        return nil, err
    }

    reason, err := removal.NewRemovalReason(doc.Reason)
    if err != nil {
        return nil, err
    }

    // ... 他の値オブジェクトも再構築

    return removal.Reconstruct(
        id,
        idolID,
        requester,
        reason,
        contactInfo,
        evidence,
        description,
        status,
        doc.CreatedAt,
        doc.UpdatedAt,
    ), nil
}

// Save実装
func (r *RemovalRepository) Save(
    ctx context.Context,
    request *removal.RemovalRequest,
) error {
    doc := toRemovalDocument(request)

    // 新規作成の場合はIDを生成
    if doc.ID.IsZero() {
        doc.ID = bson.NewObjectID()
        doc.CreatedAt = time.Now()
        doc.UpdatedAt = time.Now()
    }

    _, err := r.collection.InsertOne(ctx, doc)
    if err != nil {
        return fmt.Errorf("削除申請の保存エラー: %w", err)
    }

    return nil
}

// FindByID実装
func (r *RemovalRepository) FindByID(
    ctx context.Context,
    id removal.RemovalID,
) (*removal.RemovalRequest, error) {
    objectID, err := bson.ObjectIDFromHex(id.Value())
    if err != nil {
        return nil, fmt.Errorf("無効なID形式: %w", err)
    }

    var doc removalDocument
    err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("削除申請が見つかりません")
        }
        return nil, fmt.Errorf("削除申請取得エラー: %w", err)
    }

    return toRemovalDomain(&doc)
}
```

**重要なポイント**:
- ドメインモデルとDB構造は別物として扱う
- 変換ロジック（`toDocument`, `toDomain`）を実装
- MongoDB固有のコード（`bson.ObjectID`など）はここだけ
- エラーハンドリングとラップ

---

## 4. プレゼンテーション層（Interface/Presentation Layer）

**責務**: 外部（HTTP、CLI、gRPCなど）とのやり取り。

### 4-1. HTTPハンドラー

**ファイル名**: `internal/interface/handlers/removal_handler.go`

**役割**:
- HTTPリクエストの受付
- アプリケーションサービスの呼び出し
- HTTPレスポンスの返却

**実装パターン**:

```go
type RemovalHandler struct {
    removalService *removal.ApplicationService
}

func NewRemovalHandler(
    removalService *removal.ApplicationService,
) *RemovalHandler {
    return &RemovalHandler{
        removalService: removalService,
    }
}

// CreateRemovalRequest は削除申請を作成する
// POST /api/v1/removal-requests
func (h *RemovalHandler) CreateRemovalRequest(c *gin.Context) {
    // 1. HTTPリクエスト → コマンドへのバインド
    var cmd removal.CreateRemovalRequestCommand
    if err := c.ShouldBindJSON(&cmd); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "リクエストが不正です",
            "details": err.Error(),
        })
        return
    }

    // 2. アプリケーションサービス呼び出し
    dto, err := h.removalService.CreateRemovalRequest(
        c.Request.Context(),
        cmd,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "削除申請の作成に失敗しました",
            "details": err.Error(),
        })
        return
    }

    // 3. HTTPレスポンス返却
    c.JSON(http.StatusCreated, dto)
}

// GetRemovalRequest は削除申請を取得する
// GET /api/v1/removal-requests/:id
func (h *RemovalHandler) GetRemovalRequest(c *gin.Context) {
    id := c.Param("id")

    dto, err := h.removalService.GetRemovalRequest(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error":   "削除申請が見つかりません",
            "details": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, dto)
}
```

**重要なポイント**:
- HTTP固有の処理のみ（リクエスト解析、ステータスコード設定）
- ビジネスロジックはアプリケーションサービスに委譲
- エラーハンドリングとHTTPレスポンス変換

---

### 4-2. 依存性注入とルーティング

**ファイル名**: `cmd/api/main.go`

**実装パターン**:

```go
func main() {
    // 設定の読み込み
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("設定読み込みエラー:", err)
    }

    // MongoDBに接続
    db, err := database.Connect(cfg.MongoDBURI, cfg.MongoDBDatabase)
    if err != nil {
        log.Fatal("データベース接続エラー:", err)
    }
    defer db.Close()

    // DDD構造での初期化
    // インフラ層: リポジトリ
    idolRepo := mongodb.NewIdolRepository(db.Database)
    removalRepo := mongodb.NewRemovalRepository(db.Database)

    // アプリケーション層: アプリケーションサービス
    idolAppService := idol.NewApplicationService(idolRepo)
    removalAppService := removal.NewApplicationService(removalRepo, idolRepo)

    // プレゼンテーション層: ハンドラー
    idolHandler := handlers.NewIdolHandlerDDD(idolAppService)
    removalHandler := handlers.NewRemovalHandler(removalAppService)

    // Ginルーターのセットアップ
    router := gin.Default()

    v1 := router.Group("/api/v1")
    {
        idols := v1.Group("/idols")
        {
            idols.POST("", idolHandler.CreateIdol)
            idols.GET("", idolHandler.ListIdols)
            idols.GET("/:id", idolHandler.GetIdol)
            idols.PUT("/:id", idolHandler.UpdateIdol)
            idols.DELETE("/:id", idolHandler.DeleteIdol)
        }

        removalRequests := v1.Group("/removal-requests")
        {
            removalRequests.POST("", removalHandler.CreateRemovalRequest)
            removalRequests.GET("", removalHandler.ListAllRemovalRequests)
            removalRequests.GET("/pending", removalHandler.ListPendingRemovalRequests)
            removalRequests.GET("/:id", removalHandler.GetRemovalRequest)
            removalRequests.PUT("/:id", removalHandler.UpdateStatus)
        }
    }

    // サーバー起動
    addr := fmt.Sprintf(":%s", cfg.ServerPort)
    if err := router.Run(addr); err != nil {
        log.Fatal("サーバー起動エラー:", err)
    }
}
```

---

## 実装フロー全体像

新機能を追加する際の実装順序:

### Step 1: ドメイン層から実装（技術非依存）

```bash
# 値オブジェクト
internal/domain/xxx/value_object.go
→ バリデーションロジック実装

# エンティティID
internal/domain/xxx/xxx_id.go
→ ID値オブジェクト実装

# エンティティ
internal/domain/xxx/entity.go
→ ビジネスロジック実装

# リポジトリインターフェース
internal/domain/xxx/repository.go
→ 必要なメソッドを定義

# ドメインエラー
internal/domain/xxx/error.go
→ ドメイン固有のエラー定義
```

### Step 2: アプリケーション層（ユースケース）

```bash
# コマンド/DTO
internal/application/xxx/command.go
internal/application/xxx/query.go

# アプリケーションサービス
internal/application/xxx/service.go
→ ドメインオブジェクトを組み合わせてユースケース実装
```

### Step 3: インフラ層（技術詳細）

```bash
# リポジトリ実装
internal/infrastructure/persistence/mongodb/xxx_repository.go
→ ドメインリポジトリIFを実装
→ ドメインモデル ↔ DB構造の変換
```

### Step 4: プレゼンテーション層（外部IF）

```bash
# ハンドラー
internal/interface/handlers/xxx_handler.go
→ HTTPリクエスト処理

# ルーティング
cmd/api/main.go
→ 依存性注入とルート設定
```

---

## 実装時の重要なポイント

### ✅ 依存の方向

```
プレゼンテーション層 ──→ アプリケーション層 ──→ ドメイン層
        ↓                      ↓
インフラ層 ─────────────────→ ドメイン層（IFのみ）
```

**原則**:
- **ドメイン層**: 他の層に依存しない（最も重要）
- **アプリケーション層**: ドメイン層のみ依存
- **インフラ層**: ドメイン層のインターフェースを実装
- **プレゼンテーション層**: アプリケーション層を呼び出し

### ✅ 各層の責務分離

| 層 | やること | やらないこと |
|----|---------|------------|
| ドメイン | ビジネスルール実装 | DB、HTTP、外部API操作 |
| アプリケーション | ユースケース実行・フロー制御 | ビジネスルール判断 |
| インフラ | DB/外部API実装 | ビジネスロジック |
| プレゼンテーション | HTTPリクエスト処理 | ビジネスロジック |

### ✅ コンストラクタパターン

```go
// ドメイン層: 新規作成
func NewRemovalRequest(...) *RemovalRequest

// ドメイン層: 永続化データからの復元
func Reconstruct(...) *RemovalRequest

// 値オブジェクト: バリデーション付き作成
func NewRemovalReason(value string) (RemovalReason, error)

// 値オブジェクト: ID生成
func NewRemovalID(value string) (RemovalID, error)
```

### ✅ エラーハンドリング

```go
// ドメイン層
return NewDomainError("ビジネスルール違反")

// アプリケーション層
if err != nil {
    return nil, fmt.Errorf("コンテキスト情報: %w", err)
}

// インフラ層
if err != nil {
    return fmt.Errorf("技術的詳細: %w", err)
}

// プレゼンテーション層
c.JSON(http.StatusBadRequest, gin.H{
    "error": "ユーザー向けメッセージ",
    "details": err.Error(),
})
```

---

## 新機能実装の例（モデレーション機能）

新しい機能を実装する場合の手順例:

### 1. ドメインモデルを考える

- **エンティティ**: `ModerationRequest`, `FlaggedContent`
- **値オブジェクト**: `ModerationStatus`, `FlagReason`
- **ビジネスルール**: 「3件以上の通報で自動フラグ」など

### 2. この順序で実装

```bash
# ドメイン層
internal/domain/moderation/value_object.go
internal/domain/moderation/moderation_id.go
internal/domain/moderation/moderation.go
internal/domain/moderation/repository.go
internal/domain/moderation/error.go

# アプリケーション層
internal/application/moderation/command.go
internal/application/moderation/query.go
internal/application/moderation/service.go

# インフラ層
internal/infrastructure/persistence/mongodb/moderation_repository.go

# プレゼンテーション層
internal/interface/handlers/moderation_handler.go

# 依存性注入
cmd/api/main.go
```

### 3. テストも同じ順序で

```bash
# ドメイン層のテストから書く（ビジネスロジック検証）
internal/domain/moderation/moderation_test.go
internal/domain/moderation/value_object_test.go

# アプリケーション層のテスト
internal/application/moderation/service_test.go

# インフラ層のテスト
internal/infrastructure/persistence/mongodb/moderation_repository_test.go
```

---

## まとめ

このDDD構造により、以下のメリットが得られます:

1. **ビジネスロジックの独立性**: 技術的詳細から分離され、変更に強い
2. **テスト可能性**: 各層を独立してテスト可能
3. **保守性**: 責務が明確で、コードの意図が理解しやすい
4. **拡張性**: 新機能追加時のパターンが明確
5. **技術的柔軟性**: DBやフレームワークの変更が容易

不明点があれば質問してください！
