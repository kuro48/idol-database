package idol

import (
	"context"
	"errors"
	"testing"

	domain "github.com/kuro48/idol-api/internal/domain/idol"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type idolRepoStub struct {
	data map[string]*domain.Idol
}

func newIdolRepoStub() *idolRepoStub {
	return &idolRepoStub{data: make(map[string]*domain.Idol)}
}

func (r *idolRepoStub) Save(_ context.Context, idol *domain.Idol) error {
	if idol.ID().Value() == "" {
		id, _ := domain.NewIdolID("idol-1")
		idol.SetID(id)
	}
	r.data[idol.ID().Value()] = idol
	return nil
}

func (r *idolRepoStub) FindByID(_ context.Context, id domain.IdolID) (*domain.Idol, error) {
	idol, ok := r.data[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return idol, nil
}

func (r *idolRepoStub) FindAll(context.Context) ([]*domain.Idol, error) {
	return nil, nil
}

func (r *idolRepoStub) Update(_ context.Context, idol *domain.Idol) error {
	r.data[idol.ID().Value()] = idol
	return nil
}

func (r *idolRepoStub) Delete(_ context.Context, id domain.IdolID) error {
	delete(r.data, id.Value())
	return nil
}

func (r *idolRepoStub) Restore(context.Context, domain.IdolID) error {
	return nil
}

func (r *idolRepoStub) ExistsByName(_ context.Context, name domain.IdolName) (bool, error) {
	for _, idol := range r.data {
		if idol.Name().Value() == name.Value() {
			return true, nil
		}
	}
	return false, nil
}

func (r *idolRepoStub) FindByAgencyID(context.Context, string) ([]*domain.Idol, error) {
	return nil, nil
}

func (r *idolRepoStub) Search(context.Context, domain.SearchCriteria) ([]*domain.Idol, error) {
	return nil, nil
}

func (r *idolRepoStub) Count(context.Context, domain.SearchCriteria) (int64, error) {
	return 0, nil
}

func (r *idolRepoStub) FindByExternalID(context.Context, domain.ExternalIDKind, string) (*domain.Idol, error) {
	return nil, nil
}

type webhookPublishCall struct {
	event   domainWebhook.EventType
	payload interface{}
}

type webhookPublisherStub struct {
	calls []webhookPublishCall
}

func (p *webhookPublisherStub) Publish(_ context.Context, event domainWebhook.EventType, payload interface{}) error {
	p.calls = append(p.calls, webhookPublishCall{event: event, payload: payload})
	return nil
}

func TestApplicationService_PublishesWebhookOnCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	repo := newIdolRepoStub()
	publisher := &webhookPublisherStub{}
	svc := NewApplicationService(repo, publisher)

	created, err := svc.CreateIdol(context.Background(), CreateInput{Name: "星野みく"})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 1)
	assert.Equal(t, domainWebhook.EventIdolCreated, publisher.calls[0].event)

	payload, ok := publisher.calls[0].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "星野みく", payload["name"])

	newName := "星野みく改"
	err = svc.UpdateIdol(context.Background(), UpdateInput{ID: created.ID().Value(), Name: &newName})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 2)
	assert.Equal(t, domainWebhook.EventIdolUpdated, publisher.calls[1].event)

	payload, ok = publisher.calls[1].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "星野みく改", payload["name"])

	err = svc.DeleteIdol(context.Background(), created.ID().Value())
	require.NoError(t, err)
	require.Len(t, publisher.calls, 3)
	assert.Equal(t, domainWebhook.EventIdolDeleted, publisher.calls[2].event)

	payload, ok = publisher.calls[2].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
}
