package removal

import (
	"context"
	"testing"
	"time"

	domainGroup "github.com/kuro48/idol-api/internal/domain/group"
	domainIdol "github.com/kuro48/idol-api/internal/domain/idol"
	domainRemoval "github.com/kuro48/idol-api/internal/domain/removal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type overdueRemovalAppStub struct {
	pending []*domainRemoval.RemovalRequest
}

func (s *overdueRemovalAppStub) CreateRemovalRequest(context.Context, RemovalCreateInput) (*RemovalCreateResult, error) {
	return nil, nil
}

func (s *overdueRemovalAppStub) GetRemovalRequest(context.Context, string) (*domainRemoval.RemovalRequest, error) {
	return nil, nil
}

func (s *overdueRemovalAppStub) ListAllRemovalRequests(context.Context) ([]*domainRemoval.RemovalRequest, error) {
	return nil, nil
}

func (s *overdueRemovalAppStub) ListPendingRemovalRequests(context.Context) ([]*domainRemoval.RemovalRequest, error) {
	return s.pending, nil
}

func (s *overdueRemovalAppStub) UpdateRemovalRequest(context.Context, *domainRemoval.RemovalRequest) error {
	return nil
}

type noopRemovalIdolStub struct{}

func (s *noopRemovalIdolStub) GetIdol(context.Context, string) (*domainIdol.Idol, error) { return nil, nil }
func (s *noopRemovalIdolStub) DeleteIdol(context.Context, string) error                    { return nil }

type noopRemovalGroupStub struct{}

func (s *noopRemovalGroupStub) GetGroup(context.Context, string) (*domainGroup.Group, error) { return nil, nil }
func (s *noopRemovalGroupStub) DeleteGroup(context.Context, string) error                      { return nil }

func TestListOverdueRemovalRequests_ReturnsOnlyOverduePending(t *testing.T) {
	t.Parallel()

	overdue := reconstructRemovalRequestAt(t, time.Now().Add(-(removalSLAWindow + time.Hour)), domainRemoval.StatusPending)
	recent := reconstructRemovalRequestAt(t, time.Now().Add(-time.Hour), domainRemoval.StatusPending)

	app := &overdueRemovalAppStub{
		pending: []*domainRemoval.RemovalRequest{overdue, recent},
	}
	uc := NewUsecase(app, &noopRemovalIdolStub{}, &noopRemovalGroupStub{}, nil, nil)

	result, err := uc.ListOverdueRemovalRequests(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, overdue.ID().Value(), result[0].ID)
	assert.True(t, result[0].SLAOverdue)
}

func reconstructRemovalRequestAt(t *testing.T, createdAt time.Time, status domainRemoval.RemovalStatus) *domainRemoval.RemovalRequest {
	t.Helper()

	id, err := domainRemoval.NewRemovalID("507f1f77bcf86cd799439012")
	require.NoError(t, err)
	requester, err := domainRemoval.NewRequester("third_party")
	require.NoError(t, err)
	reason, err := domainRemoval.NewRemovalReason("権利侵害のため削除を申請します")
	require.NoError(t, err)
	contact, err := domainRemoval.NewContactInfo("owner@example.com")
	require.NoError(t, err)
	evidence, err := domainRemoval.NewEvidenceURL("https://example.com/evidence")
	require.NoError(t, err)
	description, err := domainRemoval.NewRemovalReason("公開継続に問題があるため詳細を記載します")
	require.NoError(t, err)

	return domainRemoval.Reconstruct(
		id,
		"idol-1",
		domainRemoval.TargetTypeIdol,
		requester,
		reason,
		contact,
		"hashed-token",
		evidence,
		description,
		status,
		createdAt,
		createdAt,
	)
}
