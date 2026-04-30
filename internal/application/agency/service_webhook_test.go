package agency

import (
	"context"
	"errors"
	"testing"

	domain "github.com/kuro48/idol-api/internal/domain/agency"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type agencyRepoStub struct {
	data map[string]*domain.Agency
}

func newAgencyRepoStub() *agencyRepoStub {
	return &agencyRepoStub{data: make(map[string]*domain.Agency)}
}

func (r *agencyRepoStub) Save(_ context.Context, agency *domain.Agency) error {
	r.data[agency.ID().Value()] = agency
	return nil
}

func (r *agencyRepoStub) FindByID(_ context.Context, id domain.AgencyID) (*domain.Agency, error) {
	agency, ok := r.data[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return agency, nil
}

func (r *agencyRepoStub) FindAll(context.Context) ([]*domain.Agency, error) {
	return nil, nil
}

func (r *agencyRepoStub) FindWithPagination(context.Context, domain.SearchOptions) (*domain.SearchResult, error) {
	return nil, nil
}

func (r *agencyRepoStub) Update(_ context.Context, agency *domain.Agency) error {
	r.data[agency.ID().Value()] = agency
	return nil
}

func (r *agencyRepoStub) Delete(_ context.Context, id domain.AgencyID) error {
	delete(r.data, id.Value())
	return nil
}

func (r *agencyRepoStub) Restore(context.Context, domain.AgencyID) error {
	return nil
}

func (r *agencyRepoStub) ExistsByID(_ context.Context, id domain.AgencyID) (bool, error) {
	_, ok := r.data[id.Value()]
	return ok, nil
}

func (r *agencyRepoStub) ExistsByName(_ context.Context, name domain.AgencyName) (bool, error) {
	for _, agency := range r.data {
		if agency.Name().Value() == name.Value() {
			return true, nil
		}
	}
	return false, nil
}

type agencyWebhookPublisherStub struct {
	calls []struct {
		event   domainWebhook.EventType
		payload interface{}
	}
}

func (p *agencyWebhookPublisherStub) Publish(_ context.Context, event domainWebhook.EventType, payload interface{}) error {
	p.calls = append(p.calls, struct {
		event   domainWebhook.EventType
		payload interface{}
	}{event: event, payload: payload})
	return nil
}

func TestApplicationService_PublishesWebhookOnCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	repo := newAgencyRepoStub()
	publisher := &agencyWebhookPublisherStub{}
	svc := NewApplicationService(repo, publisher)

	created, err := svc.CreateAgency(context.Background(), CreateInput{
		Name:    "テスト事務所",
		Country: "日本",
	})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 1)
	assert.Equal(t, domainWebhook.EventAgencyCreated, publisher.calls[0].event)

	payload, ok := publisher.calls[0].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "テスト事務所", payload["name"])

	newName := "更新後事務所"
	err = svc.UpdateAgency(context.Background(), UpdateInput{ID: created.ID().Value(), Name: &newName})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 2)
	assert.Equal(t, domainWebhook.EventAgencyUpdated, publisher.calls[1].event)

	payload, ok = publisher.calls[1].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "更新後事務所", payload["name"])

	err = svc.DeleteAgency(context.Background(), created.ID().Value())
	require.NoError(t, err)
	require.Len(t, publisher.calls, 3)
	assert.Equal(t, domainWebhook.EventAgencyDeleted, publisher.calls[2].event)

	payload, ok = publisher.calls[2].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
}
