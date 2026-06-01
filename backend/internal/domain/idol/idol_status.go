package idol

import "errors"

// IdolStatus はアイドルの活動状態
type IdolStatus string

const (
	IdolStatusActive    IdolStatus = "active"
	IdolStatusGraduated IdolStatus = "graduated"
	IdolStatusHiatus    IdolStatus = "hiatus"
	IdolStatusRetired   IdolStatus = "retired"
)

func NewIdolStatus(s string) (IdolStatus, error) {
	switch IdolStatus(s) {
	case IdolStatusActive, IdolStatusGraduated, IdolStatusHiatus, IdolStatusRetired:
		return IdolStatus(s), nil
	}
	return "", errors.New("無効なアイドルステータスです: active / graduated / hiatus / retired のいずれかを指定してください")
}

func (s IdolStatus) IsValid() bool {
	_, err := NewIdolStatus(string(s))
	return err == nil
}
