package event

import (
	"errors"
	"strings"
)

// EventID はイベントID値オブジェクト
type EventID struct {
	value string
}

// NewEventID はイベントIDを生成する
func NewEventID(value string) (EventID, error) {
	if value == "" {
		return EventID{}, errors.New("イベントIDは空にできません")
	}
	return EventID{value: value}, nil
}

// Value はIDの値を返す
func (id EventID) Value() string {
	return id.value
}

// EventTitle はイベントタイトル値オブジェクト
type EventTitle struct {
	value string
}

// NewEventTitle はイベントタイトルを生成する
func NewEventTitle(value string) (EventTitle, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return EventTitle{}, errors.New("イベントタイトルは空にできません")
	}
	if len(trimmed) > 200 {
		return EventTitle{}, errors.New("イベントタイトルは200文字以内にしてください")
	}
	return EventTitle{value: trimmed}, nil
}

// Value はタイトルの値を返す
func (t EventTitle) Value() string {
	return t.value
}

// EventType はイベントタイプ値オブジェクト
type EventType struct {
	value string
}

// イベントタイプ定数
const (
	EventTypeLive       = "live"
	EventTypeHandshake  = "handshake"
	EventTypeRelease    = "release"
	EventTypeFanMeeting = "fan_meeting"
	EventTypeOnline     = "online"
)

// NewEventType はイベントタイプを生成する
func NewEventType(value string) (EventType, error) {
	validTypes := []string{
		EventTypeLive,
		EventTypeHandshake,
		EventTypeRelease,
		EventTypeFanMeeting,
		EventTypeOnline,
	}
	if !contains(validTypes, value) {
		return EventType{}, errors.New("無効なイベントタイプです")
	}
	return EventType{value: value}, nil
}

// Value はイベントタイプの値を返す
func (t EventType) Value() string {
	return t.value
}

// EventStatus はイベントの状態
type EventStatus string

const (
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusPostponed EventStatus = "postponed"
	EventStatusCompleted EventStatus = "completed"
)

func NewEventStatus(value string) (EventStatus, error) {
	switch EventStatus(value) {
	case EventStatusScheduled, EventStatusCancelled, EventStatusPostponed, EventStatusCompleted:
		return EventStatus(value), nil
	}
	return "", errors.New("無効なイベントステータスです")
}

func (s EventStatus) IsValid() bool {
	_, err := NewEventStatus(string(s))
	return err == nil
}

// BillingStatus はパフォーマーのビリングステータス
type BillingStatus string

const (
	BillingStatusHeadliner    BillingStatus = "headliner"
	BillingStatusSupport      BillingStatus = "support"
	BillingStatusSpecialGuest BillingStatus = "special_guest"
	BillingStatusUnknown      BillingStatus = "unknown"
)

func NewBillingStatus(value string) BillingStatus {
	switch BillingStatus(value) {
	case BillingStatusHeadliner, BillingStatusSupport, BillingStatusSpecialGuest:
		return BillingStatus(value)
	}
	return BillingStatusUnknown
}

// Performer はイベントへの出演者
type Performer struct {
	PerformerID   string
	BillingStatus BillingStatus
}

func NewPerformer(performerID string, billingStatus string) (Performer, error) {
	if performerID == "" {
		return Performer{}, errors.New("パフォーマーIDは空にできません")
	}
	return Performer{
		PerformerID:   performerID,
		BillingStatus: NewBillingStatus(billingStatus),
	}, nil
}

// contains はスライスに要素が含まれているかチェック
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
