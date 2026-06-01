package edithistory

import "errors"

// EditHistoryID は編集履歴ID値オブジェクト
type EditHistoryID struct {
	value string
}

func NewEditHistoryID(value string) (EditHistoryID, error) {
	if value == "" {
		return EditHistoryID{}, errors.New("編集履歴IDは空にできません")
	}
	return EditHistoryID{value: value}, nil
}

func (id EditHistoryID) Value() string { return id.value }

// EntityType は履歴対象のエンティティ種別
type EntityType string

const (
	EntityTypeIdol       EntityType = "idol"
	EntityTypeGroup      EntityType = "group"
	EntityTypeAgency     EntityType = "agency"
	EntityTypeEvent      EntityType = "event"
	EntityTypeRelease    EntityType = "release"
	EntityTypeVenue      EntityType = "venue"
	EntityTypeMembership EntityType = "membership"
)

func NewEntityType(value string) (EntityType, error) {
	switch EntityType(value) {
	case EntityTypeIdol, EntityTypeGroup, EntityTypeAgency, EntityTypeEvent,
		EntityTypeRelease, EntityTypeVenue, EntityTypeMembership:
		return EntityType(value), nil
	}
	return "", errors.New("無効なエンティティ種別です")
}

func (t EntityType) Value() string { return string(t) }

// Action は編集操作の種別
type Action string

const (
	ActionCreate  Action = "create"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionRestore Action = "restore"
)

func NewAction(value string) (Action, error) {
	switch Action(value) {
	case ActionCreate, ActionUpdate, ActionDelete, ActionRestore:
		return Action(value), nil
	}
	return "", errors.New("無効なアクションです")
}

func (a Action) Value() string { return string(a) }

// FieldChange はフィールドの変更内容
type FieldChange struct {
	Before interface{}
	After  interface{}
}
