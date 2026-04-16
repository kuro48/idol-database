package submission_test

import (
	"testing"

	"github.com/kuro48/idol-api/internal/domain/submission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ヘルパー ---

func newValidEmail(t *testing.T) submission.ContributorEmail {
	t.Helper()
	email, err := submission.NewContributorEmail("test@example.com")
	require.NoError(t, err)
	return email
}

func newValidSourceURLs(t *testing.T) []submission.SourceURL {
	t.Helper()
	u, err := submission.NewSourceURL("https://example.com")
	require.NoError(t, err)
	return []submission.SourceURL{u}
}

func newPendingSubmission(t *testing.T) *submission.Submission {
	t.Helper()
	return submission.NewSubmission(
		submission.SubmissionTypeIdol,
		`{"name":"テスト"}`,
		newValidSourceURLs(t),
		newValidEmail(t),
	)
}

// --- NewSubmission ---

func TestNewSubmission(t *testing.T) {
	s := newPendingSubmission(t)

	assert.Equal(t, submission.SubmissionTypeIdol, s.TargetType())
	assert.Equal(t, `{"name":"テスト"}`, s.Payload())
	assert.Equal(t, submission.StatusPending, s.Status())
	assert.Empty(t, s.ID().Value())
	assert.Empty(t, s.ReviewedBy())
	assert.Nil(t, s.ReviewedAt())
	assert.Empty(t, s.RevisionNote())
	assert.False(t, s.CreatedAt().IsZero())
	assert.False(t, s.UpdatedAt().IsZero())
}

// --- Approve ---

func TestSubmission_Approve(t *testing.T) {
	t.Run("pending → approved に遷移できる", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Approve("admin1"))

		assert.Equal(t, submission.StatusApproved, s.Status())
		assert.Equal(t, "admin1", s.ReviewedBy())
		assert.NotNil(t, s.ReviewedAt())
		assert.True(t, s.IsApproved())
	})

	t.Run("approved 状態からは承認できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Approve("admin1"))

		err := s.Approve("admin2")
		assert.Error(t, err)
		assert.Equal(t, submission.StatusApproved, s.Status())
	})

	t.Run("rejected 状態からは承認できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Reject("admin1"))

		err := s.Approve("admin2")
		assert.Error(t, err)
	})

	t.Run("needs_revision 状態からは承認できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.RequestRevision("admin1", "修正してください"))

		err := s.Approve("admin2")
		assert.Error(t, err)
	})
}

// --- Reject ---

func TestSubmission_Reject(t *testing.T) {
	t.Run("pending → rejected に遷移できる", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Reject("admin1"))

		assert.Equal(t, submission.StatusRejected, s.Status())
		assert.Equal(t, "admin1", s.ReviewedBy())
		assert.NotNil(t, s.ReviewedAt())
		assert.True(t, s.IsRejected())
	})

	t.Run("approved 状態からは却下できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Approve("admin1"))

		err := s.Reject("admin2")
		assert.Error(t, err)
	})

	t.Run("needs_revision 状態からは却下できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.RequestRevision("admin1", "修正してください"))

		err := s.Reject("admin2")
		assert.Error(t, err)
	})
}

// --- RequestRevision ---

func TestSubmission_RequestRevision(t *testing.T) {
	t.Run("pending → needs_revision に遷移できる", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.RequestRevision("admin1", "参照元URLが不正です"))

		assert.Equal(t, submission.StatusNeedsRevision, s.Status())
		assert.Equal(t, "admin1", s.ReviewedBy())
		assert.NotNil(t, s.ReviewedAt())
		assert.Equal(t, "参照元URLが不正です", s.RevisionNote())
		assert.True(t, s.NeedsRevision())
	})

	t.Run("approved 状態からは差し戻しできない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Approve("admin1"))

		err := s.RequestRevision("admin2", "差し戻し")
		assert.Error(t, err)
	})
}

// --- Resubmit ---

func TestSubmission_Resubmit(t *testing.T) {
	t.Run("needs_revision → pending に戻れる", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.RequestRevision("admin1", "修正してください"))
		require.NoError(t, s.Resubmit())

		assert.Equal(t, submission.StatusPending, s.Status())
		assert.Empty(t, s.RevisionNote()) // revision_note はクリアされる
		assert.True(t, s.IsPending())
	})

	t.Run("pending 状態からは再投稿できない", func(t *testing.T) {
		s := newPendingSubmission(t)

		err := s.Resubmit()
		assert.Error(t, err)
	})

	t.Run("approved 状態からは再投稿できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Approve("admin1"))

		err := s.Resubmit()
		assert.Error(t, err)
	})

	t.Run("rejected 状態からは再投稿できない", func(t *testing.T) {
		s := newPendingSubmission(t)
		require.NoError(t, s.Reject("admin1"))

		err := s.Resubmit()
		assert.Error(t, err)
	})
}

// --- 状態確認メソッド ---

func TestSubmission_StatusCheckers(t *testing.T) {
	tests := []struct {
		name              string
		setup             func(*submission.Submission)
		wantPending       bool
		wantApproved      bool
		wantRejected      bool
		wantNeedsRevision bool
	}{
		{
			name:        "新規作成は pending",
			setup:       func(s *submission.Submission) {},
			wantPending: true,
		},
		{
			name:         "承認後は approved",
			setup:        func(s *submission.Submission) { _ = s.Approve("admin") },
			wantApproved: true,
		},
		{
			name:         "却下後は rejected",
			setup:        func(s *submission.Submission) { _ = s.Reject("admin") },
			wantRejected: true,
		},
		{
			name:              "差し戻し後は needs_revision",
			setup:             func(s *submission.Submission) { _ = s.RequestRevision("admin", "note") },
			wantNeedsRevision: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newPendingSubmission(t)
			tt.setup(s)
			assert.Equal(t, tt.wantPending, s.IsPending())
			assert.Equal(t, tt.wantApproved, s.IsApproved())
			assert.Equal(t, tt.wantRejected, s.IsRejected())
			assert.Equal(t, tt.wantNeedsRevision, s.NeedsRevision())
		})
	}
}
