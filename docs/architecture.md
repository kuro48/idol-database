# アーキテクチャ設計書

## アーキテクチャパターン

### レイヤードアーキテクチャ（Clean Architecture）

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │  ← Handler（Gin）
├─────────────────────────────────────────┤
│         Application Layer               │  ← Service（ビジネスロジック）
├─────────────────────────────────────────┤
│         Domain Layer                    │  ← Model（エンティティ）
├─────────────────────────────────────────┤
│         Infrastructure Layer            │  ← Repository（MongoDB）
└─────────────────────────────────────────┘
```

### 依存関係の方向
```
Handler → Service → Repository → MongoDB
  ↓         ↓          ↓
Model ← Interface ← Implementation
```

---

## ディレクトリ構造

```
idol-api/
├── cmd/
│   └── api/
│       └── main.go                    # エントリーポイント
│
├── internal/
│   ├── domain/                        # ドメイン層
│   │   ├── model/
│   │   │   ├── idol.go               # Idolエンティティ
│   │   │   ├── group.go              # Groupエンティティ
│   │   │   ├── submission.go         # Submissionエンティティ（Phase 2）
│   │   │   └── admin.go              # Adminエンティティ（Phase 3）
│   │   │
│   │   └── repository/                # リポジトリインターフェース
│   │       ├── idol_repository.go
│   │       ├── group_repository.go
│   │       ├── submission_repository.go
│   │       └── admin_repository.go
│   │
│   ├── usecase/                       # アプリケーション層（Service）
│   │   ├── idol_service.go
│   │   ├── group_service.go
│   │   ├── submission_service.go
│   │   └── admin_service.go
│   │
│   ├── interface/                     # インターフェース層
│   │   ├── handler/                   # HTTPハンドラー
│   │   │   ├── idol_handler.go
│   │   │   ├── group_handler.go
│   │   │   ├── submission_handler.go
│   │   │   └── response.go           # 共通レスポンス処理
│   │   │
│   │   ├── middleware/                # ミドルウェア
│   │   │   ├── auth.go               # 認証ミドルウェア
│   │   │   ├── cors.go               # CORS
│   │   │   ├── rate_limit.go         # レート制限
│   │   │   └── logger.go             # ロギング
│   │   │
│   │   └── validator/                 # バリデーター
│   │       ├── idol_validator.go
│   │       └── group_validator.go
│   │
│   ├── infrastructure/                # インフラ層
│   │   ├── database/
│   │   │   ├── mongodb.go            # MongoDB接続
│   │   │   └── migration.go          # インデックス作成
│   │   │
│   │   └── repository/                # リポジトリ実装
│   │       ├── idol_repository_impl.go
│   │       ├── group_repository_impl.go
│   │       ├── submission_repository_impl.go
│   │       └── admin_repository_impl.go
│   │
│   └── config/                        # 設定
│       ├── config.go                  # 設定構造体
│       └── env.go                     # 環境変数読み込み
│
├── pkg/                               # 公開パッケージ
│   ├── utils/
│   │   ├── string.go                 # 文字列ユーティリティ
│   │   ├── time.go                   # 時刻ユーティリティ
│   │   └── pagination.go             # ページネーション
│   │
│   └── errors/
│       └── errors.go                  # カスタムエラー定義
│
├── docs/                              # ドキュメント
│   ├── data-model.md
│   ├── api-specification.md
│   ├── architecture.md
│   └── implementation-roadmap.md
│
├── scripts/                           # スクリプト
│   ├── setup_indexes.sh              # インデックス作成
│   └── generate_api_key.sh           # API Key生成
│
├── .docker/
│   └── Dockerfile
│
├── docker-compose.yml
├── go.mod
├── go.sum
├── main.go                            # 後で削除（cmd/api/main.goに移行）
├── .env.example                       # 環境変数サンプル
├── .gitignore
├── CLAUDE.md
└── README.md
```

---

## 各層の責務

### 1. Presentation Layer（interface/handler）

**責務:**
- HTTPリクエストの受信とレスポンスの返却
- リクエストのバリデーション
- Serviceの呼び出し
- エラーハンドリングとステータスコード設定

**例: idol_handler.go**
```go
type IdolHandler struct {
    idolService usecase.IdolService
}

func (h *IdolHandler) GetIdols(c *gin.Context) {
    // 1. クエリパラメータの取得
    params := parseQueryParams(c)

    // 2. Serviceの呼び出し
    idols, pagination, err := h.idolService.GetIdols(c.Request.Context(), params)
    if err != nil {
        handleError(c, err)
        return
    }

    // 3. レスポンスの返却
    c.JSON(http.StatusOK, gin.H{
        "data": idols,
        "pagination": pagination,
    })
}
```

---

### 2. Application Layer（usecase）

**責務:**
- ビジネスロジックの実装
- トランザクション管理
- 複数Repositoryの協調
- データ整合性の担保

**例: idol_service.go**
```go
type IdolService interface {
    GetIdols(ctx context.Context, params QueryParams) ([]model.Idol, Pagination, error)
    GetIdolByID(ctx context.Context, id string) (*model.Idol, error)
    CreateIdol(ctx context.Context, idol *model.Idol) error
    UpdateIdol(ctx context.Context, id string, idol *model.Idol) error
    DeleteIdol(ctx context.Context, id string) error
}

type idolServiceImpl struct {
    idolRepo  repository.IdolRepository
    groupRepo repository.GroupRepository
}

func (s *idolServiceImpl) CreateIdol(ctx context.Context, idol *model.Idol) error {
    // 1. グループの存在確認
    for _, membership := range idol.GroupMemberships {
        exists, err := s.groupRepo.ExistsByID(ctx, membership.GroupID)
        if err != nil {
            return err
        }
        if !exists {
            return errors.New("group not found")
        }

        // 2. グループ名をキャッシュ
        group, err := s.groupRepo.GetByID(ctx, membership.GroupID)
        if err != nil {
            return err
        }
        membership.GroupName = group.Name
    }

    // 3. 重複チェック
    duplicate, err := s.idolRepo.FindDuplicate(ctx, idol.Name, idol.BirthDate)
    if err != nil {
        return err
    }
    if duplicate != nil {
        return errors.New("potential duplicate found")
    }

    // 4. is_active の自動計算
    idol.IsActive = idol.GraduationDate == nil

    // 5. 作成日時の設定
    now := time.Now()
    idol.CreatedAt = now
    idol.UpdatedAt = now

    // 6. 保存
    return s.idolRepo.Create(ctx, idol)
}
```

---

### 3. Domain Layer（domain/model & repository）

**責務:**
- エンティティの定義
- ドメインロジックのカプセル化
- リポジトリインターフェースの定義

**例: idol.go**
```go
package model

type Idol struct {
    ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
    Name            string              `json:"name" bson:"name" binding:"required"`
    NameKana        string              `json:"name_kana" bson:"name_kana" binding:"required"`
    // ... 他のフィールド
}

// ドメインロジック
func (i *Idol) CalculateAge() int {
    now := time.Now()
    age := now.Year() - i.BirthDate.Year()
    if now.YearDay() < i.BirthDate.YearDay() {
        age--
    }
    return age
}

func (i *Idol) UpdateIsActive() {
    i.IsActive = i.GraduationDate == nil
}
```

**例: idol_repository.go（インターフェース）**
```go
package repository

type IdolRepository interface {
    Create(ctx context.Context, idol *model.Idol) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*model.Idol, error)
    Find(ctx context.Context, filter IdolFilter) ([]*model.Idol, error)
    Update(ctx context.Context, id primitive.ObjectID, idol *model.Idol) error
    Delete(ctx context.Context, id primitive.ObjectID) error
    FindDuplicate(ctx context.Context, name string, birthDate time.Time) (*model.Idol, error)
    Count(ctx context.Context, filter IdolFilter) (int64, error)
}
```

---

### 4. Infrastructure Layer（infrastructure/repository）

**責務:**
- MongoDBとの実際の通信
- クエリの実装
- インデックスの管理

**例: idol_repository_impl.go**
```go
type idolRepositoryImpl struct {
    collection *mongo.Collection
}

func NewIdolRepository(db *mongo.Database) repository.IdolRepository {
    return &idolRepositoryImpl{
        collection: db.Collection("idols"),
    }
}

func (r *idolRepositoryImpl) Create(ctx context.Context, idol *model.Idol) error {
    result, err := r.collection.InsertOne(ctx, idol)
    if err != nil {
        return err
    }
    idol.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

func (r *idolRepositoryImpl) Find(ctx context.Context, filter IdolFilter) ([]*model.Idol, error) {
    // フィルターをMongoDBクエリに変換
    mongoFilter := bson.M{}

    if filter.Name != "" {
        mongoFilter["name"] = bson.M{"$regex": filter.Name, "$options": "i"}
    }

    if filter.IsActive != nil {
        mongoFilter["is_active"] = *filter.IsActive
    }

    if filter.GroupID != nil {
        mongoFilter["group_memberships.group_id"] = *filter.GroupID
    }

    // ソート
    opts := options.Find()
    if filter.Sort != "" {
        sortOrder := 1
        if filter.Order == "desc" {
            sortOrder = -1
        }
        opts.SetSort(bson.D{{Key: filter.Sort, Value: sortOrder}})
    }

    // ページネーション
    if filter.Page > 0 && filter.Limit > 0 {
        skip := (filter.Page - 1) * filter.Limit
        opts.SetSkip(int64(skip))
        opts.SetLimit(int64(filter.Limit))
    }

    cursor, err := r.collection.Find(ctx, mongoFilter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var idols []*model.Idol
    if err := cursor.All(ctx, &idols); err != nil {
        return nil, err
    }

    return idols, nil
}
```

---

## ミドルウェア設計

### 認証ミドルウェア（auth.go）

```go
func AuthMiddleware(apiKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        requestKey := c.GetHeader("X-API-Key")

        if requestKey == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{
                    "code": "UNAUTHORIZED",
                    "message": "API Key is required",
                },
            })
            c.Abort()
            return
        }

        if requestKey != apiKey {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": gin.H{
                    "code": "UNAUTHORIZED",
                    "message": "Invalid API Key",
                },
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### レート制限ミドルウェア（rate_limit.go）

```go
// Phase 1では簡易実装、Phase 2以降でRedis等を使用
func RateLimitMiddleware(maxRequests int, duration time.Duration) gin.HandlerFunc {
    // IP別のリクエストカウンター
    requestCounts := make(map[string]*rateLimitInfo)
    mu := &sync.RWMutex{}

    return func(c *gin.Context) {
        ip := c.ClientIP()

        mu.Lock()
        defer mu.Unlock()

        info, exists := requestCounts[ip]
        if !exists {
            info = &rateLimitInfo{
                count: 0,
                resetTime: time.Now().Add(duration),
            }
            requestCounts[ip] = info
        }

        if time.Now().After(info.resetTime) {
            info.count = 0
            info.resetTime = time.Now().Add(duration)
        }

        if info.count >= maxRequests {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": gin.H{
                    "code": "RATE_LIMIT_EXCEEDED",
                    "message": "Too many requests",
                },
            })
            c.Abort()
            return
        }

        info.count++
        c.Next()
    }
}
```

---

## エラーハンドリング戦略

### カスタムエラー定義（pkg/errors/errors.go）

```go
type AppError struct {
    Code       string
    Message    string
    StatusCode int
    Details    []ErrorDetail
}

type ErrorDetail struct {
    Field   string
    Message string
}

var (
    ErrNotFound = &AppError{
        Code:       "NOT_FOUND",
        Message:    "Resource not found",
        StatusCode: http.StatusNotFound,
    }

    ErrValidation = &AppError{
        Code:       "VALIDATION_ERROR",
        Message:    "Validation failed",
        StatusCode: http.StatusBadRequest,
    }

    ErrUnauthorized = &AppError{
        Code:       "UNAUTHORIZED",
        Message:    "Authentication required",
        StatusCode: http.StatusUnauthorized,
    }

    ErrDuplicate = &AppError{
        Code:       "DUPLICATE",
        Message:    "Resource already exists",
        StatusCode: http.StatusConflict,
    }
)
```

### エラーハンドリングヘルパー（interface/handler/response.go）

```go
func HandleError(c *gin.Context, err error) {
    var appErr *AppError

    if errors.As(err, &appErr) {
        c.JSON(appErr.StatusCode, gin.H{
            "error": gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
                "details": appErr.Details,
            },
        })
        return
    }

    // 予期しないエラー
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": gin.H{
            "code":    "INTERNAL_ERROR",
            "message": "An unexpected error occurred",
        },
    })
}
```

---

## 設定管理

### 環境変数（.env.example）

```env
# Server
PORT=8081
GIN_MODE=debug

# MongoDB
MONGODB_URI=mongodb://admin:password@localhost:27017
MONGODB_DATABASE=idol_api

# Authentication
API_KEY=your-secret-api-key-here

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=1h

# CORS
CORS_ALLOWED_ORIGINS=*
```

### 設定構造体（internal/config/config.go）

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    RateLimit RateLimitConfig
    CORS     CORSConfig
}

type ServerConfig struct {
    Port    string
    GinMode string
}

type DatabaseConfig struct {
    URI      string
    Database string
}

type AuthConfig struct {
    APIKey string
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        // .envファイルがない場合は環境変数から読み込む
    }

    return &Config{
        Server: ServerConfig{
            Port:    getEnv("PORT", "8081"),
            GinMode: getEnv("GIN_MODE", "debug"),
        },
        Database: DatabaseConfig{
            URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
            Database: getEnv("MONGODB_DATABASE", "idol_api"),
        },
        Auth: AuthConfig{
            APIKey: getEnv("API_KEY", ""),
        },
    }, nil
}
```

---

## ルーティング設計（cmd/api/main.go）

```go
func setupRouter(
    idolHandler *handler.IdolHandler,
    groupHandler *handler.GroupHandler,
    config *config.Config,
) *gin.Engine {
    r := gin.Default()

    // ミドルウェア
    r.Use(middleware.CORS(config.CORS))
    r.Use(middleware.Logger())

    // API v1
    v1 := r.Group("/api/v1")
    {
        // Public endpoints（認証不要）
        idols := v1.Group("/idols")
        {
            idols.GET("", idolHandler.GetIdols)
            idols.GET("/:id", idolHandler.GetIdolByID)
            idols.GET("/search", idolHandler.SearchIdols)
        }

        groups := v1.Group("/groups")
        {
            groups.GET("", groupHandler.GetGroups)
            groups.GET("/:id", groupHandler.GetGroupByID)
            groups.GET("/:id/members", groupHandler.GetGroupMembers)
        }

        // Protected endpoints（認証必須）
        auth := v1.Group("")
        auth.Use(middleware.AuthMiddleware(config.Auth.APIKey))
        auth.Use(middleware.RateLimitMiddleware(100, time.Hour))
        {
            auth.POST("/idols", idolHandler.CreateIdol)
            auth.PUT("/idols/:id", idolHandler.UpdateIdol)
            auth.PATCH("/idols/:id", idolHandler.PatchIdol)
            auth.DELETE("/idols/:id", idolHandler.DeleteIdol)

            auth.POST("/groups", groupHandler.CreateGroup)
            auth.PUT("/groups/:id", groupHandler.UpdateGroup)
            auth.PATCH("/groups/:id", groupHandler.PatchGroup)
            auth.DELETE("/groups/:id", groupHandler.DeleteGroup)
        }
    }

    return r
}
```

---

## テスト戦略

### ユニットテスト
- 各層ごとにモックを使用したテスト
- `testify/mock` を使用

### 統合テスト
- MongoDB Testcontainersを使用
- 実際のDBとの連携をテスト

### E2Eテスト
- HTTPリクエストベースのテスト
- `httptest` を使用

**ディレクトリ構造:**
```
internal/
├── usecase/
│   ├── idol_service.go
│   └── idol_service_test.go
├── infrastructure/
│   └── repository/
│       ├── idol_repository_impl.go
│       └── idol_repository_impl_test.go
```
