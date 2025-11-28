package agency

import (
	"errors"
	"strings"
)

// AgencyID は事務所ID値オブジェクト
type AgencyID struct {
	value string
}

// NewAgencyID は事務所IDを生成する
func NewAgencyID(value string) (AgencyID, error) {
	if value == "" {
		return AgencyID{}, errors.New("事務所IDは空にできません")
	}
	return AgencyID{value: value}, nil
}

// Value はIDの値を返す
func (id AgencyID) Value() string {
	return id.value
}

// AgencyName は事務所名値オブジェクト
type AgencyName struct {
	value string
}

// NewAgencyName は事務所名を生成する
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

// Value は事務所名の値を返す
func (n AgencyName) Value() string {
	return n.value
}

// Country は国値オブジェクト
type Country struct {
	value string
}

// NewCountry は国を生成する
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

// Value は国の値を返す
func (c Country) Value() string {
	return c.value
}

// contains はスライスに要素が含まれているかチェック
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
