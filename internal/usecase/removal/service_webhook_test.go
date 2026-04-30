package removal

import (
	"context"
	"errors"
	"testing"

	domainGroup "github.com/kuro48/idol-api/internal/domain/group"
	domainIdol "github.com/kuro48/idol-api/internal/domain/idol"
	domainRemoval "github.com/kuro48/idol-api/internal/domain/removal"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type removalAppStub struct {
	request *domainRemoval.RemovalRequest
	updated bool
}

func (s *removalAppStub) CreateRemovalRequest(context.Context, RemovalCreateInput) (*RemovalCreateResult, error) {
	return nil, errors.New("not implemented")
}

func (s *removalAppStub) GetRemovalRequest(context.Context, string) (*domainRemoval.RemovalRequest, error) {
	return s.request, nil
}

func (s *removalAppStub) ListAllRemovalRequests(context.Context) ([]*domainRemoval.RemovalRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *removalAppStub) ListPendingRemovalRequests(context.Context) ([]*domainRemoval.RemovalRequest, error) {
	return nil, errors.New("not implemented")
}

func (s *removalAppStub) UpdateRemovalRequest(context.Context, *domainRemoval.RemovalRequest) error {
	s.updated = true
	return nil
}

type removalIdolStub struct {
	deletedID string
}

func (s *removalIdolStub) GetIdol(context.Context, string) (*domainIdol.Idol, error) {
	return nil, nil
}

func (s *removalIdolStub) DeleteIdol(_ context.Context, id string) error {
	s.deletedID = id
	return nil
}

type removalGroupStub struct{}

func (s *removalGroupStub) GetGroup(context.Context, string) (*domainGroup.Group, error) {
	return nil, nil
}

func (s *removalGroupStub) DeleteGroup(context.Context, string) error {
	return nil
}

type removalWebhookPublisherStub struct {
	event   domainWebhook.EventType
	payload interface{}
	called  bool
}

func (s *removalWebhookPublisherStub) Publish(_ context.Context, event domainWebhook.EventType, payload interface{}) error {
	s.called = true
	s.event = event
	s.payload = payload
	return nil
}

func TestUpdateStatus_ApprovedPublishesWebhook(t *testing.T) {
	t.Parallel()

	request := newPendingRemovalRequest(t, domainRemoval.TargetTypeIdol, "idol-1")
	removalApp := &removalAppStub{request: request}
	idolApp := &removalIdolStub{}
	publisher := &removalWebhookPublisherStub{}
	uc := NewUsecase(removalApp, idolApp, &removalGroupStub{}, publisher)

	dto, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:     request.ID().Value(),
		Status: "approved",
	})

	require.NoError(t, err)
	require.NotNil(t, dto)
	assert.True(t, removalApp.updated)
	assert.Equal(t, "idol-1", idolApp.deletedID)
	assert.True(t, publisher.called)
	assert.Equal(t, domainWebhook.EventRemovalApproved, publisher.event)

	payload, ok := publisher.payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, request.ID().Value(), payload["id"])
	assert.Equal(t, "idol-1", payload["target_id"])
	assert.Equal(t, "idol", payload["target_type"])
	assert.Equal(t, "approved", payload["status"])
}

func newPendingRemovalRequest(t *testing.T, targetType domainRemoval.TargetType, targetID string) *domainRemoval.RemovalRequest {
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

	req := domainRemoval.NewRemovalRequest(targetID, targetType, requester, reason, contact, "hashed-token", evidence, description)
	req.SetID(id)
	return req
}
