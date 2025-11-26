package removal

import (
	"errors"
	"regexp"
)

// RequesterType は申請者のタイプ
type RequesterType string

const (
	RequesterIdolThemself RequesterType = "idol_themself" // アイドル本人
	RequesterAgency       RequesterType = "agency"        // 事務所
	RequesterThirdParty   RequesterType = "third_party"   // 第三者
)

// TargetType は削除対象のタイプ
type TargetType string

const (
	TargetTypeIdol  TargetType = "idol"
	TargetTypeGroup TargetType = "group"
)

// NewTargetType は新しいターゲットタイプを作成する
func NewTargetType(targetType string) (TargetType, error) {
	tt := TargetType(targetType)
	switch tt {
	case TargetTypeIdol, TargetTypeGroup:
		return tt, nil
	default:
		return "", errors.New("無効なターゲットタイプです")
	}
}

// Requester は申請者情報
type Requester struct {
	requesterType RequesterType
}

// NewRequester は新しい申請者を作成する
func NewRequester(requesterType string) (Requester, error) {
	rt := RequesterType(requesterType)

	switch rt {
	case RequesterIdolThemself, RequesterAgency, RequesterThirdParty:
		return Requester{requesterType: rt}, nil
	default:
		return Requester{}, errors.New("無効な申請者タイプです")
	}
}

// Type は申請者タイプを返す
func (r Requester) Type() RequesterType {
	return r.requesterType
}

// RemovalReason は削除理由
type RemovalReason struct {
	value string
}

// NewRemovalReason は新しい削除理由を作成する
func NewRemovalReason(value string) (RemovalReason, error) {
	if value == "" {
		return RemovalReason{}, errors.New("削除理由は必須です")
	}

	if len(value) < 10 {
		return RemovalReason{}, errors.New("削除理由は10文字以上で入力してください")
	}

	if len(value) > 1000 {
		return RemovalReason{}, errors.New("削除理由は1000文字以内で入力してください")
	}

	return RemovalReason{value: value}, nil
}

// Value は削除理由の値を返す
func (r RemovalReason) Value() string {
	return r.value
}

// ContactInfo は連絡先情報
type ContactInfo struct {
	value string
}

// NewContactInfo は新しい連絡先情報を作成する
func NewContactInfo(email string) (ContactInfo, error) {
	if email == "" {
		return ContactInfo{}, errors.New("連絡先メールアドレスは必須です")
	}

	// 簡易的なメールアドレスバリデーション
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ContactInfo{}, errors.New("有効なメールアドレスを入力してください")
	}

	return ContactInfo{value: email}, nil
}

// Value は連絡先情報の値を返す
func (c ContactInfo) Value() string {
	return c.value
}

// EvidenceURL は証拠資料のURL
type EvidenceURL struct {
	value string
}

// NewEvidenceURL は新しい証拠資料URLを作成する（オプショナル）
func NewEvidenceURL(url string) (EvidenceURL, error) {
	// 空の場合は許可（オプショナル）
	if url == "" {
		return EvidenceURL{}, nil
	}

	// URLの形式チェック
	urlRegex := regexp.MustCompile(`^https?://[^\s]+$`)
	if !urlRegex.MatchString(url) {
		return EvidenceURL{}, errors.New("有効なURLを入力してください")
	}

	return EvidenceURL{value: url}, nil
}

// Value は証拠資料URLの値を返す
func (e EvidenceURL) Value() string {
	return e.value
}

// IsEmpty は証拠資料URLが空かどうかをチェック
func (e EvidenceURL) IsEmpty() bool {
	return e.value == ""
}

// RemovalStatus は削除申請のステータス
type RemovalStatus string

const (
	StatusPending  RemovalStatus = "pending"  // 保留中
	StatusApproved RemovalStatus = "approved" // 承認済み
	StatusRejected RemovalStatus = "rejected" // 却下
)

// NewRemovalStatus は新しいステータスを作成する
func NewRemovalStatus(status string) (RemovalStatus, error) {
	rs := RemovalStatus(status)

	switch rs {
	case StatusPending, StatusApproved, StatusRejected:
		return rs, nil
	default:
		return "", errors.New("無効なステータスです")
	}
}
