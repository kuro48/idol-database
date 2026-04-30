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

type notifyingRemovalAppStub struct {
	createResult *RemovalCreateResult
	request      *domainRemoval.RemovalRequest
	updated      bool
}

func (s *notifyingRemovalAppStub) CreateRemovalRequest(context.Context, RemovalCreateInput) (*RemovalCreateResult, error) {
	return s.createResult, nil
}

func (s *notifyingRemovalAppStub) GetRemovalRequest(context.Context, string) (*domainRemoval.RemovalRequest, error) {
	return s.request, nil
}

func (s *notifyingRemovalAppStub) ListAllRemovalRequests(context.Context) ([]*domainRemoval.RemovalRequest, error) {
	return nil, nil
}

func (s *notifyingRemovalAppStub) ListPendingRemovalRequests(context.Context) ([]*domainRemoval.RemovalRequest, error) {
	return nil, nil
}

func (s *notifyingRemovalAppStub) UpdateRemovalRequest(context.Context, *domainRemoval.RemovalRequest) error {
	s.updated = true
	return nil
}

type notifyingRemovalIdolStub struct {
	deletedID string
}

func (s *notifyingRemovalIdolStub) GetIdol(context.Context, string) (*domainIdol.Idol, error) {
	return nil, nil
}

func (s *notifyingRemovalIdolStub) DeleteIdol(_ context.Context, id string) error {
	s.deletedID = id
	return nil
}

type notifyingRemovalGroupStub struct{}

func (s *notifyingRemovalGroupStub) GetGroup(context.Context, string) (*domainGroup.Group, error) {
	return nil, nil
}

func (s *notifyingRemovalGroupStub) DeleteGroup(context.Context, string) error {
	return nil
}

type removalNotifierStub struct {
	received []ReceivedNotification
	resolved []ResolvedNotification
}

func (s *removalNotifierStub) NotifyReceived(_ context.Context, notification ReceivedNotification) error {
	s.received = append(s.received, notification)
	return nil
}

func (s *removalNotifierStub) NotifyResolved(_ context.Context, notification ResolvedNotification) error {
	s.resolved = append(s.resolved, notification)
	return nil
}

func TestCreateRemovalRequest_SendsReceivedNotificationAndSLA(t *testing.T) {
	t.Parallel()

	request := newPendingRemovalRequest(t, domainRemoval.TargetTypeIdol, "idol-1")
	app := &notifyingRemovalAppStub{
		createResult: &RemovalCreateResult{
			Request:     request,
			AccessToken: "access-token",
		},
	}
	notifier := &removalNotifierStub{}
	uc := NewUsecase(app, &notifyingRemovalIdolStub{}, &notifyingRemovalGroupStub{}, notifier, nil)

	result, err := uc.CreateRemovalRequest(context.Background(), CreateRemovalRequestCommand{
		TargetType:    "idol",
		TargetID:      "idol-1",
		RequesterType: "third_party",
		Reason:        "権利侵害のため削除を申請します",
		ContactInfo:   "owner@example.com",
		Evidence:      "https://example.com/evidence",
		Description:   "公開継続に問題があるため詳細を記載します",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, notifier.received, 1)
	assert.Equal(t, "owner@example.com", notifier.received[0].To)
	assert.Equal(t, request.ID().Value(), notifier.received[0].RequestID)
	assert.False(t, result.RemovalRequest.SLAOverdue)
	assert.WithinDuration(t, request.CreatedAt().Add(removalSLAWindow), result.RemovalRequest.SLADueAt, time.Second)
}

func TestUpdateStatus_SendsResolvedNotification(t *testing.T) {
	t.Parallel()

	request := newPendingRemovalRequest(t, domainRemoval.TargetTypeIdol, "idol-1")
	app := &notifyingRemovalAppStub{request: request}
	idolApp := &notifyingRemovalIdolStub{}
	notifier := &removalNotifierStub{}
	uc := NewUsecase(app, idolApp, &notifyingRemovalGroupStub{}, notifier, nil)

	_, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:     request.ID().Value(),
		Status: "approved",
	})

	require.NoError(t, err)
	require.True(t, app.updated)
	require.Equal(t, "idol-1", idolApp.deletedID)
	require.Len(t, notifier.resolved, 1)
	assert.Equal(t, "owner@example.com", notifier.resolved[0].To)
	assert.Equal(t, request.ID().Value(), notifier.resolved[0].RequestID)
	assert.Equal(t, "approved", notifier.resolved[0].Status)
}
