package group

import (
	"errors"
)

// GroupName はグループ名（正式名称・読み仮名・ローマ字表記を含む）
type GroupName struct {
	value string
	kana  *string // 読み仮名（ひらがな/カタカナ）
	latin *string // ローマ字/英語表記
}

func NewGroupName(value string) (GroupName, error) {
	if value == "" {
		return GroupName{value: ""}, nil
	}
	if len(value) > 100 {
		return GroupName{}, errors.New("グループ名は100文字以内である必要があります")
	}
	return GroupName{value: value}, nil
}

// NewGroupNameFull は読み仮名・ローマ字表記を含む名前を生成する
func NewGroupNameFull(value string, kana, latin *string) (GroupName, error) {
	name, err := NewGroupName(value)
	if err != nil {
		return GroupName{}, err
	}
	if kana != nil && len(*kana) > 200 {
		return GroupName{}, errors.New("読み仮名は200文字以内である必要があります")
	}
	if latin != nil && len(*latin) > 200 {
		return GroupName{}, errors.New("ローマ字表記は200文字以内である必要があります")
	}
	name.kana = kana
	name.latin = latin
	return name, nil
}

func (g GroupName) Value() string  { return g.value }
func (g GroupName) Kana() *string  { return g.kana }
func (g GroupName) Latin() *string { return g.latin }
func (g GroupName) IsEmpty() bool  { return g.value == "" }
