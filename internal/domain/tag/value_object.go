package tag

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// TagID はタグの一意な識別子
type TagID struct {
	value string
}

// NewTagID は新しいTagIDを生成する
func NewTagID(value string) (TagID, error) {
	if value == "" {
		return TagID{}, errors.New("タグIDは必須です")
	}

	// UUIDの形式チェック
	if _, err := uuid.Parse(value); err != nil {
		return TagID{}, errors.New("無効なタグIDフォーマットです")
	}

	return TagID{value: value}, nil
}

// GenerateTagID は新しいUUID v7ベースのTagIDを生成する
func GenerateTagID() TagID {
	return TagID{value: uuid.NewString()}
}

// String はTagIDの文字列表現を返す
func (t TagID) String() string {
	return t.value
}

// Equals は2つのTagIDが等しいか判定する
func (t TagID) Equals(other TagID) bool {
	return t.value == other.value
}

// TagName はタグ名
type TagName struct {
	value string
}

// NewTagName は新しいTagNameを生成する
func NewTagName(value string) (TagName, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return TagName{}, errors.New("タグ名は必須です")
	}

	if len(trimmed) > 50 {
		return TagName{}, errors.New("タグ名は50文字以内である必要があります")
	}

	return TagName{value: trimmed}, nil
}

// String はTagNameの文字列表現を返す
func (t TagName) String() string {
	return t.value
}

// Equals は2つのTagNameが等しいか判定する
func (t TagName) Equals(other TagName) bool {
	return t.value == other.value
}

// TagCategory はタグのカテゴリ
type TagCategory string

const (
	// TagCategoryGenre はジャンル（例：アイドル、声優、モデル）
	TagCategoryGenre TagCategory = "genre"

	// TagCategoryRegion は地域（例：関東、関西、海外）
	TagCategoryRegion TagCategory = "region"

	// TagCategoryStyle はスタイル（例：かわいい系、クール系）
	TagCategoryStyle TagCategory = "style"

	// TagCategoryOther はその他
	TagCategoryOther TagCategory = "other"
)

// NewTagCategory は新しいTagCategoryを生成する
func NewTagCategory(value string) (TagCategory, error) {
	category := TagCategory(value)

	switch category {
	case TagCategoryGenre, TagCategoryRegion, TagCategoryStyle, TagCategoryOther:
		return category, nil
	default:
		return "", errors.New("無効なタグカテゴリです。genre, region, style, other のいずれかを指定してください")
	}
}

// String はTagCategoryの文字列表現を返す
func (t TagCategory) String() string {
	return string(t)
}

// IsValid はTagCategoryが有効か判定する
func (t TagCategory) IsValid() bool {
	switch t {
	case TagCategoryGenre, TagCategoryRegion, TagCategoryStyle, TagCategoryOther:
		return true
	default:
		return false
	}
}
