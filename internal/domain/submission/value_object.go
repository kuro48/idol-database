package submission

import (
	"errors"
	"regexp"
)

// SubmissionStatus は投稿審査のステータス
type SubmissionStatus string

const (
	// StatusPending は審査待ちステータス
	StatusPending SubmissionStatus = "pending"
	// StatusApproved は承認済みステータス
	StatusApproved SubmissionStatus = "approved"
	// StatusRejected は却下済みステータス
	StatusRejected SubmissionStatus = "rejected"
	// StatusNeedsRevision は差し戻しステータス
	StatusNeedsRevision SubmissionStatus = "needs_revision"
)

// NewSubmissionStatus は新しいステータスを作成する
func NewSubmissionStatus(status string) (SubmissionStatus, error) {
	ss := SubmissionStatus(status)
	switch ss {
	case StatusPending, StatusApproved, StatusRejected, StatusNeedsRevision:
		return ss, nil
	default:
		return "", errors.New("無効なステータスです")
	}
}

// SubmissionType は投稿対象のタイプ
type SubmissionType string

const (
	// SubmissionTypeIdol はアイドル投稿タイプ
	SubmissionTypeIdol SubmissionType = "idol"
	// SubmissionTypeGroup はグループ投稿タイプ
	SubmissionTypeGroup SubmissionType = "group"
	// SubmissionTypeAgency は事務所投稿タイプ
	SubmissionTypeAgency SubmissionType = "agency"
	// SubmissionTypeEvent はイベント投稿タイプ
	SubmissionTypeEvent SubmissionType = "event"
)

// NewSubmissionType は新しい投稿タイプを作成する
func NewSubmissionType(targetType string) (SubmissionType, error) {
	st := SubmissionType(targetType)
	switch st {
	case SubmissionTypeIdol, SubmissionTypeGroup, SubmissionTypeAgency, SubmissionTypeEvent:
		return st, nil
	default:
		return "", errors.New("無効な投稿タイプです")
	}
}

// urlPattern は https?:// で始まるURLパターン
var urlPattern = regexp.MustCompile(`^https?://[^\s]+$`)

// emailPattern は簡易メールアドレスバリデーションパターン
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// SourceURL は参照元URLの値オブジェクト
type SourceURL struct {
	value string
}

// NewSourceURL は新しい参照元URLを作成する
func NewSourceURL(url string) (SourceURL, error) {
	if url == "" {
		return SourceURL{}, errors.New("参照元URLは必須です")
	}

	if !urlPattern.MatchString(url) {
		return SourceURL{}, errors.New("参照元URLは https:// または http:// で始まる有効なURLである必要があります")
	}

	return SourceURL{value: url}, nil
}

// Value は参照元URLの値を返す
func (s SourceURL) Value() string {
	return s.value
}

// ContributorEmail は投稿者メールアドレスの値オブジェクト
type ContributorEmail struct {
	value string
}

// NewContributorEmail は新しい投稿者メールアドレスを作成する
func NewContributorEmail(email string) (ContributorEmail, error) {
	if email == "" {
		return ContributorEmail{}, errors.New("投稿者メールアドレスは必須です")
	}

	if !emailPattern.MatchString(email) {
		return ContributorEmail{}, errors.New("有効なメールアドレスを入力してください")
	}

	return ContributorEmail{value: email}, nil
}

// Value は投稿者メールアドレスの値を返す
func (c ContributorEmail) Value() string {
	return c.value
}
