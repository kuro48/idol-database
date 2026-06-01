package edithistory

import (
	"context"

	domain "github.com/kuro48/idol-api/internal/domain/edithistory"
)

// EditHistoryAppPort はユースケースがアプリケーションサービスに要求する契約
type EditHistoryAppPort interface {
	GetHistory(ctx context.Context, id string) (*domain.EditHistory, error)
	SearchHistory(ctx context.Context, criteria domain.SearchCriteria) ([]*domain.EditHistory, int64, error)
}
