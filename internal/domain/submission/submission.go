package submission

import (
	"time"
)

// Submission は投稿審査のエンティティ（Aggregate Root）
type Submission struct {
	id               SubmissionID
	targetType       SubmissionType
	payload          string // JSON文字列（投稿内容）
	sourceURLs       []SourceURL
	contributorEmail ContributorEmail
	snsUserID        string     // nullable（空文字で未設定）
	status           SubmissionStatus
	revisionNote     string     // 差し戻し理由（空文字で未設定）
	reviewedBy       string     // 審査者ID（空文字で未設定）
	reviewedAt       *time.Time // 審査日時（nilで未設定）
	createdAt        time.Time
	updatedAt        time.Time
}

// NewSubmission は新しい投稿審査を作成する
func NewSubmission(
	targetType SubmissionType,
	payload string,
	sourceURLs []SourceURL,
	contributorEmail ContributorEmail,
) *Submission {
	now := time.Now()

	return &Submission{
		// IDは空（保存時に生成される）
		targetType:       targetType,
		payload:          payload,
		sourceURLs:       sourceURLs,
		contributorEmail: contributorEmail,
		snsUserID:        "",
		status:           StatusPending, // 初期状態は審査待ち
		revisionNote:     "",
		reviewedBy:       "",
		reviewedAt:       nil,
		createdAt:        now,
		updatedAt:        now,
	}
}

// Reconstruct は既存の投稿審査を再構築する（リポジトリから取得時に使用）
func Reconstruct(
	id SubmissionID,
	targetType SubmissionType,
	payload string,
	sourceURLs []SourceURL,
	contributorEmail ContributorEmail,
	snsUserID string,
	status SubmissionStatus,
	revisionNote string,
	reviewedBy string,
	reviewedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Submission {
	return &Submission{
		id:               id,
		targetType:       targetType,
		payload:          payload,
		sourceURLs:       sourceURLs,
		contributorEmail: contributorEmail,
		snsUserID:        snsUserID,
		status:           status,
		revisionNote:     revisionNote,
		reviewedBy:       reviewedBy,
		reviewedAt:       reviewedAt,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}
}

// ID は投稿審査IDを返す
func (s *Submission) ID() SubmissionID {
	return s.id
}

// TargetType は対象タイプを返す
func (s *Submission) TargetType() SubmissionType {
	return s.targetType
}

// Payload は投稿内容のJSON文字列を返す
func (s *Submission) Payload() string {
	return s.payload
}

// SourceURLs は参照元URL一覧を返す
func (s *Submission) SourceURLs() []SourceURL {
	return s.sourceURLs
}

// ContributorEmail は投稿者メールアドレスを返す
func (s *Submission) ContributorEmail() ContributorEmail {
	return s.contributorEmail
}

// SnsUserID はSNSユーザーIDを返す（未設定の場合は空文字）
func (s *Submission) SnsUserID() string {
	return s.snsUserID
}

// Status はステータスを返す
func (s *Submission) Status() SubmissionStatus {
	return s.status
}

// RevisionNote は差し戻し理由を返す（未設定の場合は空文字）
func (s *Submission) RevisionNote() string {
	return s.revisionNote
}

// ReviewedBy は審査者IDを返す（未設定の場合は空文字）
func (s *Submission) ReviewedBy() string {
	return s.reviewedBy
}

// ReviewedAt は審査日時を返す（未設定の場合はnil）
func (s *Submission) ReviewedAt() *time.Time {
	return s.reviewedAt
}

// CreatedAt は作成日時を返す
func (s *Submission) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt は更新日時を返す
func (s *Submission) UpdatedAt() time.Time {
	return s.updatedAt
}

// SetID はIDを設定する（永続化後に使用）
func (s *Submission) SetID(id SubmissionID) {
	s.id = id
}

// Approve は投稿審査を承認する（pending のみ可）
func (s *Submission) Approve(reviewedBy string) error {
	if s.status != StatusPending {
		return NewDomainError("承認できるのは審査待ちの投稿のみです")
	}

	now := time.Now()
	s.status = StatusApproved
	s.reviewedBy = reviewedBy
	s.reviewedAt = &now
	s.updatedAt = now
	return nil
}

// Reject は投稿審査を却下する（pending のみ可）
func (s *Submission) Reject(reviewedBy string) error {
	if s.status != StatusPending {
		return NewDomainError("却下できるのは審査待ちの投稿のみです")
	}

	now := time.Now()
	s.status = StatusRejected
	s.reviewedBy = reviewedBy
	s.reviewedAt = &now
	s.updatedAt = now
	return nil
}

// RequestRevision は差し戻しを行う（pending のみ可）
func (s *Submission) RequestRevision(reviewedBy, note string) error {
	if s.status != StatusPending {
		return NewDomainError("差し戻しできるのは審査待ちの投稿のみです")
	}

	now := time.Now()
	s.status = StatusNeedsRevision
	s.reviewedBy = reviewedBy
	s.reviewedAt = &now
	s.revisionNote = note
	s.updatedAt = now
	return nil
}

// Resubmit は差し戻し後の再投稿を行う（needs_revision のみ可）
func (s *Submission) Resubmit() error {
	if s.status != StatusNeedsRevision {
		return NewDomainError("再投稿できるのは差し戻し中の投稿のみです")
	}

	s.status = StatusPending
	s.revisionNote = ""
	s.updatedAt = time.Now()
	return nil
}

// IsPending は審査待ちかどうかをチェック
func (s *Submission) IsPending() bool {
	return s.status == StatusPending
}

// IsApproved は承認済みかどうかをチェック
func (s *Submission) IsApproved() bool {
	return s.status == StatusApproved
}

// IsRejected は却下済みかどうかをチェック
func (s *Submission) IsRejected() bool {
	return s.status == StatusRejected
}

// NeedsRevision は差し戻し中かどうかをチェック
func (s *Submission) NeedsRevision() bool {
	return s.status == StatusNeedsRevision
}
