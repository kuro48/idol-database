package group

import (
	"context"
	"errors"
	"testing"

	domain "github.com/kuro48/idol-api/internal/domain/group"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type groupRepoStub struct {
	data map[string]*domain.Group
}

func newGroupRepoStub() *groupRepoStub {
	return &groupRepoStub{data: make(map[string]*domain.Group)}
}

func (r *groupRepoStub) Save(_ context.Context, g *domain.Group) error {
	if g.ID().Value() == "" {
		id, _ := domain.NewGroupID("group-1")
		g.SetID(id)
	}
	r.data[g.ID().Value()] = g
	return nil
}

func (r *groupRepoStub) FindByID(_ context.Context, id domain.GroupID) (*domain.Group, error) {
	group, ok := r.data[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return group, nil
}

func (r *groupRepoStub) FindAll(context.Context) ([]*domain.Group, error) {
	return nil, nil
}

func (r *groupRepoStub) FindWithPagination(context.Context, domain.SearchOptions) (*domain.SearchResult, error) {
	return nil, nil
}

func (r *groupRepoStub) Update(_ context.Context, g *domain.Group) error {
	r.data[g.ID().Value()] = g
	return nil
}

func (r *groupRepoStub) Delete(_ context.Context, id domain.GroupID) error {
	delete(r.data, id.Value())
	return nil
}

func (r *groupRepoStub) Restore(context.Context, domain.GroupID) error {
	return nil
}

func (r *groupRepoStub) ExistsByName(_ context.Context, name domain.GroupName) (bool, error) {
	for _, group := range r.data {
		if group.Name().Value() == name.Value() {
			return true, nil
		}
	}
	return false, nil
}

type groupWebhookPublisherStub struct {
	calls []struct {
		event   domainWebhook.EventType
		payload interface{}
	}
}

func (p *groupWebhookPublisherStub) Publish(_ context.Context, event domainWebhook.EventType, payload interface{}) error {
	p.calls = append(p.calls, struct {
		event   domainWebhook.EventType
		payload interface{}
	}{event: event, payload: payload})
	return nil
}

func TestApplicationService_PublishesWebhookOnCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	repo := newGroupRepoStub()
	publisher := &groupWebhookPublisherStub{}
	svc := NewApplicationService(repo, publisher)

	created, err := svc.CreateGroup(context.Background(), CreateInput{Name: "テストグループ"})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 1)
	assert.Equal(t, domainWebhook.EventGroupCreated, publisher.calls[0].event)

	payload, ok := publisher.calls[0].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "テストグループ", payload["name"])

	newName := "更新後グループ"
	err = svc.UpdateGroup(context.Background(), UpdateInput{ID: created.ID().Value(), Name: &newName})
	require.NoError(t, err)
	require.Len(t, publisher.calls, 2)
	assert.Equal(t, domainWebhook.EventGroupUpdated, publisher.calls[1].event)

	payload, ok = publisher.calls[1].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
	assert.Equal(t, "更新後グループ", payload["name"])

	err = svc.DeleteGroup(context.Background(), created.ID().Value())
	require.NoError(t, err)
	require.Len(t, publisher.calls, 3)
	assert.Equal(t, domainWebhook.EventGroupDeleted, publisher.calls[2].event)

	payload, ok = publisher.calls[2].payload.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, created.ID().Value(), payload["id"])
}
