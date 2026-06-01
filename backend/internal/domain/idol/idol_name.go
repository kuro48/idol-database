package idol

import (
	"errors"
)

// IdolName はアイドルの名前（正式名称・読み仮名・ローマ字表記を含む）
type IdolName struct {
	value string
	kana  *string // 読み仮名（ひらがな/カタカナ）
	latin *string // ローマ字/英語表記
}

func NewIdolName(value string) (IdolName, error) {
	if value == "" {
		return IdolName{}, errors.New("アイドル名は空にできません")
	}
	if len(value) > 100 {
		return IdolName{}, errors.New("アイドル名は100文字以内である必要があります")
	}
	return IdolName{value: value}, nil
}

// NewIdolNameFull は読み仮名・ローマ字表記を含む名前を生成する
func NewIdolNameFull(value string, kana, latin *string) (IdolName, error) {
	name, err := NewIdolName(value)
	if err != nil {
		return IdolName{}, err
	}
	if kana != nil && len(*kana) > 200 {
		return IdolName{}, errors.New("読み仮名は200文字以内である必要があります")
	}
	if latin != nil && len(*latin) > 200 {
		return IdolName{}, errors.New("ローマ字表記は200文字以内である必要があります")
	}
	name.kana = kana
	name.latin = latin
	return name, nil
}

func (n IdolName) Value() string  { return n.value }
func (n IdolName) Kana() *string  { return n.kana }
func (n IdolName) Latin() *string { return n.latin }
