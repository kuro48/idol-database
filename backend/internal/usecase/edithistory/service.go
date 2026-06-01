package edithistory

import (
	"context"
	"time"

	domain "github.com/kuro48/idol-api/internal/domain/edithistory"
)

// Usecase は編集履歴のユースケース
type Usecase struct {
	appService EditHistoryAppPort
}

// NewUsecase はユースケースを作成する
func NewUsecase(appService EditHistoryAppPort) *Usecase {
	return &Usecase{appService: appService}
}

// GetEditHistory は編集履歴を取得する
func (u *Usecase) GetEditHistory(ctx context.Context, query GetEditHistoryQuery) (*EditHistoryDTO, error) {
	h, err := u.appService.GetHistory(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	dto := toDTO(h)
	return &dto, nil
}

// ListEditHistory は編集履歴一覧を取得する
func (u *Usecase) ListEditHistory(ctx context.Context, query ListEditHistoryQuery) (*EditHistoryListResult, error) {
	criteria := toCriteria(query)

	entries, total, err := u.appService.SearchHistory(ctx, criteria)
	if err != nil {
		return nil, err
	}

	dtos := make([]*EditHistoryDTO, 0, len(entries))
	for _, h := range entries {
		dto := toDTO(h)
		dtos = append(dtos, &dto)
	}

	return &EditHistoryListResult{
		Data:  dtos,
		Total: total,
		Page:  *query.Page,
		Limit: *query.Limit,
	}, nil
}

func toCriteria(query ListEditHistoryQuery) domain.SearchCriteria {
	criteria := domain.SearchCriteria{
		EntityID:  query.EntityID,
		ChangedBy: query.ChangedBy,
		Offset:    (*query.Page - 1) * *query.Limit,
		Limit:     *query.Limit,
	}
	if query.EntityType != nil {
		et, err := domain.NewEntityType(*query.EntityType)
		if err == nil {
			criteria.EntityType = &et
		}
	}
	if query.Action != nil {
		a, err := domain.NewAction(*query.Action)
		if err == nil {
			criteria.Action = &a
		}
	}
	return criteria
}

func toDTO(h *domain.EditHistory) EditHistoryDTO {
	changes := make(map[string]FieldChangeDTO, len(h.Changes()))
	for field, fc := range h.Changes() {
		changes[field] = FieldChangeDTO{
			Before: fc.Before,
			After:  fc.After,
		}
	}
	return EditHistoryDTO{
		ID:         h.ID().Value(),
		EntityType: h.EntityType().Value(),
		EntityID:   h.EntityID(),
		Action:     h.Action().Value(),
		Changes:    changes,
		ChangedBy:  h.ChangedBy(),
		CreatedAt:  h.CreatedAt().Format(time.RFC3339),
	}
}
