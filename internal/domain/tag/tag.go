package tag

import (
	"errors"
	"strings"
	"time"
)

// Tag はタグエンティティ
type Tag struct {
	id          TagID
	name        TagName
	category    TagCategory
	description string
	createdAt   time.Time
}

// NewTag は新しいTagエンティティを作成する
func NewTag(name string, category string, description string) (*Tag, error) {
	tagName, err := NewTagName(name)
	if err != nil {
		return nil, err
	}

	tagCategory, err := NewTagCategory(category)
	if err != nil {
		return nil, err
	}

	// 説明文のバリデーション
	trimmedDesc := strings.TrimSpace(description)
	if len(trimmedDesc) > 200 {
		return nil, errors.New("説明は200文字以内である必要があります")
	}

	return &Tag{
		id:          GenerateTagID(),
		name:        tagName,
		category:    tagCategory,
		description: trimmedDesc,
		createdAt:   time.Now(),
	}, nil
}

// Reconstruct は既存のデータからTagエンティティを再構築する
func Reconstruct(
	id string,
	name string,
	category string,
	description string,
	createdAt time.Time,
) (*Tag, error) {
	tagID, err := NewTagID(id)
	if err != nil {
		return nil, err
	}

	tagName, err := NewTagName(name)
	if err != nil {
		return nil, err
	}

	tagCategory, err := NewTagCategory(category)
	if err != nil {
		return nil, err
	}

	return &Tag{
		id:          tagID,
		name:        tagName,
		category:    tagCategory,
		description: description,
		createdAt:   createdAt,
	}, nil
}

// ID はタグIDを返す
func (t *Tag) ID() TagID {
	return t.id
}

// Name はタグ名を返す
func (t *Tag) Name() TagName {
	return t.name
}

// Category はタグカテゴリを返す
func (t *Tag) Category() TagCategory {
	return t.category
}

// Description は説明を返す
func (t *Tag) Description() string {
	return t.description
}

// CreatedAt は作成日時を返す
func (t *Tag) CreatedAt() time.Time {
	return t.createdAt
}

// SetID はIDを設定する（永続化後に使用）
func (t *Tag) SetID(id TagID) {
	t.id = id
}

// UpdateName はタグ名を更新する
func (t *Tag) UpdateName(name string) error {
	tagName, err := NewTagName(name)
	if err != nil {
		return err
	}

	t.name = tagName
	return nil
}

// UpdateCategory はタグカテゴリを更新する
func (t *Tag) UpdateCategory(category string) error {
	tagCategory, err := NewTagCategory(category)
	if err != nil {
		return err
	}

	t.category = tagCategory
	return nil
}

// UpdateDescription は説明を更新する
func (t *Tag) UpdateDescription(description string) error {
	trimmedDesc := strings.TrimSpace(description)
	if len(trimmedDesc) > 200 {
		return errors.New("説明は200文字以内である必要があります")
	}

	t.description = trimmedDesc
	return nil
}
