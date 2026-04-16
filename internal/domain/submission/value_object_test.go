package submission_test

import (
	"testing"

	"github.com/kuro48/idol-api/internal/domain/submission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NewSubmissionStatus ---

func TestNewSubmissionStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    submission.SubmissionStatus
		wantErr bool
	}{
		{"pending", "pending", submission.StatusPending, false},
		{"approved", "approved", submission.StatusApproved, false},
		{"rejected", "rejected", submission.StatusRejected, false},
		{"needs_revision", "needs_revision", submission.StatusNeedsRevision, false},
		{"無効な値", "unknown", "", true},
		{"空文字", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := submission.NewSubmissionStatus(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// --- NewSubmissionType ---

func TestNewSubmissionType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    submission.SubmissionType
		wantErr bool
	}{
		{"idol", "idol", submission.SubmissionTypeIdol, false},
		{"group", "group", submission.SubmissionTypeGroup, false},
		{"agency", "agency", submission.SubmissionTypeAgency, false},
		{"event", "event", submission.SubmissionTypeEvent, false},
		{"無効な値", "unknown", "", true},
		{"空文字", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := submission.NewSubmissionType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// --- NewSourceURL ---

func TestNewSourceURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"https URL", "https://example.com", false},
		{"http URL", "http://example.com/path", false},
		{"パスあり", "https://example.com/idol/123", false},
		{"クエリあり", "https://example.com/search?q=test", false},
		{"空文字", "", true},
		{"http/https でない", "ftp://example.com", true},
		{"スキームなし", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := submission.NewSourceURL(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.input, got.Value())
			}
		})
	}
}

// --- NewContributorEmail ---

func TestNewContributorEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"有効なメールアドレス", "user@example.com", false},
		{"サブドメインあり", "user@mail.example.co.jp", false},
		{"プラス記号あり", "user+tag@example.com", false},
		{"空文字", "", true},
		{"@がない", "notanemail", true},
		{"ドメインなし", "user@", true},
		{"ローカル部なし", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := submission.NewContributorEmail(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.input, got.Value())
			}
		})
	}
}
