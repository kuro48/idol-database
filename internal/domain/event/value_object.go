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

// contains はスライスに要素が含まれているかチェック
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
