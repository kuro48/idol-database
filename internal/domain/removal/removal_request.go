package removal

import (
	"time"
)

// RemovalRequest は削除申請のエンティティ（Aggregate Root）
type RemovalRequest struct {
	id          RemovalID
	targetID    string
	targetType  TargetType
	requester   Requester
	reason      RemovalReason
	contactInfo ContactInfo
	evidence    EvidenceURL
	description RemovalReason
	status      RemovalStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// NewRemovalRequest は新しい削除申請を作成する
func NewRemovalRequest(
	targetID string,
	targetType TargetType,
	requester Requester,
	reason RemovalReason,
	contactInfo ContactInfo,
	evidence EvidenceURL,
	description RemovalReason,
) *RemovalRequest {
	now := time.Now()

	return &RemovalRequest{
		// IDは空（保存時に生成される）
		targetID:    targetID,
		targetType:  targetType,
		requester:   requester,
		reason:      reason,
		contactInfo: contactInfo,
		evidence:    evidence,
		description: description,
		status:      StatusPending, // 初期状態は保留中
		createdAt:   now,
		updatedAt:   now,
	}
}

// Reconstruct は既存の削除申請を再構築する（リポジトリから取得時に使用）
func Reconstruct(
	id RemovalID,
	targetID string,
	targetType TargetType,
	requester Requester,
	reason RemovalReason,
	contactInfo ContactInfo,
	evidence EvidenceURL,
	description RemovalReason,
	status RemovalStatus,
	createdAt time.Time,
	updatedAt time.Time,
) *RemovalRequest {
	return &RemovalRequest{
		id:          id,
		targetID:    targetID,
		targetType:  targetType,
		requester:   requester,
		reason:      reason,
		contactInfo: contactInfo,
		evidence:    evidence,
		description: description,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ID は削除申請IDを返す
func (r *RemovalRequest) ID() RemovalID {
	return r.id
}

// TargetID は対象IDを返す
func (r *RemovalRequest) TargetID() string {
	return r.targetID
}

// TargetType は対象タイプを返す
func (r *RemovalRequest) TargetType() TargetType {
	return r.targetType
}

// Requester は申請者情報を返す
func (r *RemovalRequest) Requester() Requester {
	return r.requester
}

// Reason は削除理由を返す
func (r *RemovalRequest) Reason() RemovalReason {
	return r.reason
}

// ContactInfo は連絡先情報を返す
func (r *RemovalRequest) ContactInfo() ContactInfo {
	return r.contactInfo
}

// Evidence は証拠資料URLを返す
func (r *RemovalRequest) Evidence() EvidenceURL {
	return r.evidence
}

// Description は詳細説明を返す
func (r *RemovalRequest) Description() RemovalReason {
	return r.description
}

// Status はステータスを返す
func (r *RemovalRequest) Status() RemovalStatus {
	return r.status
}

// CreatedAt は作成日時を返す
func (r *RemovalRequest) CreatedAt() time.Time {
	return r.createdAt
}

// UpdatedAt は更新日時を返す
func (r *RemovalRequest) UpdatedAt() time.Time {
	return r.updatedAt
}

// SetID はIDを設定する（永続化後に使用）
func (r *RemovalRequest) SetID(id RemovalID) {
	r.id = id
}

// Approve は削除申請を承認する
func (r *RemovalRequest) Approve() error {
	if r.status != StatusPending {
		return NewDomainError("承認できるのは保留中の申請のみです")
	}

	r.status = StatusApproved
	r.updatedAt = time.Now()
	return nil
}

// Reject は削除申請を却下する
func (r *RemovalRequest) Reject() error {
	if r.status != StatusPending {
		return NewDomainError("却下できるのは保留中の申請のみです")
	}

	r.status = StatusRejected
	r.updatedAt = time.Now()
	return nil
}

// IsPending は保留中かどうかをチェック
func (r *RemovalRequest) IsPending() bool {
	return r.status == StatusPending
}

// IsApproved は承認済みかどうかをチェック
func (r *RemovalRequest) IsApproved() bool {
	return r.status == StatusApproved
}

// IsRejected は却下済みかどうかをチェック
func (r *RemovalRequest) IsRejected() bool {
	return r.status == StatusRejected
}
