# Idol API - API-as-a-Platform 実装計画

**作成日**: 2025-11-26
**実装期間**: 3ヶ月（2025年12月〜2026年2月）
**ゴール**: 様々なWebサービス・アプリで利用可能な包括的アイドル情報APIの構築

---

## 🎯 ビジョン

このAPIを基盤として、以下のような多様なWebサービスが構築できるプラットフォームを目指す：

### 想定ユースケース
1. **📚 ファン向けデータベースサイト**
   - アイドル検索・プロフィール閲覧
   - グループ情報・所属メンバー確認
   - 事務所別アイドル一覧

2. **🎪 イベント・ライブ情報集約サイト**
   - イベント検索・カレンダー表示
   - 出演者から逆引き検索
   - 地域・期間フィルタリング

3. **👥 ファンコミュニティ形成促進サイト**
   - お気に入り・フォロー機能
   - アクティビティフィード
   - イベントレビュー・コメント

### API設計方針
- **柔軟性**: 複数条件での検索・フィルタリング対応
- **効率性**: ページネーション、関連データ一括取得
- **開発者体験**: 充実したドキュメント、一貫したエラーハンドリング
- **拡張性**: 段階的な機能追加、後方互換性の維持

---

## 📅 3ヶ月スケジュール概要

### **Month 1（12月）: 基盤強化**
- Week 1-2: 検索・フィルタリング基盤
- Week 3: ページネーション + メタデータ
- Week 4: 事務所（Agency）エンティティ

**成果物**: 柔軟な検索APIと構造化されたデータモデル

---

### **Month 2（1月）: データ拡張**
- Week 1-2: イベント/ライブエンティティ
- Week 3: 会場（Venue）管理
- Week 4: SNS/外部リンク管理

**成果物**: イベント集約サイト構築可能なAPI

---

### **Month 3（2月）: 開発者体験向上**
- Week 1-2: OpenAPI仕様書 + Swagger UI
- Week 3: タグ・カテゴリシステム
- Week 4: パフォーマンス最適化 + 総合テスト

**成果物**: プロダクション環境対応の包括的API

---

## 🔴 Phase 1: 検索・フィルタリング基盤（Week 1-2）

### 目的
どんなアプリでも必須となる「アイドルを探す」機能の実装

### 実装期間
**5-7営業日**

---

### Task 1.1: クエリパラメータ処理基盤

**工数**: 2日
**優先度**: 🔴 Critical

#### 実装内容

##### 1. リクエストパラメータDTO作成
```go
// internal/application/idol/query.go

type ListIdolsQuery struct {
    // 検索条件
    Name        *string `form:"name"`         // 部分一致検索
    Nationality *string `form:"nationality"`  // 完全一致
    GroupID     *string `form:"group_id"`     // グループIDフィルター
    AgencyID    *string `form:"agency_id"`    // 事務所IDフィルター（後で実装）

    // 年齢範囲
    AgeMin      *int    `form:"age_min"`
    AgeMax      *int    `form:"age_max"`

    // 生年月日範囲
    BirthdateFrom *string `form:"birthdate_from"` // YYYY-MM-DD
    BirthdateTo   *string `form:"birthdate_to"`   // YYYY-MM-DD

    // タグ（将来実装）
    Tags        []string `form:"tags"`

    // ソート
    Sort        *string `form:"sort"`   // name, birthdate, created_at
    Order       *string `form:"order"`  // asc, desc

    // ページネーション
    Page        *int    `form:"page"`
    Limit       *int    `form:"limit"`
}

// デフォルト値設定
func (q *ListIdolsQuery) ApplyDefaults() {
    if q.Page == nil || *q.Page < 1 {
        defaultPage := 1
        q.Page = &defaultPage
    }
    if q.Limit == nil || *q.Limit < 1 {
        defaultLimit := 20
        q.Limit = &defaultLimit
    }
    if q.Limit != nil && *q.Limit > 100 {
        maxLimit := 100
        q.Limit = &maxLimit
    }
    if q.Sort == nil {
        defaultSort := "created_at"
        q.Sort = &defaultSort
    }
    if q.Order == nil {
        defaultOrder := "desc"
        q.Order = &defaultOrder
    }
}

// バリデーション
func (q *ListIdolsQuery) Validate() error {
    if q.Sort != nil {
        allowedSorts := []string{"name", "birthdate", "created_at"}
        if !contains(allowedSorts, *q.Sort) {
            return errors.New("無効なソート項目です")
        }
    }
    if q.Order != nil {
        allowedOrders := []string{"asc", "desc"}
        if !contains(allowedOrders, *q.Order) {
            return errors.New("無効なソート順です")
        }
    }
    return nil
}
```

##### 2. リポジトリインターフェース拡張
```go
// internal/domain/idol/repository.go

type SearchCriteria struct {
    Name          *string
    Nationality   *string
    GroupID       *string
    AgeMin        *int
    AgeMax        *int
    BirthdateFrom *time.Time
    BirthdateTo   *time.Time

    Sort  string
    Order string

    Offset int
    Limit  int
}

type Repository interface {
    // 既存メソッド
    Save(ctx context.Context, idol *Idol) error
    FindByID(ctx context.Context, id IdolID) (*Idol, error)
    Update(ctx context.Context, idol *Idol) error
    Delete(ctx context.Context, id IdolID) error

    // 新規追加
    Search(ctx context.Context, criteria SearchCriteria) ([]*Idol, error)
    Count(ctx context.Context, criteria SearchCriteria) (int64, error)
}
```

##### 3. MongoDBリポジトリ実装
```go
// internal/infrastructure/persistence/mongodb/idol_repository.go

func (r *IdolRepository) Search(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.Idol, error) {
    filter := buildMongoFilter(criteria)

    opts := options.Find()

    // ソート設定
    sortOrder := 1
    if criteria.Order == "desc" {
        sortOrder = -1
    }
    opts.SetSort(bson.D{{Key: criteria.Sort, Value: sortOrder}})

    // ページネーション
    opts.SetSkip(int64(criteria.Offset))
    opts.SetLimit(int64(criteria.Limit))

    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var idols []*domain.Idol
    if err := cursor.All(ctx, &idols); err != nil {
        return nil, err
    }

    return idols, nil
}

func buildMongoFilter(criteria domain.SearchCriteria) bson.M {
    filter := bson.M{}

    // 名前検索（部分一致）
    if criteria.Name != nil {
        filter["name"] = bson.M{"$regex": *criteria.Name, "$options": "i"}
    }

    // 国籍（完全一致）
    if criteria.Nationality != nil {
        filter["nationality"] = *criteria.Nationality
    }

    // グループID
    if criteria.GroupID != nil {
        filter["group_id"] = *criteria.GroupID
    }

    // 年齢範囲（生年月日から逆算）
    if criteria.AgeMin != nil || criteria.AgeMax != nil {
        now := time.Now()
        birthdateFilter := bson.M{}

        if criteria.AgeMax != nil {
            // AgeMax歳より若い → 生年月日がこれより後
            minBirthdate := now.AddDate(-*criteria.AgeMax-1, 0, 0)
            birthdateFilter["$gte"] = minBirthdate
        }
        if criteria.AgeMin != nil {
            // AgeMin歳以上 → 生年月日がこれより前
            maxBirthdate := now.AddDate(-*criteria.AgeMin, 0, 0)
            birthdateFilter["$lte"] = maxBirthdate
        }

        if len(birthdateFilter) > 0 {
            filter["birthdate"] = birthdateFilter
        }
    }

    // 生年月日範囲
    if criteria.BirthdateFrom != nil || criteria.BirthdateTo != nil {
        birthdateFilter := bson.M{}
        if criteria.BirthdateFrom != nil {
            birthdateFilter["$gte"] = *criteria.BirthdateFrom
        }
        if criteria.BirthdateTo != nil {
            birthdateFilter["$lte"] = *criteria.BirthdateTo
        }
        filter["birthdate"] = birthdateFilter
    }

    return filter
}

func (r *IdolRepository) Count(ctx context.Context, criteria domain.SearchCriteria) (int64, error) {
    filter := buildMongoFilter(criteria)
    return r.collection.CountDocuments(ctx, filter)
}
```

#### 完了条件
- ✅ 複数条件での検索が動作
- ✅ ソート機能が正常動作
- ✅ MongoDBインデックスの作成

#### テストケース
```go
func TestSearchIdols(t *testing.T) {
    tests := []struct {
        name     string
        criteria domain.SearchCriteria
        want     int
    }{
        {
            name: "名前検索",
            criteria: domain.SearchCriteria{
                Name: stringPtr("山田"),
            },
            want: 2,
        },
        {
            name: "年齢範囲",
            criteria: domain.SearchCriteria{
                AgeMin: intPtr(20),
                AgeMax: intPtr(30),
            },
            want: 5,
        },
    }
    // ...
}
```

---

### Task 1.2: ハンドラー・エンドポイント実装

**工数**: 1日
**優先度**: 🔴 Critical

#### エンドポイント仕様

```
GET /api/v1/idols?name=山田&nationality=日本&sort=birthdate&order=desc&page=1&limit=20
```

**クエリパラメータ:**
- `name`: 名前（部分一致）
- `nationality`: 国籍（完全一致）
- `group_id`: グループID
- `age_min`, `age_max`: 年齢範囲
- `birthdate_from`, `birthdate_to`: 生年月日範囲（YYYY-MM-DD）
- `sort`: ソート項目（name, birthdate, created_at）
- `order`: ソート順（asc, desc）
- `page`: ページ番号（デフォルト: 1）
- `limit`: 1ページあたりの件数（デフォルト: 20、最大: 100）

**レスポンス例:**
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "name": "山田花子",
      "group": "Sample Group",
      "group_id": "507f1f77bcf86cd799439012",
      "birthdate": "2000-05-15",
      "age": 24,
      "nationality": "日本",
      "image_url": "https://example.com/image.jpg",
      "created_at": "2025-11-15T10:00:00Z",
      "updated_at": "2025-11-15T10:00:00Z"
    }
  ],
  "meta": {
    "total": 150,
    "page": 1,
    "per_page": 20,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  },
  "links": {
    "first": "/api/v1/idols?page=1",
    "prev": null,
    "next": "/api/v1/idols?page=2",
    "last": "/api/v1/idols?page=8"
  }
}
```

#### 実装
```go
// internal/interface/handlers/idol_handler.go

func (h *IdolHandler) ListIdols(c *gin.Context) {
    var query idol.ListIdolsQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": "無効なクエリパラメータです"})
        return
    }

    query.ApplyDefaults()
    if err := query.Validate(); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    result, err := h.service.SearchIdols(c.Request.Context(), query)
    if err != nil {
        c.JSON(500, gin.H{"error": "検索に失敗しました"})
        return
    }

    c.JSON(200, result)
}
```

#### 完了条件
- ✅ すべてのクエリパラメータが動作
- ✅ エラーハンドリングが適切
- ✅ APIドキュメント作成

---

### Task 1.3: MongoDBインデックス作成

**工数**: 0.5日
**優先度**: 🟡 High

#### 実装内容

```go
// internal/infrastructure/persistence/mongodb/idol_repository.go

func (r *IdolRepository) EnsureIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        {
            Keys: bson.D{
                {Key: "name", Value: 1},
            },
        },
        {
            Keys: bson.D{
                {Key: "nationality", Value: 1},
            },
        },
        {
            Keys: bson.D{
                {Key: "group_id", Value: 1},
            },
        },
        {
            Keys: bson.D{
                {Key: "birthdate", Value: 1},
            },
        },
        {
            Keys: bson.D{
                {Key: "created_at", Value: -1},
            },
        },
        // 複合インデックス
        {
            Keys: bson.D{
                {Key: "nationality", Value: 1},
                {Key: "birthdate", Value: 1},
            },
        },
    }

    _, err := r.collection.Indexes().CreateMany(ctx, indexes)
    return err
}
```

#### 完了条件
- ✅ インデックスが作成される
- ✅ 検索パフォーマンスが向上

---

## 🟡 Phase 2: ページネーション + メタデータ（Week 3）

### 目的
大量データ対応と開発者体験の向上

### 実装期間
**2-3営業日**

---

### Task 2.1: ページネーションロジック実装

**工数**: 1日
**優先度**: 🔴 Critical

#### 実装内容

##### 1. ページネーション計算ロジック
```go
// internal/application/idol/pagination.go

type PaginationMeta struct {
    Total      int64  `json:"total"`
    Page       int    `json:"page"`
    PerPage    int    `json:"per_page"`
    TotalPages int    `json:"total_pages"`
    HasNext    bool   `json:"has_next"`
    HasPrev    bool   `json:"has_prev"`
}

type PaginationLinks struct {
    First string  `json:"first"`
    Prev  *string `json:"prev"`
    Next  *string `json:"next"`
    Last  string  `json:"last"`
}

func CalculatePagination(total int64, page, perPage int, baseURL string) (PaginationMeta, PaginationLinks) {
    totalPages := int((total + int64(perPage) - 1) / int64(perPage))

    meta := PaginationMeta{
        Total:      total,
        Page:       page,
        PerPage:    perPage,
        TotalPages: totalPages,
        HasNext:    page < totalPages,
        HasPrev:    page > 1,
    }

    links := PaginationLinks{
        First: fmt.Sprintf("%s?page=1", baseURL),
        Last:  fmt.Sprintf("%s?page=%d", baseURL, totalPages),
    }

    if meta.HasNext {
        next := fmt.Sprintf("%s?page=%d", baseURL, page+1)
        links.Next = &next
    }

    if meta.HasPrev {
        prev := fmt.Sprintf("%s?page=%d", baseURL, page-1)
        links.Prev = &prev
    }

    return meta, links
}
```

##### 2. ApplicationService拡張
```go
// internal/application/idol/service.go

type SearchIdolsResult struct {
    Data  []IdolDTO       `json:"data"`
    Meta  PaginationMeta  `json:"meta"`
    Links PaginationLinks `json:"links"`
}

func (s *ApplicationService) SearchIdols(ctx context.Context, query ListIdolsQuery) (*SearchIdolsResult, error) {
    // SearchCriteriaに変換
    criteria := domain.SearchCriteria{
        Name:        query.Name,
        Nationality: query.Nationality,
        GroupID:     query.GroupID,
        AgeMin:      query.AgeMin,
        AgeMax:      query.AgeMax,
        Sort:        *query.Sort,
        Order:       *query.Order,
        Offset:      (*query.Page - 1) * *query.Limit,
        Limit:       *query.Limit,
    }

    // 並行処理: データ取得と件数取得
    var idols []*domain.Idol
    var total int64
    var errData, errCount error

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        idols, errData = s.repo.Search(ctx, criteria)
    }()

    go func() {
        defer wg.Done()
        total, errCount = s.repo.Count(ctx, criteria)
    }()

    wg.Wait()

    if errData != nil {
        return nil, errData
    }
    if errCount != nil {
        return nil, errCount
    }

    // DTOに変換
    dtos := make([]IdolDTO, len(idols))
    for i, idol := range idols {
        dtos[i] = s.toDTO(idol)
    }

    // ページネーション計算
    baseURL := "/api/v1/idols" // TODO: クエリパラメータを含める
    meta, links := CalculatePagination(total, *query.Page, *query.Limit, baseURL)

    return &SearchIdolsResult{
        Data:  dtos,
        Meta:  meta,
        Links: links,
    }, nil
}
```

#### 完了条件
- ✅ ページネーションが正確に動作
- ✅ メタデータが正しく計算される
- ✅ リンクが正しく生成される

---

### Task 2.2: カーソルベースページネーション（オプション）

**工数**: 1日
**優先度**: 🟢 Low

#### 概要
大規模データセット向けのカーソルベースページネーション実装（将来的に追加）

```
GET /api/v1/idols?cursor=eyJpZCI6IjUwN2YxZjc3In0&limit=20
```

**実装は後回し** - オフセットベースで十分な段階

---

## 🟢 Phase 3: 事務所（Agency）エンティティ（Week 4）

### 目的
データの構造化と付加価値向上

### 実装期間
**3-4営業日**

---

### Task 3.1: Agencyドメインモデル作成

**工数**: 1日
**優先度**: 🔴 Critical

#### ディレクトリ構造
```
internal/
├── domain/
│   └── agency/
│       ├── value_object.go
│       ├── agency.go
│       ├── repository.go
│       └── service.go
├── application/
│   └── agency/
│       ├── command.go
│       ├── query.go
│       └── service.go
├── infrastructure/
│   └── persistence/
│       └── mongodb/
│           └── agency_repository.go
└── interface/
    └── handlers/
        └── agency_handler.go
```

#### ドメインモデル

##### 1. 値オブジェクト
```go
// internal/domain/agency/value_object.go

type AgencyID struct {
    value string
}

func NewAgencyID(value string) (AgencyID, error) {
    if value == "" {
        return AgencyID{}, errors.New("事務所IDは空にできません")
    }
    return AgencyID{value: value}, nil
}

func (id AgencyID) Value() string {
    return id.value
}

type AgencyName struct {
    value string
}

func NewAgencyName(value string) (AgencyName, error) {
    trimmed := strings.TrimSpace(value)
    if trimmed == "" {
        return AgencyName{}, errors.New("事務所名は空にできません")
    }
    if len(trimmed) > 200 {
        return AgencyName{}, errors.New("事務所名は200文字以内にしてください")
    }
    return AgencyName{value: trimmed}, nil
}

func (n AgencyName) Value() string {
    return n.value
}

type Country struct {
    value string
}

func NewCountry(value string) (Country, error) {
    if value == "" {
        return Country{}, errors.New("国は空にできません")
    }
    validCountries := []string{"日本", "韓国", "中国", "台湾", "アメリカ", "その他"}
    if !contains(validCountries, value) {
        return Country{}, errors.New("無効な国です")
    }
    return Country{value: value}, nil
}

func (c Country) Value() string {
    return c.value
}
```

##### 2. エンティティ
```go
// internal/domain/agency/agency.go

type Agency struct {
    id              AgencyID
    name            AgencyName
    nameEn          *string      // 英語名（オプション）
    foundedDate     *time.Time   // 設立日（オプション）
    country         Country
    officialWebsite *string      // 公式サイトURL
    description     *string      // 説明
    logoURL         *string      // ロゴ画像URL
    createdAt       time.Time
    updatedAt       time.Time
}

func NewAgency(
    id AgencyID,
    name AgencyName,
    country Country,
) *Agency {
    now := time.Now()
    return &Agency{
        id:        id,
        name:      name,
        country:   country,
        createdAt: now,
        updatedAt: now,
    }
}

// Getters
func (a *Agency) ID() AgencyID           { return a.id }
func (a *Agency) Name() AgencyName       { return a.name }
func (a *Agency) NameEn() *string        { return a.nameEn }
func (a *Agency) FoundedDate() *time.Time { return a.foundedDate }
func (a *Agency) Country() Country       { return a.country }
func (a *Agency) OfficialWebsite() *string { return a.officialWebsite }
func (a *Agency) Description() *string   { return a.description }
func (a *Agency) LogoURL() *string       { return a.logoURL }
func (a *Agency) CreatedAt() time.Time   { return a.createdAt }
func (a *Agency) UpdatedAt() time.Time   { return a.updatedAt }

// ビジネスロジック
func (a *Agency) UpdateDetails(
    name *AgencyName,
    nameEn *string,
    foundedDate *time.Time,
    officialWebsite *string,
    description *string,
    logoURL *string,
) {
    if name != nil {
        a.name = *name
    }
    a.nameEn = nameEn
    a.foundedDate = foundedDate
    a.officialWebsite = officialWebsite
    a.description = description
    a.logoURL = logoURL
    a.updatedAt = time.Now()
}
```

##### 3. リポジトリインターフェース
```go
// internal/domain/agency/repository.go

type Repository interface {
    Save(ctx context.Context, agency *Agency) error
    FindByID(ctx context.Context, id AgencyID) (*Agency, error)
    FindAll(ctx context.Context) ([]*Agency, error)
    Update(ctx context.Context, agency *Agency) error
    Delete(ctx context.Context, id AgencyID) error
    ExistsByID(ctx context.Context, id AgencyID) (bool, error)
}
```

#### 完了条件
- ✅ Agencyドメインモデルが完成
- ✅ バリデーションが正常動作
- ✅ ユニットテスト作成

---

### Task 3.2: Idolエンティティへの関連付け

**工数**: 0.5日
**優先度**: 🔴 Critical

#### 実装内容

##### Idolエンティティの修正
```go
// internal/domain/idol/idol.go

type Idol struct {
    // ... 既存フィールド

    // 追加
    agencyID *agency.AgencyID  // 事務所ID（オプション）
}

func NewIdol(
    id IdolID,
    name IdolName,
    nationality Nationality,
    agencyID *agency.AgencyID,  // 追加
) *Idol {
    // ...
}

func (i *Idol) AgencyID() *agency.AgencyID {
    return i.agencyID
}

func (i *Idol) AssignToAgency(agencyID agency.AgencyID) {
    i.agencyID = &agencyID
    i.updatedAt = time.Now()
}

func (i *Idol) RemoveFromAgency() {
    i.agencyID = nil
    i.updatedAt = time.Now()
}
```

##### リポジトリメソッド追加
```go
// internal/domain/idol/repository.go

type Repository interface {
    // ... 既存メソッド

    // 追加
    FindByAgencyID(ctx context.Context, agencyID agency.AgencyID) ([]*Idol, error)
}
```

#### 完了条件
- ✅ IdolとAgencyが関連付けられる
- ✅ 事務所からアイドル一覧を取得できる

---

### Task 3.3: Agency API実装

**工数**: 1日
**優先度**: 🟡 High

#### エンドポイント

```
POST   /api/v1/agencies         # 事務所作成
GET    /api/v1/agencies         # 事務所一覧
GET    /api/v1/agencies/:id     # 事務所詳細
PUT    /api/v1/agencies/:id     # 事務所更新
DELETE /api/v1/agencies/:id     # 事務所削除

# 関連データ取得
GET    /api/v1/agencies/:id?include=idols,groups
GET    /api/v1/agencies/:id/idols    # 事務所所属アイドル
```

#### リクエスト/レスポンス例

**POST /api/v1/agencies**
```json
{
  "name": "ABC事務所",
  "name_en": "ABC Agency",
  "founded_date": "2015-04-01",
  "country": "日本",
  "official_website": "https://abc-agency.com",
  "description": "アイドル育成に特化した事務所",
  "logo_url": "https://example.com/logo.png"
}
```

**GET /api/v1/agencies/:id?include=idols**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "ABC事務所",
  "name_en": "ABC Agency",
  "founded_date": "2015-04-01",
  "country": "日本",
  "official_website": "https://abc-agency.com",
  "description": "アイドル育成に特化した事務所",
  "logo_url": "https://example.com/logo.png",
  "idols": [
    {
      "id": "...",
      "name": "山田花子"
    }
  ],
  "created_at": "2025-11-15T10:00:00Z",
  "updated_at": "2025-11-15T10:00:00Z"
}
```

#### 完了条件
- ✅ 全エンドポイントが動作
- ✅ 関連データの取得が可能
- ✅ APIドキュメント更新

---

### Task 3.4: include パラメータ実装

**工数**: 1日
**優先度**: 🟡 Medium

#### 実装内容

##### クエリパラメータ処理
```go
// internal/application/idol/query.go

type GetIdolQuery struct {
    ID      string   `uri:"id" binding:"required"`
    Include []string `form:"include"` // agency, group, social_links
}

type ListIdolsQuery struct {
    // ... 既存フィールド

    Include []string `form:"include"` // agency, group
}
```

##### ApplicationService実装
```go
// internal/application/idol/service.go

func (s *ApplicationService) GetIdol(ctx context.Context, query GetIdolQuery) (*IdolDetailDTO, error) {
    id, err := idol.NewIdolID(query.ID)
    if err != nil {
        return nil, err
    }

    i, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    dto := s.toDetailDTO(i)

    // include パラメータ処理
    for _, include := range query.Include {
        switch include {
        case "agency":
            if i.AgencyID() != nil {
                agency, err := s.agencyRepo.FindByID(ctx, *i.AgencyID())
                if err == nil {
                    dto.Agency = s.agencyToDTO(agency)
                }
            }
        case "group":
            if i.GroupID() != nil {
                group, err := s.groupRepo.FindByID(ctx, *i.GroupID())
                if err == nil {
                    dto.Group = s.groupToDTO(group)
                }
            }
        }
    }

    return dto, nil
}
```

##### DTO定義
```go
// internal/application/idol/query.go

type IdolDetailDTO struct {
    IdolDTO

    // 関連データ（includeで取得）
    Agency      *AgencyDTO      `json:"agency,omitempty"`
    Group       *GroupDTO       `json:"group,omitempty"`
    SocialLinks *SocialLinksDTO `json:"social_links,omitempty"`
}
```

#### 完了条件
- ✅ include パラメータが動作
- ✅ N+1問題が発生しない
- ✅ パフォーマンステスト実施

---

## 🎪 Phase 4: イベント/ライブエンティティ（Month 2, Week 1-2）

### 目的
イベント集約サイト構築を可能にする

### 実装期間
**7-10営業日**

---

### Task 4.1: Eventドメインモデル作成

**工数**: 2日
**優先度**: 🔴 Critical

#### ドメインモデル

##### 値オブジェクト
```go
// internal/domain/event/value_object.go

type EventID struct {
    value string
}

type EventTitle struct {
    value string
}

type EventType struct {
    value string
}

const (
    EventTypeLive        = "live"
    EventTypeHandshake   = "handshake"
    EventTypeRelease     = "release"
    EventTypeFanMeeting  = "fan_meeting"
    EventTypeOnline      = "online"
)

func NewEventType(value string) (EventType, error) {
    validTypes := []string{
        EventTypeLive,
        EventTypeHandshake,
        EventTypeRelease,
        EventTypeFanMeeting,
        EventTypeOnline,
    }
    if !contains(validTypes, value) {
        return EventType{}, errors.New("無効なイベントタイプです")
    }
    return EventType{value: value}, nil
}
```

##### エンティティ
```go
// internal/domain/event/event.go

type Event struct {
    id            EventID
    title         EventTitle
    eventType     EventType
    startDateTime time.Time
    endDateTime   *time.Time      // オプション
    venueID       *venue.VenueID  // 会場ID
    performerIDs  []string        // アイドルまたはグループのID
    ticketURL     *string
    officialURL   *string
    description   *string
    tags          []string
    createdAt     time.Time
    updatedAt     time.Time
}

func NewEvent(
    id EventID,
    title EventTitle,
    eventType EventType,
    startDateTime time.Time,
) *Event {
    now := time.Now()
    return &Event{
        id:            id,
        title:         title,
        eventType:     eventType,
        startDateTime: startDateTime,
        performerIDs:  []string{},
        tags:          []string{},
        createdAt:     now,
        updatedAt:     now,
    }
}

// ビジネスロジック
func (e *Event) AddPerformer(performerID string) error {
    // 重複チェック
    for _, id := range e.performerIDs {
        if id == performerID {
            return errors.New("既に追加されています")
        }
    }
    e.performerIDs = append(e.performerIDs, performerID)
    e.updatedAt = time.Now()
    return nil
}

func (e *Event) RemovePerformer(performerID string) {
    for i, id := range e.performerIDs {
        if id == performerID {
            e.performerIDs = append(e.performerIDs[:i], e.performerIDs[i+1:]...)
            break
        }
    }
    e.updatedAt = time.Now()
}

func (e *Event) IsUpcoming() bool {
    return e.startDateTime.After(time.Now())
}

func (e *Event) IsPast() bool {
    if e.endDateTime != nil {
        return e.endDateTime.Before(time.Now())
    }
    return e.startDateTime.Before(time.Now())
}
```

##### リポジトリインターフェース
```go
// internal/domain/event/repository.go

type SearchCriteria struct {
    EventType     *EventType
    StartDateFrom *time.Time
    StartDateTo   *time.Time
    VenueID       *venue.VenueID
    PerformerID   *string
    Tags          []string
    Prefecture    *string  // 会場の都道府県

    Sort   string
    Order  string
    Offset int
    Limit  int
}

type Repository interface {
    Save(ctx context.Context, event *Event) error
    FindByID(ctx context.Context, id EventID) (*Event, error)
    Search(ctx context.Context, criteria SearchCriteria) ([]*Event, error)
    Count(ctx context.Context, criteria SearchCriteria) (int64, error)
    Update(ctx context.Context, event *Event) error
    Delete(ctx context.Context, id EventID) error

    // 便利メソッド
    FindUpcoming(ctx context.Context, limit int) ([]*Event, error)
    FindByPerformer(ctx context.Context, performerID string, limit int) ([]*Event, error)
}
```

#### 完了条件
- ✅ Eventドメインモデルが完成
- ✅ ビジネスロジックが正常動作
- ✅ ユニットテスト作成

---

### Task 4.2: Venueドメインモデル作成

**工数**: 1日
**優先度**: 🟡 High

#### ドメインモデル

```go
// internal/domain/venue/venue.go

type Venue struct {
    id         VenueID
    name       VenueName
    address    string
    prefecture Prefecture  // 都道府県
    city       string
    latitude   *float64
    longitude  *float64
    capacity   *int
    createdAt  time.Time
    updatedAt  time.Time
}

type Prefecture struct {
    value string
}

const (
    PrefectureTokyo   = "東京都"
    PrefectureOsaka   = "大阪府"
    // ... 47都道府県
)

func NewPrefecture(value string) (Prefecture, error) {
    validPrefectures := getValidPrefectures()
    if !contains(validPrefectures, value) {
        return Prefecture{}, errors.New("無効な都道府県です")
    }
    return Prefecture{value: value}, nil
}
```

#### 完了条件
- ✅ Venueドメインモデルが完成
- ✅ リポジトリ実装完了

---

### Task 4.3: Event API実装

**工数**: 2日
**優先度**: 🔴 Critical

#### エンドポイント

```
POST   /api/v1/events                  # イベント作成
GET    /api/v1/events                  # イベント検索
GET    /api/v1/events/:id              # イベント詳細
PUT    /api/v1/events/:id              # イベント更新
DELETE /api/v1/events/:id              # イベント削除

# 検索・フィルタリング
GET    /api/v1/events?type=live&start_date_from=2025-12-01&prefecture=東京都
GET    /api/v1/events?performer_id=xxx

# カレンダー
GET    /api/v1/events/calendar?year=2025&month=12

# アイドル/グループのイベント
GET    /api/v1/idols/:id/events
GET    /api/v1/groups/:id/events
```

#### リクエスト/レスポンス例

**POST /api/v1/events**
```json
{
  "title": "Sample Group ワンマンライブ",
  "type": "live",
  "start_datetime": "2025-12-15T18:00:00+09:00",
  "end_datetime": "2025-12-15T21:00:00+09:00",
  "venue_id": "507f1f77bcf86cd799439011",
  "performer_ids": ["idol_xxx", "group_yyy"],
  "ticket_url": "https://example.com/tickets",
  "official_url": "https://example.com/event",
  "description": "待望のワンマンライブ！",
  "tags": ["ワンマンライブ", "全国ツアー"]
}
```

**GET /api/v1/events?start_date_from=2025-12-01&prefecture=東京都**
```json
{
  "data": [
    {
      "id": "event_123",
      "title": "Sample Group ワンマンライブ",
      "type": "live",
      "start_datetime": "2025-12-15T18:00:00+09:00",
      "end_datetime": "2025-12-15T21:00:00+09:00",
      "venue": {
        "id": "venue_xxx",
        "name": "東京ドーム",
        "prefecture": "東京都",
        "city": "文京区"
      },
      "performers": [
        {
          "id": "group_yyy",
          "type": "group",
          "name": "Sample Group"
        }
      ],
      "ticket_url": "https://example.com/tickets",
      "tags": ["ワンマンライブ"]
    }
  ],
  "meta": {
    "total": 25,
    "page": 1,
    "per_page": 20
  }
}
```

#### 完了条件
- ✅ 全エンドポイントが動作
- ✅ 時系列検索が正確
- ✅ APIドキュメント更新

---

### Task 4.4: カレンダーAPI実装

**工数**: 1日
**優先度**: 🟡 Medium

#### エンドポイント

```
GET /api/v1/events/calendar?year=2025&month=12&performer_id=xxx
```

#### レスポンス例

```json
{
  "year": 2025,
  "month": 12,
  "events_by_date": {
    "2025-12-15": [
      {
        "id": "event_123",
        "title": "Sample Group ワンマンライブ",
        "start_time": "18:00",
        "venue": "東京ドーム"
      }
    ],
    "2025-12-20": [
      {
        "id": "event_124",
        "title": "握手会",
        "start_time": "13:00",
        "venue": "幕張メッセ"
      }
    ]
  },
  "total_events": 8
}
```

#### 完了条件
- ✅ カレンダー形式で取得可能
- ✅ iCalendar形式エクスポート（オプション）

---

## 📱 Phase 5: SNS/外部リンク管理（Month 2, Week 3）

### 実装期間
**2-3営業日**

---

### Task 5.1: SocialLinksドメインモデル作成

**工数**: 0.5日
**優先度**: 🟡 High

#### ドメインモデル

```go
// internal/domain/idol/social_links.go

type SocialLinks struct {
    twitter   *string
    instagram *string
    tiktok    *string
    youtube   *string
    facebook  *string
    official  *string
    fanClub   *string
}

func NewSocialLinks() *SocialLinks {
    return &SocialLinks{}
}

func (s *SocialLinks) SetTwitter(url string) error {
    if err := validateURL(url); err != nil {
        return err
    }
    if !strings.Contains(url, "twitter.com") && !strings.Contains(url, "x.com") {
        return errors.New("無効なTwitter URLです")
    }
    s.twitter = &url
    return nil
}

func (s *SocialLinks) SetInstagram(url string) error {
    if err := validateURL(url); err != nil {
        return err
    }
    if !strings.Contains(url, "instagram.com") {
        return errors.New("無効なInstagram URLです")
    }
    s.instagram = &url
    return nil
}

// 他のSNSも同様

func validateURL(url string) error {
    if url == "" {
        return nil
    }
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        return errors.New("URLはhttp://またはhttps://で始まる必要があります")
    }
    return nil
}
```

#### Idolエンティティへの追加

```go
// internal/domain/idol/idol.go

type Idol struct {
    // ... 既存フィールド

    socialLinks *SocialLinks
}

func (i *Idol) SocialLinks() *SocialLinks {
    return i.socialLinks
}

func (i *Idol) UpdateSocialLinks(links *SocialLinks) {
    i.socialLinks = links
    i.updatedAt = time.Now()
}
```

#### 完了条件
- ✅ SocialLinksドメインモデルが完成
- ✅ URLバリデーションが動作

---

### Task 5.2: SNSリンクAPI実装

**工数**: 1日
**優先度**: 🟡 High

#### エンドポイント

```
PUT /api/v1/idols/:id/social-links      # SNSリンク更新
GET /api/v1/idols/:id?include=social_links
```

#### リクエスト例

**PUT /api/v1/idols/:id/social-links**
```json
{
  "twitter": "https://twitter.com/idol_username",
  "instagram": "https://instagram.com/idol_username",
  "tiktok": "https://tiktok.com/@idol_username",
  "youtube": "https://youtube.com/@idol_channel",
  "official_website": "https://idol-official.com",
  "fan_club": "https://fanclub.idol.com"
}
```

#### 完了条件
- ✅ SNSリンク更新が動作
- ✅ include=social_links で取得可能

---

## 📚 Phase 6: OpenAPI仕様書 + ドキュメント（Month 3, Week 1-2）

### 実装期間
**5-7営業日**

---

### Task 6.1: Swaggerアノテーション追加

**工数**: 3日
**優先度**: 🔴 Critical

#### ツールインストール

```bash
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

#### アノテーション例

```go
// internal/interface/handlers/idol_handler.go

// ListIdols godoc
// @Summary      アイドル一覧取得
// @Description  条件を指定してアイドル一覧を取得
// @Tags         idols
// @Accept       json
// @Produce      json
// @Param        name query string false "名前（部分一致）"
// @Param        nationality query string false "国籍"
// @Param        group_id query string false "グループID"
// @Param        age_min query int false "最小年齢"
// @Param        age_max query int false "最大年齢"
// @Param        sort query string false "ソート項目" Enums(name, birthdate, created_at)
// @Param        order query string false "ソート順" Enums(asc, desc)
// @Param        page query int false "ページ番号"
// @Param        limit query int false "1ページあたりの件数"
// @Success      200 {object} idol.SearchIdolsResult
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/v1/idols [get]
func (h *IdolHandler) ListIdols(c *gin.Context) {
    // ...
}
```

#### 完了条件
- ✅ 全エンドポイントにアノテーション追加
- ✅ `swag init` でドキュメント生成成功

---

### Task 6.2: Swagger UI設定

**工数**: 1日
**優先度**: 🔴 Critical

#### 実装

```go
// cmd/api/main.go

import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

    _ "github.com/kuro48/idol-api/docs" // swag init で生成されるドキュメント
)

// @title           Idol API
// @version         1.0
// @description     包括的アイドル情報API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8081
// @BasePath  /api/v1

func main() {
    // ... 既存の設定

    // Swagger UI
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // ... サーバー起動
}
```

#### アクセス

```
http://localhost:8081/swagger/index.html
```

#### 完了条件
- ✅ Swagger UIが正常表示
- ✅ 全エンドポイントがテスト可能

---

### Task 6.3: APIドキュメント作成

**工数**: 2日
**優先度**: 🟡 Medium

#### ドキュメント内容

1. **Getting Started**
   - 認証方法（将来実装）
   - レート制限
   - エラーハンドリング

2. **エンドポイント一覧**
   - アイドル検索
   - イベント検索
   - 事務所情報

3. **データモデル**
   - Idol
   - Group
   - Agency
   - Event
   - Venue

4. **使用例**
   - JavaScript/TypeScriptサンプル
   - Pythonサンプル
   - cURLコマンド

#### 完了条件
- ✅ ドキュメントサイト公開
- ✅ サンプルコード動作確認

---

## 🏷️ Phase 7: タグ・カテゴリシステム（Month 3, Week 3）

### 実装期間
**3-4営業日**

---

### Task 7.1: Tagドメインモデル作成

**工数**: 1日
**優先度**: 🟢 Low

#### ドメインモデル

```go
// internal/domain/tag/tag.go

type Tag struct {
    id          TagID
    name        TagName
    category    TagCategory  // genre, region, style, etc.
    description *string
    createdAt   time.Time
}

type TagCategory struct {
    value string
}

const (
    TagCategoryGenre  = "genre"   // K-POP, J-POP, アイドルポップ
    TagCategoryRegion = "region"  // 関東、関西、韓国
    TagCategoryStyle  = "style"   // ダンス、ボーカル、ラップ
    TagCategoryOther  = "other"
)
```

#### Idol/Group/Eventへのタグ付け

```go
// internal/domain/idol/idol.go

type Idol struct {
    // ... 既存フィールド

    tags []tag.TagID
}

func (i *Idol) AddTag(tagID tag.TagID) error {
    // 重複チェック
    for _, t := range i.tags {
        if t.Value() == tagID.Value() {
            return errors.New("既に追加されています")
        }
    }
    i.tags = append(i.tags, tagID)
    i.updatedAt = time.Now()
    return nil
}

func (i *Idol) RemoveTag(tagID tag.TagID) {
    for idx, t := range i.tags {
        if t.Value() == tagID.Value() {
            i.tags = append(i.tags[:idx], i.tags[idx+1:]...)
            break
        }
    }
    i.updatedAt = time.Now()
}
```

#### 完了条件
- ✅ Tagドメインモデルが完成
- ✅ アイドル/グループ/イベントにタグ付け可能

---

### Task 7.2: タグ検索API実装

**工数**: 1日
**優先度**: 🟢 Low

#### エンドポイント

```
GET /api/v1/idols?tags=K-POP,ダンス
GET /api/v1/events?tags=ワンマンライブ

POST   /api/v1/tags        # タグ作成
GET    /api/v1/tags        # タグ一覧
```

#### 完了条件
- ✅ タグ検索が動作
- ✅ タグ管理API完成

---

## ⚡ Phase 8: パフォーマンス最適化（Month 3, Week 4）

### 実装期間
**3-5営業日**

---

### Task 8.1: MongoDBクエリ最適化

**工数**: 2日
**優先度**: 🟡 High

#### 実装内容

1. **複合インデックス追加**
```go
// 複合インデックスの作成
indexes := []mongo.IndexModel{
    {
        Keys: bson.D{
            {Key: "nationality", Value: 1},
            {Key: "birthdate", Value: 1},
            {Key: "created_at", Value: -1},
        },
    },
    {
        Keys: bson.D{
            {Key: "agency_id", Value: 1},
            {Key: "created_at", Value: -1},
        },
    },
}
```

2. **クエリプロファイリング**
   - スロークエリの特定
   - explain() での実行計画確認

3. **集計パイプライン最適化**

#### 完了条件
- ✅ 検索レスポンス時間 < 200ms
- ✅ スロークエリ解消

---

### Task 8.2: キャッシュ実装

**工数**: 2日
**優先度**: 🟢 Medium（将来実装）

#### 概要

Redis導入による頻繁にアクセスされるデータのキャッシュ

**現時点では実装せず**、トラフィック増加時に対応

---

## 🧪 テスト計画

### ユニットテスト
- ドメイン層: 全値オブジェクト、エンティティ
- アプリケーション層: ApplicationService
- **カバレッジ目標**: 80%以上

### 統合テスト
- リポジトリ層: MongoDBテストコンテナ使用
- API層: httptest使用

### E2Eテスト
- 主要ユースケースのシナリオテスト
- 検索フロー
- イベント作成〜検索フロー

---

## 📊 マイルストーン

### **Milestone 1 (Month 1終了時)**
**期限**: 2025年12月31日

**成果物:**
- ✅ 柔軟な検索・フィルタリングAPI
- ✅ ページネーション実装
- ✅ 事務所エンティティ完成
- ✅ 関連データ取得（include パラメータ）

**検証基準:**
- 複数条件での検索が正常動作
- ページネーションが正確
- アイドル-事務所が関連付けられる

---

### **Milestone 2 (Month 2終了時)**
**期限**: 2026年1月31日

**成果物:**
- ✅ イベント/ライブエンティティ完成
- ✅ 会場管理機能
- ✅ 時系列検索・フィルタリング
- ✅ SNS/外部リンク管理

**検証基準:**
- イベント検索が正常動作
- カレンダーAPIが機能
- SNSリンクが正しく表示

---

### **Milestone 3 (Month 3終了時)**
**期限**: 2026年2月28日

**成果物:**
- ✅ OpenAPI仕様書 + Swagger UI
- ✅ タグ・カテゴリシステム
- ✅ パフォーマンス最適化
- ✅ 総合テスト完了

**検証基準:**
- Swagger UIでの全エンドポイントテスト成功
- 検索レスポンス < 200ms
- テストカバレッジ 80%以上

---

## 🎯 優先度マトリクス

### 🔴 P0 (最優先)
- Phase 1: 検索・フィルタリング基盤
- Phase 2: ページネーション
- Phase 3: 事務所エンティティ
- Phase 4: イベントエンティティ
- Phase 6: OpenAPI仕様書

### 🟡 P1 (高優先度)
- Phase 4: 会場管理
- Phase 5: SNSリンク管理
- Phase 6: APIドキュメント
- Phase 8: パフォーマンス最適化

### 🟢 P2 (中優先度)
- Phase 7: タグ・カテゴリシステム
- Phase 8: キャッシュ実装

---

## 🚀 次のアクション

### Week 1 開始前（準備）
1. ✅ 実装計画レビュー
2. ✅ 開発環境確認
3. ✅ Gitブランチ戦略決定

### Week 1 Day 1
1. **Phase 1 開始**: 検索・フィルタリング基盤
   - Task 1.1: クエリパラメータ処理基盤
   - リポジトリインターフェース拡張

### 毎週金曜日
- 週次レビュー
- 進捗確認
- 次週計画調整

---

## 📝 補足資料

### 技術スタック
- **言語**: Go 1.24.4
- **Webフレームワーク**: Gin
- **データベース**: MongoDB v2
- **ドキュメント**: Swagger/OpenAPI 3.0
- **テスト**: testify, httptest

### 関連ドキュメント
- README.md - プロジェクト概要
- docs/ddd-architecture-guide.md - DDDアーキテクチャガイド
- docs/implementation-plan.md - Phase 1 MVP実装計画（旧）

### 参考リンク
- [Gin Framework](https://gin-gonic.com/)
- [Swaggo](https://github.com/swaggo/swag)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)

---

**最終更新**: 2025-11-26
