package adapters

import (
	"context"

	appEditHistory "github.com/kuro48/idol-api/internal/application/edithistory"
	domainEditHistory "github.com/kuro48/idol-api/internal/domain/edithistory"
	ucEditHistory "github.com/kuro48/idol-api/internal/usecase/edithistory"
)

// EditHistoryAppAdapter は appEditHistory.ApplicationService を ucEditHistory.EditHistoryAppPort に適合させる
type EditHistoryAppAdapter struct {
	svc *appEditHistory.ApplicationService
}

// NewEditHistoryAppAdapter は EditHistoryAppAdapter を生成する
func NewEditHistoryAppAdapter(svc *appEditHistory.ApplicationService) ucEditHistory.EditHistoryAppPort {
	return &EditHistoryAppAdapter{svc: svc}
}

func (a *EditHistoryAppAdapter) GetHistory(ctx context.Context, id string) (*domainEditHistory.EditHistory, error) {
	return a.svc.GetHistory(ctx, id)
}

func (a *EditHistoryAppAdapter) SearchHistory(ctx context.Context, criteria domainEditHistory.SearchCriteria) ([]*domainEditHistory.EditHistory, int64, error) {
	return a.svc.SearchHistory(ctx, criteria)
}
