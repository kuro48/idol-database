package edithistory

import "time"

// EditHistory は編集履歴エンティティ（追記専用）
type EditHistory struct {
	id         EditHistoryID
	entityType EntityType
	entityID   string
	action     Action
	changes    map[string]FieldChange
	changedBy  string
	createdAt  time.Time
}

// NewEditHistory は新しい編集履歴を作成する
func NewEditHistory(
	entityType EntityType,
	entityID string,
	action Action,
	changes map[string]FieldChange,
	changedBy string,
) *EditHistory {
	return &EditHistory{
		entityType: entityType,
		entityID:   entityID,
		action:     action,
		changes:    changes,
		changedBy:  changedBy,
		createdAt:  time.Now(),
	}
}

// Reconstruct はデータストアから編集履歴を再構築する
func Reconstruct(
	id EditHistoryID,
	entityType EntityType,
	entityID string,
	action Action,
	changes map[string]FieldChange,
	changedBy string,
	createdAt time.Time,
) *EditHistory {
	return &EditHistory{
		id:         id,
		entityType: entityType,
		entityID:   entityID,
		action:     action,
		changes:    changes,
		changedBy:  changedBy,
		createdAt:  createdAt,
	}
}

func (h *EditHistory) ID() EditHistoryID               { return h.id }
func (h *EditHistory) EntityType() EntityType          { return h.entityType }
func (h *EditHistory) EntityID() string                { return h.entityID }
func (h *EditHistory) Action() Action                  { return h.action }
func (h *EditHistory) Changes() map[string]FieldChange { return h.changes }
func (h *EditHistory) ChangedBy() string               { return h.changedBy }
func (h *EditHistory) CreatedAt() time.Time            { return h.createdAt }

func (h *EditHistory) SetID(id EditHistoryID) { h.id = id }
