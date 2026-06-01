package adapters

import (
	"context"

	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appEvent "github.com/kuro48/idol-api/internal/application/event"
	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	ucSubmission "github.com/kuro48/idol-api/internal/usecase/submission"
)

// SubmissionTargetAppAdapter は承認済み投稿を各アプリケーションサービスへ転送する。
type SubmissionTargetAppAdapter struct {
	idolSvc   *appIdol.ApplicationService
	groupSvc  *appGroup.ApplicationService
	agencySvc *appAgency.ApplicationService
	eventSvc  *appEvent.ApplicationService
}

// NewSubmissionTargetAppAdapter は SubmissionTargetAppAdapter を生成する。
func NewSubmissionTargetAppAdapter(
	idolSvc *appIdol.ApplicationService,
	groupSvc *appGroup.ApplicationService,
	agencySvc *appAgency.ApplicationService,
	eventSvc *appEvent.ApplicationService,
) ucSubmission.SubmissionTargetPort {
	return &SubmissionTargetAppAdapter{
		idolSvc:   idolSvc,
		groupSvc:  groupSvc,
		agencySvc: agencySvc,
		eventSvc:  eventSvc,
	}
}

func (a *SubmissionTargetAppAdapter) CreateIdol(ctx context.Context, input ucSubmission.IdolCreateInput) error {
	_, err := a.idolSvc.CreateIdol(ctx, appIdol.CreateInput{
		Name:      input.Name,
		Birthdate: input.Birthdate,
		AgencyID:  input.AgencyID,
		Aliases:   input.Aliases,
	})
	return err
}

func (a *SubmissionTargetAppAdapter) CreateGroup(ctx context.Context, input ucSubmission.GroupCreateInput) error {
	_, err := a.groupSvc.CreateGroup(ctx, appGroup.CreateInput{
		Name:          input.Name,
		FormationDate: input.FormationDate,
		DisbandDate:   input.DisbandDate,
	})
	return err
}

func (a *SubmissionTargetAppAdapter) CreateAgency(ctx context.Context, input ucSubmission.AgencyCreateInput) error {
	_, err := a.agencySvc.CreateAgency(ctx, appAgency.CreateInput{
		Name:            input.Name,
		NameEn:          input.NameEn,
		FoundedDate:     input.FoundedDate,
		Country:         input.Country,
		OfficialWebsite: input.OfficialWebsite,
		Description:     input.Description,
		LogoURL:         input.LogoURL,
	})
	return err
}

func (a *SubmissionTargetAppAdapter) CreateEvent(ctx context.Context, input ucSubmission.EventCreateInput) error {
	performers := make([]appEvent.PerformerInput, 0, len(input.Performers))
	for _, p := range input.Performers {
		performers = append(performers, appEvent.PerformerInput{
			PerformerID:   p.PerformerID,
			BillingStatus: p.BillingStatus,
		})
	}
	_, err := a.eventSvc.CreateEvent(ctx, appEvent.CreateInput{
		Title:         input.Title,
		EventType:     input.EventType,
		StartDateTime: input.StartDateTime,
		EndDateTime:   input.EndDateTime,
		VenueID:       input.VenueID,
		Performers:    performers,
		TicketURL:     input.TicketURL,
		OfficialURL:   input.OfficialURL,
		Description:   input.Description,
		Tags:          input.Tags,
	})
	return err
}
