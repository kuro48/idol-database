# グループドメイン実装例

## value_object.go の完成形

```go
package group

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// GroupID はグループの一意識別子
type GroupID struct {
	value string
}

func NewGroupID(value string) (GroupID, error) {
	if value == "" {
		return GroupID{}, errors.New("グループIDは空にできません")
	}

	// MongoDBのObjectID形式チェック
	if _, err := bson.ObjectIDFromHex(value); err != nil {
		return GroupID{}, errors.New("無効なグループID形式です")
	}

	return GroupID{value: value}, nil
}

func (id GroupID) Value() string {
	return id.value
}

func (id GroupID) Equals(other GroupID) bool {
	return id.value == other.value
}

// GroupName はグループ名
type GroupName struct {
	value string
}

func NewGroupName(value string) (GroupName, error) {
	if value == "" {
		return GroupName{}, errors.New("グループ名は必須です")
	}

	if len(value) > 100 {
		return GroupName{}, errors.New("グループ名は100文字以内です")
	}

	return GroupName{value: value}, nil
}

func (n GroupName) Value() string {
	return n.value
}

// FormationDate は結成日
type FormationDate struct {
	value time.Time
}

func NewFormationDate(year, month, day int) (FormationDate, error) {
	// バリデーション
	if year < 1900 || year > 9999 {
		return FormationDate{}, errors.New("年は1900〜9999の範囲で指定してください")
	}
	if month < 1 || month > 12 {
		return FormationDate{}, errors.New("月は1〜12の範囲で指定してください")
	}
	if day < 1 || day > 31 {
		return FormationDate{}, errors.New("日は1〜31の範囲で指定してください")
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	// 未来の日付チェック
	if date.After(time.Now()) {
		return FormationDate{}, errors.New("結成日は未来の日付にできません")
	}

	return FormationDate{value: date}, nil
}

func (d FormationDate) Value() time.Time {
	return d.value
}

// MemberCount はメンバー数
type MemberCount struct {
	value int
}

func NewMemberCount(value int) (MemberCount, error) {
	if value < 1 {
		return MemberCount{}, errors.New("メンバー数は1人以上である必要があります")
	}

	if value > 1000 {
		return MemberCount{}, errors.New("メンバー数は1000人以下である必要があります")
	}

	return MemberCount{value: value}, nil
}

func (c MemberCount) Value() int {
	return c.value
}
```

## group.go の完成形

```go
package group

import (
	"errors"
	"time"
)

// Group はアイドルグループのエンティティ
type Group struct {
	id            GroupID
	name          GroupName
	formationDate FormationDate
	memberCount   MemberCount
	createdAt     time.Time
	updatedAt     time.Time
}

// NewGroup は新しいグループを作成する
func NewGroup(
	name GroupName,
	formationDate FormationDate,
	memberCount MemberCount,
) (*Group, error) {
	now := time.Now()

	return &Group{
		name:          name,
		formationDate: formationDate,
		memberCount:   memberCount,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// Reconstruct は永続化データからグループを再構築する
func Reconstruct(
	id GroupID,
	name GroupName,
	formationDate FormationDate,
	memberCount MemberCount,
	createdAt time.Time,
	updatedAt time.Time,
) *Group {
	return &Group{
		id:            id,
		name:          name,
		formationDate: formationDate,
		memberCount:   memberCount,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// Getters

func (g *Group) ID() GroupID {
	return g.id
}

func (g *Group) Name() GroupName {
	return g.name
}

func (g *Group) FormationDate() FormationDate {
	return g.formationDate
}

func (g *Group) MemberCount() MemberCount {
	return g.memberCount
}

func (g *Group) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Group) UpdatedAt() time.Time {
	return g.updatedAt
}

// ビジネスロジック

// UpdateName はグループ名を変更する
func (g *Group) UpdateName(name GroupName) {
	g.name = name
	g.updatedAt = time.Now()
}

// UpdateMemberCount はメンバー数を更新する
func (g *Group) UpdateMemberCount(count MemberCount) error {
	g.memberCount = count
	g.updatedAt = time.Now()
	return nil
}

// IsActive はグループが活動中かチェック
func (g *Group) IsActive() bool {
	// 例: 結成から30年以上経っている場合は非活動とみなす
	yearsSinceFormation := time.Since(g.formationDate.Value()).Hours() / (24 * 365)
	return yearsSinceFormation < 30
}

// GetActiveYears は活動年数を取得する
func (g *Group) GetActiveYears() int {
	years := time.Since(g.formationDate.Value()).Hours() / (24 * 365)
	return int(years)
}
```

## repository.go（追加が必要）

```go
package group

import "context"

// Repository はグループリポジトリのインターフェース
type Repository interface {
	Save(ctx context.Context, group *Group) error
	FindByID(ctx context.Context, id GroupID) (*Group, error)
	FindAll(ctx context.Context) ([]*Group, error)
	Update(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id GroupID) error

	// カスタムクエリ
	FindByName(ctx context.Context, name GroupName) (*Group, error)
	ExistsByName(ctx context.Context, name GroupName) (bool, error)
}
```

## error.go（追加が必要）

```go
package group

// DomainError はグループドメイン層のエラー
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

## 次に実装するファイル

1. **アプリケーション層**:
   - `internal/application/group/command.go`
   - `internal/application/group/query.go`
   - `internal/application/group/service.go`

2. **インフラ層**:
   - `internal/infrastructure/persistence/mongodb/group_repository.go`

3. **プレゼンテーション層**:
   - `internal/interface/handlers/group_handler.go`

4. **統合**:
   - `cmd/api/main.go` に依存性注入を追加
