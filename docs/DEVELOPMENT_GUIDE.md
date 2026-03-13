# 🛠️ 開発ガイド - DDDアーキテクチャでの開発方法

## 📋 目次
1. [開発の基本原則](#開発の基本原則)
2. [新機能の追加手順](#新機能の追加手順)
3. [各層での開発ルール](#各層での開発ルール)
4. [ファイルの命名規則](#ファイルの命名規則)
5. [コーディング規約](#コーディング規約)
6. [よくある間違いと回避方法](#よくある間違いと回避方法)
7. [具体例: 新機能追加](#具体例-新機能追加)

---

## 🎯 開発の基本原則

### 1. **ドメイン層から始める**
新機能を追加するときは、必ず**ドメイン層**から始めましょう。

```
❌ 悪い順序: ハンドラー → サービス → ドメイン → リポジトリ
✅ 良い順序: ドメイン → リポジトリIF → アプリケーション → インフラ → ハンドラー
```

**理由**: ビジネスロジックを先に決めないと、後で大幅な修正が必要になる

### 2. **依存の方向を守る**
```
プレゼンテーション層
    ↓ (依存OK)
アプリケーション層
    ↓ (依存OK)
ドメイン層 ← 誰にも依存しない!
    ↑ (実装)
インフラ層
```

**ルール**:
- ドメイン層は `fmt`, `time`, `errors` など標準ライブラリのみ使用
- 外部パッケージ (MongoDB, Gin など) は**絶対に使わない**
- インフラ層は外部パッケージOK

### 3. **1ファイル1責任**
1つのファイルには1つの役割だけを持たせる

```
✅ 良い例:
- idol_name.go     (IdolName値オブジェクトのみ)
- birthdate.go     (Birthdate値オブジェクトのみ)
- idol.go          (Idolエンティティのみ)

❌ 悪い例:
- idol_all.go      (全部まとめて書く)
```

---

## 🚀 新機能の追加手順

### ステップ1: ドメイン層を作る

#### 1-1. エンティティの作成
```bash
# ファイル: internal/domain/concert/concert.go
```

```go
package concert

import (
    "time"
    "errors"
)

// Concert はコンサートエンティティ（集約ルート）
type Concert struct {
    id          ConcertID
    title       ConcertTitle
    venue       Venue
    startTime   time.Time
    idolIDs     []IdolID
    createdAt   time.Time
    updatedAt   time.Time
}

// NewConcert は新しいコンサートを作成
func NewConcert(
    title ConcertTitle,
    venue Venue,
    startTime time.Time,
) (*Concert, error) {
    // ビジネスルールのチェック
    if startTime.Before(time.Now()) {
        return nil, errors.New("過去の日時は設定できません")
    }

    now := time.Now()
    return &Concert{
        title:     title,
        venue:     venue,
        startTime: startTime,
        idolIDs:   []IdolID{},
        createdAt: now,
        updatedAt: now,
    }, nil
}

// Reconstruct はデータストアから再構築（永続化層用）
func Reconstruct(
    id ConcertID,
    title ConcertTitle,
    venue Venue,
    startTime time.Time,
    idolIDs []IdolID,
    createdAt time.Time,
    updatedAt time.Time,
) *Concert {
    return &Concert{
        id:        id,
        title:     title,
        venue:     venue,
        startTime: startTime,
        idolIDs:   idolIDs,
        createdAt: createdAt,
        updatedAt: updatedAt,
    }
}

// ゲッター
func (c *Concert) ID() ConcertID           { return c.id }
func (c *Concert) Title() ConcertTitle     { return c.title }
func (c *Concert) Venue() Venue            { return c.venue }
func (c *Concert) StartTime() time.Time    { return c.startTime }
func (c *Concert) IdolIDs() []IdolID       { return c.idolIDs }
func (c *Concert) CreatedAt() time.Time    { return c.createdAt }
func (c *Concert) UpdatedAt() time.Time    { return c.updatedAt }

// ビジネスロジック

// AddIdol はコンサートにアイドルを追加
func (c *Concert) AddIdol(idolID IdolID) error {
    // 重複チェック
    for _, id := range c.idolIDs {
        if id.Equals(idolID) {
            return errors.New("このアイドルは既に追加されています")
        }
    }

    c.idolIDs = append(c.idolIDs, idolID)
    c.updatedAt = time.Now()
    return nil
}

// RemoveIdol はコンサートからアイドルを削除
func (c *Concert) RemoveIdol(idolID IdolID) error {
    for i, id := range c.idolIDs {
        if id.Equals(idolID) {
            c.idolIDs = append(c.idolIDs[:i], c.idolIDs[i+1:]...)
            c.updatedAt = time.Now()
            return nil
        }
    }
    return errors.New("アイドルが見つかりません")
}

// CanCancel はコンサートがキャンセル可能かチェック
func (c *Concert) CanCancel() bool {
    // 開始24時間前までキャンセル可能
    cancelDeadline := c.startTime.Add(-24 * time.Hour)
    return time.Now().Before(cancelDeadline)
}

// SetID はIDを設定（永続化後に使用）
func (c *Concert) SetID(id ConcertID) {
    c.id = id
}
```

#### 1-2. 値オブジェクトの作成
```bash
# ファイル: internal/domain/concert/concert_id.go
```

```go
package concert

import "errors"

type ConcertID struct {
    value string
}

func NewConcertID(id string) (ConcertID, error) {
    if id == "" {
        return ConcertID{}, errors.New("IDは空にできません")
    }
    return ConcertID{value: id}, nil
}

func (id ConcertID) Value() string {
    return id.value
}

func (id ConcertID) Equals(other ConcertID) bool {
    return id.value == other.value
}
```

```bash
# ファイル: internal/domain/concert/concert_title.go
```

```go
package concert

import "errors"

type ConcertTitle struct {
    value string
}

func NewConcertTitle(title string) (ConcertTitle, error) {
    if title == "" {
        return ConcertTitle{}, errors.New("タイトルは空にできません")
    }
    if len(title) > 200 {
        return ConcertTitle{}, errors.New("タイトルは200文字以内です")
    }
    return ConcertTitle{value: title}, nil
}

func (t ConcertTitle) Value() string {
    return t.value
}
```

```bash
# ファイル: internal/domain/concert/venue.go
```

```go
package concert

import "errors"

type Venue struct {
    name     string
    address  string
    capacity int
}

func NewVenue(name, address string, capacity int) (Venue, error) {
    if name == "" {
        return Venue{}, errors.New("会場名は空にできません")
    }
    if capacity <= 0 {
        return Venue{}, errors.New("収容人数は1以上である必要があります")
    }
    return Venue{
        name:     name,
        address:  address,
        capacity: capacity,
    }, nil
}

func (v Venue) Name() string     { return v.name }
func (v Venue) Address() string  { return v.address }
func (v Venue) Capacity() int    { return v.capacity }
```

#### 1-3. リポジトリインターフェースの作成
```bash
# ファイル: internal/domain/concert/repository.go
```

```go
package concert

import "context"

// Repository はコンサートリポジトリのインターフェース
type Repository interface {
    Save(ctx context.Context, concert *Concert) error
    FindByID(ctx context.Context, id ConcertID) (*Concert, error)
    FindAll(ctx context.Context) ([]*Concert, error)
    FindByIdolID(ctx context.Context, idolID IdolID) ([]*Concert, error)
    Update(ctx context.Context, concert *Concert) error
    Delete(ctx context.Context, id ConcertID) error
}
```

#### 1-4. ドメインサービスの作成（必要な場合のみ）
```bash
# ファイル: internal/domain/concert/service.go
```

```go
package concert

import (
    "context"
    "errors"
)

// DomainService はコンサートドメインサービス
type DomainService struct {
    repository Repository
}

func NewDomainService(repository Repository) *DomainService {
    return &DomainService{repository: repository}
}

// CanCreateConcert は同じ会場・時間にコンサートがないかチェック
func (s *DomainService) CanCreateConcert(ctx context.Context, venue Venue, startTime time.Time) error {
    // 複雑なビジネスルール: 同じ会場で同じ時間にコンサートは開催できない
    // （実際の実装では、時間の重複をチェックする）

    allConcerts, err := s.repository.FindAll(ctx)
    if err != nil {
        return err
    }

    for _, concert := range allConcerts {
        if concert.Venue().Name() == venue.Name() {
            // 開始時間が4時間以内なら重複とみなす
            diff := startTime.Sub(concert.StartTime())
            if diff > -4*time.Hour && diff < 4*time.Hour {
                return errors.New("同じ会場で近い時間帯にコンサートが既に予定されています")
            }
        }
    }

    return nil
}
```

#### 1-5. エラー定義（オプション）
```bash
# ファイル: internal/domain/concert/error.go
```

```go
package concert

import "errors"

var (
    ErrConcertNotFound      = errors.New("コンサートが見つかりません")
    ErrInvalidConcertTitle  = errors.New("無効なコンサートタイトル")
    ErrInvalidVenue         = errors.New("無効な会場情報")
    ErrConcertInPast        = errors.New("過去の日時は設定できません")
    ErrVenueConflict        = errors.New("会場が重複しています")
    ErrCannotCancel         = errors.New("キャンセル期限を過ぎています")
)
```

---

### ステップ2: アプリケーション層を作る

#### 2-1. コマンドとクエリの定義
```bash
# ファイル: internal/application/concert/command.go
```

```go
package concert

// CreateConcertCommand はコンサート作成コマンド
type CreateConcertCommand struct {
    Title       string
    VenueName   string
    VenueAddress string
    Capacity    int
    StartTime   string // "2024-12-31T19:00:00Z"
}

// UpdateConcertCommand はコンサート更新コマンド
type UpdateConcertCommand struct {
    ID          string
    Title       *string
    VenueName   *string
    VenueAddress *string
    Capacity    *int
    StartTime   *string
}

// DeleteConcertCommand はコンサート削除コマンド
type DeleteConcertCommand struct {
    ID string
}

// AddIdolToConcertCommand はアイドル追加コマンド
type AddIdolToConcertCommand struct {
    ConcertID string
    IdolID    string
}
```

```bash
# ファイル: internal/application/concert/query.go
```

```go
package concert

// GetConcertQuery はコンサート取得クエリ
type GetConcertQuery struct {
    ID string
}

// ListConcertsQuery はコンサート一覧取得クエリ
type ListConcertsQuery struct {
    // フィルター条件を追加可能
    IdolID *string
}

// ConcertDTO はコンサートのデータ転送オブジェクト
type ConcertDTO struct {
    ID          string   `json:"id"`
    Title       string   `json:"title"`
    VenueName   string   `json:"venue_name"`
    VenueAddress string  `json:"venue_address"`
    Capacity    int      `json:"capacity"`
    StartTime   string   `json:"start_time"`
    IdolIDs     []string `json:"idol_ids"`
    CreatedAt   string   `json:"created_at"`
    UpdatedAt   string   `json:"updated_at"`
}
```

#### 2-2. アプリケーションサービスの作成
```bash
# ファイル: internal/application/concert/service.go
```

```go
package concert

import (
    "context"
    "fmt"
    "time"

    "github.com/kuro48/idol-api/internal/domain/concert"
)

// ApplicationService はコンサートアプリケーションサービス
type ApplicationService struct {
    repository    concert.Repository
    domainService *concert.DomainService
}

func NewApplicationService(repository concert.Repository) *ApplicationService {
    return &ApplicationService{
        repository:    repository,
        domainService: concert.NewDomainService(repository),
    }
}

// CreateConcert はコンサートを作成
func (s *ApplicationService) CreateConcert(ctx context.Context, cmd CreateConcertCommand) (*ConcertDTO, error) {
    // 1. 値オブジェクトの生成
    title, err := concert.NewConcertTitle(cmd.Title)
    if err != nil {
        return nil, fmt.Errorf("タイトル生成エラー: %w", err)
    }

    venue, err := concert.NewVenue(cmd.VenueName, cmd.VenueAddress, cmd.Capacity)
    if err != nil {
        return nil, fmt.Errorf("会場情報生成エラー: %w", err)
    }

    startTime, err := time.Parse(time.RFC3339, cmd.StartTime)
    if err != nil {
        return nil, fmt.Errorf("日時パースエラー: %w", err)
    }

    // 2. ドメインサービスでバリデーション
    if err := s.domainService.CanCreateConcert(ctx, venue, startTime); err != nil {
        return nil, err
    }

    // 3. エンティティ生成
    newConcert, err := concert.NewConcert(title, venue, startTime)
    if err != nil {
        return nil, fmt.Errorf("コンサート生成エラー: %w", err)
    }

    // 4. 永続化
    if err := s.repository.Save(ctx, newConcert); err != nil {
        return nil, fmt.Errorf("保存エラー: %w", err)
    }

    return s.toDTO(newConcert), nil
}

// GetConcert はコンサートを取得
func (s *ApplicationService) GetConcert(ctx context.Context, query GetConcertQuery) (*ConcertDTO, error) {
    id, err := concert.NewConcertID(query.ID)
    if err != nil {
        return nil, fmt.Errorf("ID生成エラー: %w", err)
    }

    foundConcert, err := s.repository.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("取得エラー: %w", err)
    }

    return s.toDTO(foundConcert), nil
}

// ListConcerts はコンサート一覧を取得
func (s *ApplicationService) ListConcerts(ctx context.Context, query ListConcertsQuery) ([]*ConcertDTO, error) {
    var concerts []*concert.Concert
    var err error

    if query.IdolID != nil {
        // アイドルIDでフィルタリング
        idolID, err := concert.NewIdolID(*query.IdolID)
        if err != nil {
            return nil, fmt.Errorf("アイドルID生成エラー: %w", err)
        }
        concerts, err = s.repository.FindByIdolID(ctx, idolID)
    } else {
        concerts, err = s.repository.FindAll(ctx)
    }

    if err != nil {
        return nil, fmt.Errorf("一覧取得エラー: %w", err)
    }

    dtos := make([]*ConcertDTO, 0, len(concerts))
    for _, c := range concerts {
        dtos = append(dtos, s.toDTO(c))
    }

    return dtos, nil
}

// AddIdolToConcert はコンサートにアイドルを追加
func (s *ApplicationService) AddIdolToConcert(ctx context.Context, cmd AddIdolToConcertCommand) error {
    // 1. コンサートを取得
    concertID, err := concert.NewConcertID(cmd.ConcertID)
    if err != nil {
        return fmt.Errorf("コンサートID生成エラー: %w", err)
    }

    foundConcert, err := s.repository.FindByID(ctx, concertID)
    if err != nil {
        return fmt.Errorf("コンサート取得エラー: %w", err)
    }

    // 2. アイドルを追加（ドメインロジック）
    idolID, err := concert.NewIdolID(cmd.IdolID)
    if err != nil {
        return fmt.Errorf("アイドルID生成エラー: %w", err)
    }

    if err := foundConcert.AddIdol(idolID); err != nil {
        return err
    }

    // 3. 更新を保存
    if err := s.repository.Update(ctx, foundConcert); err != nil {
        return fmt.Errorf("更新エラー: %w", err)
    }

    return nil
}

// toDTO はドメインモデルをDTOに変換
func (s *ApplicationService) toDTO(c *concert.Concert) *ConcertDTO {
    idolIDs := make([]string, len(c.IdolIDs()))
    for i, id := range c.IdolIDs() {
        idolIDs[i] = id.Value()
    }

    return &ConcertDTO{
        ID:          c.ID().Value(),
        Title:       c.Title().Value(),
        VenueName:   c.Venue().Name(),
        VenueAddress: c.Venue().Address(),
        Capacity:    c.Venue().Capacity(),
        StartTime:   c.StartTime().Format(time.RFC3339),
        IdolIDs:     idolIDs,
        CreatedAt:   c.CreatedAt().Format(time.RFC3339),
        UpdatedAt:   c.UpdatedAt().Format(time.RFC3339),
    }
}
```

---

### ステップ3: インフラ層を作る

```bash
# ファイル: internal/infrastructure/persistence/mongodb/concert_repository.go
```

```go
package mongodb

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/kuro48/idol-api/internal/domain/concert"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

type ConcertRepository struct {
    collection *mongo.Collection
}

func NewConcertRepository(db *mongo.Database) *ConcertRepository {
    return &ConcertRepository{
        collection: db.Collection("concerts"),
    }
}

// concertDocument はMongoDBドキュメント構造
type concertDocument struct {
    ID           bson.ObjectID   `bson:"_id,omitempty"`
    Title        string          `bson:"title"`
    VenueName    string          `bson:"venue_name"`
    VenueAddress string          `bson:"venue_address"`
    Capacity     int             `bson:"capacity"`
    StartTime    time.Time       `bson:"start_time"`
    IdolIDs      []bson.ObjectID `bson:"idol_ids"`
    CreatedAt    time.Time       `bson:"created_at"`
    UpdatedAt    time.Time       `bson:"updated_at"`
}

// toDocument はドメインモデルをMongoDBドキュメントに変換
func toConcertDocument(c *concert.Concert) *concertDocument {
    objectID, _ := bson.ObjectIDFromHex(c.ID().Value())

    idolIDs := make([]bson.ObjectID, len(c.IdolIDs()))
    for i, id := range c.IdolIDs() {
        idolIDs[i], _ = bson.ObjectIDFromHex(id.Value())
    }

    return &concertDocument{
        ID:           objectID,
        Title:        c.Title().Value(),
        VenueName:    c.Venue().Name(),
        VenueAddress: c.Venue().Address(),
        Capacity:     c.Venue().Capacity(),
        StartTime:    c.StartTime(),
        IdolIDs:      idolIDs,
        CreatedAt:    c.CreatedAt(),
        UpdatedAt:    c.UpdatedAt(),
    }
}

// toDomain はMongoDBドキュメントをドメインモデルに変換
func toConcertDomain(doc *concertDocument) (*concert.Concert, error) {
    id, err := concert.NewConcertID(doc.ID.Hex())
    if err != nil {
        return nil, err
    }

    title, err := concert.NewConcertTitle(doc.Title)
    if err != nil {
        return nil, err
    }

    venue, err := concert.NewVenue(doc.VenueName, doc.VenueAddress, doc.Capacity)
    if err != nil {
        return nil, err
    }

    idolIDs := make([]concert.IdolID, len(doc.IdolIDs))
    for i, objID := range doc.IdolIDs {
        idolIDs[i], _ = concert.NewIdolID(objID.Hex())
    }

    return concert.Reconstruct(
        id,
        title,
        venue,
        doc.StartTime,
        idolIDs,
        doc.CreatedAt,
        doc.UpdatedAt,
    ), nil
}

// Save は新しいコンサートを保存
func (r *ConcertRepository) Save(ctx context.Context, c *concert.Concert) error {
    doc := toConcertDocument(c)

    if doc.ID.IsZero() {
        doc.ID = bson.NewObjectID()
        doc.CreatedAt = time.Now()
        doc.UpdatedAt = time.Now()
    }

    _, err := r.collection.InsertOne(ctx, doc)
    if err != nil {
        return fmt.Errorf("コンサートの保存エラー: %w", err)
    }

    // IDをエンティティに設定
    id, _ := concert.NewConcertID(doc.ID.Hex())
    c.SetID(id)

    return nil
}

// FindByID はIDでコンサートを検索
func (r *ConcertRepository) FindByID(ctx context.Context, id concert.ConcertID) (*concert.Concert, error) {
    objectID, err := bson.ObjectIDFromHex(id.Value())
    if err != nil {
        return nil, fmt.Errorf("無効なID形式: %w", err)
    }

    var doc concertDocument
    err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("コンサートが見つかりません")
        }
        return nil, fmt.Errorf("コンサート取得エラー: %w", err)
    }

    return toConcertDomain(&doc)
}

// FindAll は全てのコンサートを取得
func (r *ConcertRepository) FindAll(ctx context.Context) ([]*concert.Concert, error) {
    cursor, err := r.collection.Find(ctx, bson.M{})
    if err != nil {
        return nil, fmt.Errorf("コンサート一覧取得エラー: %w", err)
    }
    defer cursor.Close(ctx)

    var docs []concertDocument
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, fmt.Errorf("データ変換エラー: %w", err)
    }

    concerts := make([]*concert.Concert, 0, len(docs))
    for _, doc := range docs {
        c, err := toConcertDomain(&doc)
        if err != nil {
            return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
        }
        concerts = append(concerts, c)
    }

    return concerts, nil
}

// FindByIdolID はアイドルIDでコンサートを検索
func (r *ConcertRepository) FindByIdolID(ctx context.Context, idolID concert.IdolID) ([]*concert.Concert, error) {
    objectID, err := bson.ObjectIDFromHex(idolID.Value())
    if err != nil {
        return nil, fmt.Errorf("無効なアイドルID形式: %w", err)
    }

    cursor, err := r.collection.Find(ctx, bson.M{"idol_ids": objectID})
    if err != nil {
        return nil, fmt.Errorf("コンサート検索エラー: %w", err)
    }
    defer cursor.Close(ctx)

    var docs []concertDocument
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, fmt.Errorf("データ変換エラー: %w", err)
    }

    concerts := make([]*concert.Concert, 0, len(docs))
    for _, doc := range docs {
        c, err := toConcertDomain(&doc)
        if err != nil {
            return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
        }
        concerts = append(concerts, c)
    }

    return concerts, nil
}

// Update は既存のコンサートを更新
func (r *ConcertRepository) Update(ctx context.Context, c *concert.Concert) error {
    objectID, err := bson.ObjectIDFromHex(c.ID().Value())
    if err != nil {
        return fmt.Errorf("無効なID形式: %w", err)
    }

    doc := toConcertDocument(c)
    doc.UpdatedAt = time.Now()

    updateDoc := bson.M{
        "$set": bson.M{
            "title":         doc.Title,
            "venue_name":    doc.VenueName,
            "venue_address": doc.VenueAddress,
            "capacity":      doc.Capacity,
            "start_time":    doc.StartTime,
            "idol_ids":      doc.IdolIDs,
            "updated_at":    doc.UpdatedAt,
        },
    }

    result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
    if err != nil {
        return fmt.Errorf("コンサート更新エラー: %w", err)
    }

    if result.MatchedCount == 0 {
        return errors.New("コンサートが見つかりません")
    }

    return nil
}

// Delete はコンサートを削除
func (r *ConcertRepository) Delete(ctx context.Context, id concert.ConcertID) error {
    objectID, err := bson.ObjectIDFromHex(id.Value())
    if err != nil {
        return fmt.Errorf("無効なID形式: %w", err)
    }

    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return fmt.Errorf("コンサート削除エラー: %w", err)
    }

    if result.DeletedCount == 0 {
        return errors.New("コンサートが見つかりません")
    }

    return nil
}
```

---

### ステップ4: プレゼンテーション層を作る

```bash
# ファイル: internal/interface/handlers/concert_handler.go
```

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/kuro48/idol-api/internal/application/concert"
)

type ConcertHandler struct {
    appService *concert.ApplicationService
}

func NewConcertHandler(appService *concert.ApplicationService) *ConcertHandler {
    return &ConcertHandler{
        appService: appService,
    }
}

// CreateConcertRequest はコンサート作成リクエスト
type CreateConcertRequest struct {
    Title        string `json:"title" binding:"required"`
    VenueName    string `json:"venue_name" binding:"required"`
    VenueAddress string `json:"venue_address"`
    Capacity     int    `json:"capacity" binding:"required,min=1"`
    StartTime    string `json:"start_time" binding:"required"`
}

// UpdateConcertRequest はコンサート更新リクエスト
type UpdateConcertRequest struct {
    Title        *string `json:"title"`
    VenueName    *string `json:"venue_name"`
    VenueAddress *string `json:"venue_address"`
    Capacity     *int    `json:"capacity"`
    StartTime    *string `json:"start_time"`
}

// AddIdolRequest はアイドル追加リクエスト
type AddIdolRequest struct {
    IdolID string `json:"idol_id" binding:"required"`
}

// CreateConcert はコンサートを作成
func (h *ConcertHandler) CreateConcert(c *gin.Context) {
    var req CreateConcertRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    cmd := concert.CreateConcertCommand{
        Title:        req.Title,
        VenueName:    req.VenueName,
        VenueAddress: req.VenueAddress,
        Capacity:     req.Capacity,
        StartTime:    req.StartTime,
    }

    dto, err := h.appService.CreateConcert(c.Request.Context(), cmd)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, dto)
}

// GetConcert はコンサートを取得
func (h *ConcertHandler) GetConcert(c *gin.Context) {
    id := c.Param("id")

    query := concert.GetConcertQuery{ID: id}

    dto, err := h.appService.GetConcert(c.Request.Context(), query)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, dto)
}

// ListConcerts はコンサート一覧を取得
func (h *ConcertHandler) ListConcerts(c *gin.Context) {
    // クエリパラメータでフィルタリング
    idolID := c.Query("idol_id")

    var query concert.ListConcertsQuery
    if idolID != "" {
        query.IdolID = &idolID
    }

    dtos, err := h.appService.ListConcerts(c.Request.Context(), query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, dtos)
}

// AddIdolToConcert はコンサートにアイドルを追加
func (h *ConcertHandler) AddIdolToConcert(c *gin.Context) {
    concertID := c.Param("id")

    var req AddIdolRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    cmd := concert.AddIdolToConcertCommand{
        ConcertID: concertID,
        IdolID:    req.IdolID,
    }

    err := h.appService.AddIdolToConcert(c.Request.Context(), cmd)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "アイドルが追加されました"})
}

// 他のエンドポイント(Update, Delete等)も同様に実装...
```

---

### ステップ5: アダプターを作る（必要な場合のみ）

usecase の Output Port（インターフェース）を application サービスが直接実装していない場合、`cmd/api/adapters/` にアダプターを作成して適合させます。

```bash
# ファイル: cmd/api/adapters/concert_adapter.go
```

```go
package adapters

import (
    "context"

    appConcert "github.com/kuro48/idol-api/internal/application/concert"
    "github.com/kuro48/idol-api/internal/domain/concert"
)

// ConcertAppAdapter は ApplicationService を usecase の Output Port に適合させるアダプター
type ConcertAppAdapter struct {
    svc *appConcert.ApplicationService
}

func NewConcertAppAdapter(svc *appConcert.ApplicationService) *ConcertAppAdapter {
    return &ConcertAppAdapter{svc: svc}
}

func (a *ConcertAppAdapter) FindConcert(ctx context.Context, id string) (*concert.Concert, error) {
    return a.svc.FindByID(ctx, id)
}
```

**ポイント**:
- `cmd/api/adapters/` は Composition Root の一部なので、interface 層・infrastructure 層の両方を参照可能
- application サービスのメソッドシグネチャが usecase の Output Port と一致する場合はアダプター不要
- アダプターは薄いラッパーにとどめ、ビジネスロジックを書かない

---

### ステップ6: main.goに登録

```bash
# ファイル: cmd/api/main.go
```

```go
// 既存のimportに追加
import (
    // ... 既存のimport
    "github.com/kuro48/idol-api/internal/application/concert"
    // ... 既存のimport
)

func main() {
    // ... 既存のコード

    // リポジトリ
    idolRepo := mongodb.NewIdolRepository(db.Database)
    removalRepo := mongodb.NewRemovalRepository(db.Database)
    groupRepo := mongodb.NewGroupRepository(db.Database)
    concertRepo := mongodb.NewConcertRepository(db.Database)  // 追加

    // アプリケーションサービス
    idolAppService := idol.NewApplicationService(idolRepo)
    removalAppService := removal.NewApplicationService(removalRepo, idolRepo)
    groupAppService := group.NewApplicationService(groupRepo)
    concertAppService := concert.NewApplicationService(concertRepo)  // 追加

    // ハンドラー
    idolHandler := handlers.NewIdolHandler(idolAppService)
    removalHandler := handlers.NewRemovalHandler(removalAppService)
    groupHandler := handlers.NewGroupHandler(groupAppService)
    concertHandler := handlers.NewConcertHandler(concertAppService)  // 追加

    // ... 既存のルーティング

    v1 := router.Group("/api/v1")
    {
        // ... 既存のルート

        // コンサートルート（追加）
        concerts := v1.Group("/concerts")
        {
            concerts.POST("", concertHandler.CreateConcert)
            concerts.GET("", concertHandler.ListConcerts)
            concerts.GET("/:id", concertHandler.GetConcert)
            concerts.POST("/:id/idols", concertHandler.AddIdolToConcert)
            // concerts.PUT("/:id", concertHandler.UpdateConcert)
            // concerts.DELETE("/:id", concertHandler.DeleteConcert)
        }
    }

    // ... 既存のコード
}
```

---

## 📝 ファイルの命名規則

### ドメイン層
```
internal/domain/{ドメイン名}/
├── {エンティティ名}.go           例: idol.go, concert.go
├── {エンティティ名}_id.go        例: idol_id.go
├── {値オブジェクト名}.go         例: idol_name.go, birthdate.go
├── repository.go                (リポジトリインターフェース)
├── service.go                   (ドメインサービス)
└── error.go                     (エラー定義)
```

### アプリケーション層
```
internal/application/{ドメイン名}/
├── service.go                   (アプリケーションサービス)
├── command.go                   (コマンドDTO)
└── query.go                     (クエリDTO)
```

### インフラ層
```
internal/infrastructure/
├── database/
│   └── mongodb.go               (DB接続)
└── persistence/
    └── mongodb/
        └── {エンティティ名}_repository.go
            例: idol_repository.go, concert_repository.go
```

### プレゼンテーション層
```
internal/interface/handlers/
└── {エンティティ名}_handler.go
    例: idol_handler.go, concert_handler.go
```

---

## 🔒 コーディング規約

### エラーハンドリング

**goroutine 内のエラーは必ずログに記録する**

```go
// ❌ 悪い例
go func() {
    _, err := collection.InsertOne(ctx, doc)
    _ = err  // エラーを握り潰している
}()

// ✅ 良い例
go func() {
    _, err := collection.InsertOne(ctx, doc)
    if err != nil {
        slog.Error("ドキュメントの保存に失敗しました", "error", err)
    }
}()
```

### goroutine 内のコンテキスト

**goroutine 内で DB 操作する場合は必ずタイムアウトを設定する**

```go
// ❌ 悪い例（タイムアウトなし）
go func() {
    ctx := context.Background()
    _ = repo.Save(ctx, entity)
}()

// ✅ 良い例
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := repo.Save(ctx, entity); err != nil {
        slog.Error("保存エラー", "error", err)
    }
}()
```

### ロギング

**`log.Printf` / `fmt.Println` を使わず `log/slog` を使う**

```go
// ❌ 悪い例
log.Printf("エラー: %v", err)
fmt.Println("処理完了")

// ✅ 良い例
slog.Error("処理に失敗しました", "error", err)
slog.Info("処理が完了しました", "id", id)
```

### セキュリティ

**タイミング攻撃対策には `crypto/subtle.ConstantTimeCompare` を使う**

```go
// ❌ 悪い例（タイミング攻撃に脆弱）
if apiKey == expectedKey {
    // ...
}

// ✅ 良い例
import "crypto/subtle"

if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
    // 認証失敗
}
```

### 境界ルール（CI で自動チェック）

| レイヤー | 参照可能なパッケージ | 禁止 |
|---------|---------------------|------|
| `domain` | 標準ライブラリのみ | 外部パッケージすべて |
| `application` | `domain` のみ | それ以外すべて |
| `infrastructure` | `domain` のみ | `application` / `usecase` / `interface` |
| `usecase` | `application`, `domain` | `interface` / `infrastructure` |
| `interface` | `usecase` のみ | `infrastructure` 直接参照（middleware 除く） |
| `cmd/api/adapters` | interface・infrastructure 両方 | なし（Composition Root） |

> **注**: `internal/infrastructure/adapters/` は削除済み。アダプターは `cmd/api/adapters/` に配置してください。

---

## ⚠️ よくある間違いと回避方法

### ❌ 間違い1: ドメイン層で外部パッケージを使う

```go
// ❌ 悪い例
package idol

import "go.mongodb.org/mongo-driver/bson"  // MongoDB依存!

type Idol struct {
    ID bson.ObjectID  // MongoDBに依存している
    Name string
}
```

```go
// ✅ 良い例
package idol

type Idol struct {
    id   IdolID      // 自分の値オブジェクト
    name IdolName
}
```

**理由**: ドメイン層は技術に依存してはいけない。MongoDBからPostgreSQLに変えたときに、ドメイン層まで変更が必要になってしまう。

---

### ❌ 間違い2: エンティティの検証をハンドラーでやる

```go
// ❌ 悪い例（ハンドラー）
func (h *IdolHandler) CreateIdol(c *gin.Context) {
    var req CreateIdolRequest
    c.ShouldBindJSON(&req)

    // ハンドラーでビジネスルールをチェック
    if req.Name == "" {
        c.JSON(400, gin.H{"error": "名前は必須です"})
        return
    }
    if len(req.Name) > 100 {
        c.JSON(400, gin.H{"error": "名前は100文字以内です"})
        return
    }
    // ...
}
```

```go
// ✅ 良い例（値オブジェクト）
package idol

type IdolName struct {
    value string
}

func NewIdolName(name string) (IdolName, error) {
    if name == "" {
        return IdolName{}, errors.New("名前は必須です")
    }
    if len(name) > 100 {
        return IdolName{}, errors.New("名前は100文字以内です")
    }
    return IdolName{value: name}, nil
}
```

**理由**: ビジネスルールはドメイン層に集約すべき。複数の場所で同じチェックをすると、変更時に全部直す必要がある。

---

### ❌ 間違い3: アプリケーションサービスにビジネスロジックを書く

```go
// ❌ 悪い例
func (s *ApplicationService) CreateIdol(ctx context.Context, cmd CreateIdolCommand) error {
    // アプリケーションサービスにビジネスロジック
    if cmd.Birthdate != nil {
        age := calculateAge(*cmd.Birthdate)
        if age < 13 {
            return errors.New("13歳未満は登録できません")
        }
    }
    // ...
}
```

```go
// ✅ 良い例
// ドメイン層
func (b *Birthdate) IsValidForIdol() error {
    age := b.Age()
    if age < 13 {
        return errors.New("13歳未満は登録できません")
    }
    return nil
}

// アプリケーションサービス
func (s *ApplicationService) CreateIdol(ctx context.Context, cmd CreateIdolCommand) error {
    birthdate, err := idol.NewBirthdateFromString(*cmd.Birthdate)
    if err != nil {
        return err
    }

    // ドメインロジックを呼ぶだけ
    if err := birthdate.IsValidForIdol(); err != nil {
        return err
    }
    // ...
}
```

**理由**: アプリケーションサービスは「オーケストレーション」（指揮）だけをする。ビジネスルールはドメイン層に置く。

---

### ❌ 間違い4: 直接データベースの構造をレスポンスで返す

```go
// ❌ 悪い例
func (h *IdolHandler) GetIdol(c *gin.Context) {
    // MongoDBのドキュメント構造をそのまま返す
    var result bson.M
    collection.FindOne(ctx, filter).Decode(&result)
    c.JSON(200, result)  // 内部構造が漏れる
}
```

```go
// ✅ 良い例
func (h *IdolHandler) GetIdol(c *gin.Context) {
    dto, err := h.appService.GetIdol(ctx, query)
    c.JSON(200, dto)  // DTOを返す
}
```

**理由**: 内部のデータ構造を外部に公開すると、後で変更できなくなる。DTOを使って、外部向けの形式を明確にする。

---

### ❌ 間違い5: 集約を跨いだ参照

```go
// ❌ 悪い例
type Concert struct {
    id     ConcertID
    idols  []*idol.Idol  // 他の集約のエンティティを直接持つ
}
```

```go
// ✅ 良い例
type Concert struct {
    id      ConcertID
    idolIDs []IdolID     // IDだけを持つ
}
```

**理由**: 集約は独立して整合性を保つべき。他の集約のエンティティを直接持つと、トランザクション境界が曖昧になる。

---

## 🎯 開発時のチェックリスト

新機能を追加する際は、このチェックリストを使って確認しましょう:

### ドメイン層
- [ ] エンティティは不変条件（invariant）を守っているか
- [ ] 値オブジェクトはimmutableか
- [ ] ビジネスルールはドメイン層にあるか
- [ ] 外部パッケージ（DB、Web）に依存していないか
- [ ] リポジトリはインターフェースだけか

### アプリケーション層
- [ ] トランザクション境界は適切か
- [ ] DTOで外部とのやり取りをしているか
- [ ] ビジネスロジックを書いていないか（ドメイン層を呼ぶだけ）
- [ ] エラーハンドリングは適切か

### インフラ層
- [ ] リポジトリインターフェースを実装しているか
- [ ] ドメインモデルとドキュメントの変換は正しいか
- [ ] エラーメッセージは分かりやすいか

### プレゼンテーション層
- [ ] HTTPステータスコードは適切か
- [ ] バリデーションは最小限か（ビジネスルールはドメイン層）
- [ ] リクエスト/レスポンスの構造は明確か

---

## 📚 開発の順序まとめ

新機能を追加するときは、この順序で進めましょう:

```
1. ドメイン層
   └─ 値オブジェクト → エンティティ → リポジトリIF → ドメインサービス

2. アプリケーション層
   └─ コマンド/クエリ → ApplicationService

3. インフラ層
   └─ リポジトリ実装

4. プレゼンテーション層
   └─ ハンドラー（インターフェース経由）

5. アダプター（必要な場合のみ）
   └─ cmd/api/adapters/ に Output Port アダプターを作成

6. 統合
   └─ main.goに登録 → テスト
```

---

## 💡 まとめ

### 開発のポイント
1. **ドメイン層から始める** - ビジネスロジックが最も重要
2. **依存の方向を守る** - ドメイン層は誰にも依存しない
3. **1ファイル1責任** - ファイルを細かく分ける
4. **値オブジェクトを活用** - 不正な値を作れないようにする
5. **DTOで境界を明確に** - 内部構造を外に漏らさない

### こんなときどうする?

**Q: 既存のエンティティに新しいプロパティを追加したい**
A: ドメイン層のエンティティ → 値オブジェクト → リポジトリ実装 → DTO の順で追加

**Q: 新しいAPIエンドポイントを追加したい**
A: まず「何をしたいか」（ユースケース）を考える → ドメイン層から実装

**Q: 複雑なビジネスルールがある**
A: ドメインサービスを使う。単一エンティティで完結しない場合に有効

**Q: 2つの集約をまたがる処理がしたい**
A: アプリケーションサービスで協調させる。ただしIDで参照する（エンティティを直接持たない）

---

このガイドを参考に、クリーンで保守しやすいコードを書いていきましょう! 🚀
