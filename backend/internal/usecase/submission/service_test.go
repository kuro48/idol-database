package submission

import (
	"context"
	"errors"
	"testing"

	domain "github.com/kuro48/idol-api/internal/domain/submission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSubmissionApp struct {
	submission *domain.Submission
	updated    *domain.Submission
}

func (f *fakeSubmissionApp) CreateSubmission(context.Context, SubmissionCreateInput) (*SubmissionCreateResult, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeSubmissionApp) GetSubmission(_ context.Context, id string) (*domain.Submission, error) {
	if f.submission == nil || f.submission.ID().Value() != id {
		return nil, errors.New("not found")
	}
	return f.submission, nil
}

func (f *fakeSubmissionApp) ListAll(context.Context) ([]*domain.Submission, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeSubmissionApp) ListPending(context.Context) ([]*domain.Submission, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeSubmissionApp) UpdateSubmission(_ context.Context, submission *domain.Submission) error {
	f.updated = submission
	return nil
}

type fakeApprovedTargetPort struct {
	idolInput   *IdolCreateInput
	groupInput  *GroupCreateInput
	agencyInput *AgencyCreateInput
	eventInput  *EventCreateInput
	err         error
}

func (f *fakeApprovedTargetPort) CreateIdol(_ context.Context, input IdolCreateInput) error {
	if f.err != nil {
		return f.err
	}
	f.idolInput = &input
	return nil
}

func (f *fakeApprovedTargetPort) CreateGroup(_ context.Context, input GroupCreateInput) error {
	if f.err != nil {
		return f.err
	}
	f.groupInput = &input
	return nil
}

func (f *fakeApprovedTargetPort) CreateAgency(_ context.Context, input AgencyCreateInput) error {
	if f.err != nil {
		return f.err
	}
	f.agencyInput = &input
	return nil
}

func (f *fakeApprovedTargetPort) CreateEvent(_ context.Context, input EventCreateInput) error {
	if f.err != nil {
		return f.err
	}
	f.eventInput = &input
	return nil
}

func TestUpdateStatus_ApprovedCreatesIdol(t *testing.T) {
	t.Parallel()

	sub := newSubmissionForTest(t, domain.SubmissionTypeIdol, `{"name":"星野みく","birthdate":"2001-05-01","agency_id":"agency-1","aliases":["みく","Miku"]}`)
	app := &fakeSubmissionApp{submission: sub}
	targets := &fakeApprovedTargetPort{}
	uc := NewUsecase(app, targets, nil)

	dto, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:         sub.ID().Value(),
		Status:     "approved",
		ReviewedBy: "admin-1",
	})

	require.NoError(t, err)
	require.NotNil(t, targets.idolInput)
	assert.Equal(t, "星野みく", targets.idolInput.Name)
	require.NotNil(t, targets.idolInput.Birthdate)
	assert.Equal(t, "2001-05-01", *targets.idolInput.Birthdate)
	require.NotNil(t, targets.idolInput.AgencyID)
	assert.Equal(t, "agency-1", *targets.idolInput.AgencyID)
	assert.Equal(t, []string{"みく", "Miku"}, targets.idolInput.Aliases)
	require.NotNil(t, app.updated)
	assert.Equal(t, "approved", dto.Status)
}

func TestUpdateStatus_ApprovedCreatesGroup(t *testing.T) {
	t.Parallel()

	sub := newSubmissionForTest(t, domain.SubmissionTypeGroup, `{"name":"テストグループ","formation_date":"2020-01-01","disband_date":"2024-12-31"}`)
	app := &fakeSubmissionApp{submission: sub}
	targets := &fakeApprovedTargetPort{}
	uc := NewUsecase(app, targets, nil)

	_, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:         sub.ID().Value(),
		Status:     "approved",
		ReviewedBy: "admin-1",
	})

	require.NoError(t, err)
	require.NotNil(t, targets.groupInput)
	assert.Equal(t, "テストグループ", targets.groupInput.Name)
	require.NotNil(t, targets.groupInput.FormationDate)
	assert.Equal(t, "2020-01-01", *targets.groupInput.FormationDate)
	require.NotNil(t, targets.groupInput.DisbandDate)
	assert.Equal(t, "2024-12-31", *targets.groupInput.DisbandDate)
}

func TestUpdateStatus_ApprovedCreatesAgency(t *testing.T) {
	t.Parallel()

	sub := newSubmissionForTest(t, domain.SubmissionTypeAgency, `{"name":"テスト事務所","name_en":"Test Agency","founded_date":"2010-04-01","country":"JP","official_website":"https://agency.example.com","description":"紹介文","logo_url":"https://agency.example.com/logo.png"}`)
	app := &fakeSubmissionApp{submission: sub}
	targets := &fakeApprovedTargetPort{}
	uc := NewUsecase(app, targets, nil)

	_, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:         sub.ID().Value(),
		Status:     "approved",
		ReviewedBy: "admin-1",
	})

	require.NoError(t, err)
	require.NotNil(t, targets.agencyInput)
	assert.Equal(t, "テスト事務所", targets.agencyInput.Name)
	require.NotNil(t, targets.agencyInput.NameEn)
	assert.Equal(t, "Test Agency", *targets.agencyInput.NameEn)
	require.NotNil(t, targets.agencyInput.FoundedDate)
	assert.Equal(t, "2010-04-01", *targets.agencyInput.FoundedDate)
	assert.Equal(t, "JP", targets.agencyInput.Country)
	require.NotNil(t, targets.agencyInput.OfficialWebsite)
	assert.Equal(t, "https://agency.example.com", *targets.agencyInput.OfficialWebsite)
	require.NotNil(t, targets.agencyInput.Description)
	assert.Equal(t, "紹介文", *targets.agencyInput.Description)
	require.NotNil(t, targets.agencyInput.LogoURL)
	assert.Equal(t, "https://agency.example.com/logo.png", *targets.agencyInput.LogoURL)
}

func TestUpdateStatus_ApprovedCreatesEvent(t *testing.T) {
	t.Parallel()

	sub := newSubmissionForTest(t, domain.SubmissionTypeEvent, `{"title":"単独ライブ","event_type":"live","start_date_time":"2026-06-01T18:00:00+09:00","end_date_time":"2026-06-01T20:00:00+09:00","venue_id":"venue-1","performer_ids":["idol-1","group-1"],"ticket_url":"https://ticket.example.com","official_url":"https://event.example.com","description":"イベント説明","tags":["live","tour"]}`)
	app := &fakeSubmissionApp{submission: sub}
	targets := &fakeApprovedTargetPort{}
	uc := NewUsecase(app, targets, nil)

	_, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:         sub.ID().Value(),
		Status:     "approved",
		ReviewedBy: "admin-1",
	})

	require.NoError(t, err)
	require.NotNil(t, targets.eventInput)
	assert.Equal(t, "単独ライブ", targets.eventInput.Title)
	assert.Equal(t, "live", targets.eventInput.EventType)
	assert.Equal(t, "2026-06-01T18:00:00+09:00", targets.eventInput.StartDateTime)
	require.NotNil(t, targets.eventInput.EndDateTime)
	assert.Equal(t, "2026-06-01T20:00:00+09:00", *targets.eventInput.EndDateTime)
	require.NotNil(t, targets.eventInput.VenueID)
	assert.Equal(t, "venue-1", *targets.eventInput.VenueID)
	assert.Equal(t, []string{"idol-1", "group-1"}, targets.eventInput.PerformerIDs)
	require.NotNil(t, targets.eventInput.TicketURL)
	assert.Equal(t, "https://ticket.example.com", *targets.eventInput.TicketURL)
	require.NotNil(t, targets.eventInput.OfficialURL)
	assert.Equal(t, "https://event.example.com", *targets.eventInput.OfficialURL)
	require.NotNil(t, targets.eventInput.Description)
	assert.Equal(t, "イベント説明", *targets.eventInput.Description)
	assert.Equal(t, []string{"live", "tour"}, targets.eventInput.Tags)
}

func TestUpdateStatus_ApprovedDoesNotPersistWhenTargetCreationFails(t *testing.T) {
	t.Parallel()

	sub := newSubmissionForTest(t, domain.SubmissionTypeIdol, `{"name":"失敗ケース"}`)
	app := &fakeSubmissionApp{submission: sub}
	targets := &fakeApprovedTargetPort{err: errors.New("create failed")}
	uc := NewUsecase(app, targets, nil)

	dto, err := uc.UpdateStatus(context.Background(), UpdateStatusCommand{
		ID:         sub.ID().Value(),
		Status:     "approved",
		ReviewedBy: "admin-1",
	})

	require.Error(t, err)
	assert.Nil(t, dto)
	assert.Nil(t, app.updated)
}

func newSubmissionForTest(t *testing.T, targetType domain.SubmissionType, payload string) *domain.Submission {
	t.Helper()

	id, err := domain.NewSubmissionID("507f1f77bcf86cd799439011")
	require.NoError(t, err)
	sourceURL, err := domain.NewSourceURL("https://example.com/source")
	require.NoError(t, err)
	email, err := domain.NewContributorEmail("user@example.com")
	require.NoError(t, err)

	sub := domain.NewSubmission(targetType, payload, []domain.SourceURL{sourceURL}, email, "hashed-token")
	sub.SetID(id)
	return sub
}
