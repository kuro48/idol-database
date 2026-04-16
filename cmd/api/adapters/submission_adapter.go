package adapters

import (
	"context"

	appSubmission "github.com/kuro48/idol-api/internal/application/submission"
	domainSubmission "github.com/kuro48/idol-api/internal/domain/submission"
	ucSubmission "github.com/kuro48/idol-api/internal/usecase/submission"
)

// SubmissionAppAdapter は appSubmission.ApplicationService を ucSubmission.SubmissionAppPort に適合させる
type SubmissionAppAdapter struct {
	svc *appSubmission.ApplicationService
}

// NewSubmissionAppAdapter は SubmissionAppAdapter を生成する
func NewSubmissionAppAdapter(svc *appSubmission.ApplicationService) ucSubmission.SubmissionAppPort {
	return &SubmissionAppAdapter{svc: svc}
}

func (a *SubmissionAppAdapter) CreateSubmission(ctx context.Context, input ucSubmission.SubmissionCreateInput) (*domainSubmission.Submission, error) {
	return a.svc.CreateSubmission(ctx, appSubmission.CreateInput{
		TargetType:       input.TargetType,
		Payload:          input.Payload,
		SourceURLs:       input.SourceURLs,
		ContributorEmail: input.ContributorEmail,
	})
}

func (a *SubmissionAppAdapter) GetSubmission(ctx context.Context, id string) (*domainSubmission.Submission, error) {
	return a.svc.GetSubmission(ctx, id)
}

func (a *SubmissionAppAdapter) ListAll(ctx context.Context) ([]*domainSubmission.Submission, error) {
	return a.svc.ListAll(ctx)
}

func (a *SubmissionAppAdapter) ListPending(ctx context.Context) ([]*domainSubmission.Submission, error) {
	return a.svc.ListPending(ctx)
}

func (a *SubmissionAppAdapter) UpdateSubmission(ctx context.Context, s *domainSubmission.Submission) error {
	return a.svc.UpdateSubmission(ctx, s)
}
