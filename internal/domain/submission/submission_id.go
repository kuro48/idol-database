package submission

import (
	"errors"
	"regexp"
)

// objectIDPattern は MongoDB ObjectID 互換の24文字16進数パターン
var objectIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

// SubmissionID は投稿審査の一意識別子
type SubmissionID struct {
	value string
}

// NewSubmissionID は新しい投稿審査IDを生成する
func NewSubmissionID(value string) (SubmissionID, error) {
	if value == "" {
		return SubmissionID{}, errors.New("投稿審査IDは空にできません")
	}

	if !objectIDPattern.MatchString(value) {
		return SubmissionID{}, errors.New("無効な投稿審査ID形式です")
	}

	return SubmissionID{value: value}, nil
}

// Value はIDの値を返す
func (id SubmissionID) Value() string {
	return id.value
}

// Equals は2つのIDが等しいかチェック
func (id SubmissionID) Equals(other SubmissionID) bool {
	return id.value == other.value
}
