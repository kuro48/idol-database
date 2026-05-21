package event

import (
	"errors"
	"time"
)

// Event はイベント集約のルートエンティティ
type Event struct {
	id            EventID
	title         EventTitle
	eventType     EventType
	startDateTime time.Time
	endDateTime   *time.Time
	venueID       *string  // 会場ID（オプション）
	performerIDs  []string // アイドルまたはグループのID
	ticketURL     *string
	officialURL   *string
	description   *string
	tags          []string
	createdAt     time.Time
	updatedAt     time.Time
}

// NewEvent は新しいイベントを作成する
func NewEvent(
	title EventTitle,
	eventType EventType,
	startDateTime time.Time,
) (*Event, error) {
	now := time.Now()

	return &Event{
		title:         title,
		eventType:     eventType,
		startDateTime: startDateTime,
		performerIDs:  []string{},
		tags:          []string{},
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// Reconstruct はデータストアからイベントを再構築する（永続化層用）
func Reconstruct(
	id EventID,
	title EventTitle,
	eventType EventType,
	startDateTime time.Time,
	endDateTime *time.Time,
	venueID *string,
	performerIDs []string,
	ticketURL *string,
	officialURL *string,
	description *string,
	tags []string,
	createdAt time.Time,
	updatedAt time.Time,
) *Event {
	return &Event{
		id:            id,
		title:         title,
		eventType:     eventType,
		startDateTime: startDateTime,
		endDateTime:   endDateTime,
		venueID:       venueID,
		performerIDs:  performerIDs,
		ticketURL:     ticketURL,
		officialURL:   officialURL,
		description:   description,
		tags:          tags,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// ゲッター

func (e *Event) ID() EventID {
	return e.id
}

func (e *Event) Title() EventTitle {
	return e.title
}

func (e *Event) EventType() EventType {
	return e.eventType
}

func (e *Event) StartDateTime() time.Time {
	return e.startDateTime
}

func (e *Event) EndDateTime() *time.Time {
	return e.endDateTime
}

func (e *Event) VenueID() *string {
	return e.venueID
}

func (e *Event) PerformerIDs() []string {
	return e.performerIDs
}

func (e *Event) TicketURL() *string {
	return e.ticketURL
}

func (e *Event) OfficialURL() *string {
	return e.officialURL
}

func (e *Event) Description() *string {
	return e.description
}

func (e *Event) Tags() []string {
	return e.tags
}

func (e *Event) CreatedAt() time.Time {
	return e.createdAt
}

func (e *Event) UpdatedAt() time.Time {
	return e.updatedAt
}

// ビジネスロジック

// SetID はIDを設定する（永続化後に使用）
func (e *Event) SetID(id EventID) {
	e.id = id
}

// UpdateDetails はイベントの詳細を更新する
func (e *Event) UpdateDetails(
	title *EventTitle,
	startDateTime *time.Time,
	endDateTime *time.Time,
	venueID *string,
	ticketURL *string,
	officialURL *string,
	description *string,
) {
	if title != nil {
		e.title = *title
	}
	if startDateTime != nil {
		e.startDateTime = *startDateTime
	}
	if endDateTime != nil {
		e.endDateTime = endDateTime
	}
	if venueID != nil {
		e.venueID = venueID
	}
	if ticketURL != nil {
		e.ticketURL = ticketURL
	}
	if officialURL != nil {
		e.officialURL = officialURL
	}
	if description != nil {
		e.description = description
	}
	e.updatedAt = time.Now()
}

// AddPerformer はパフォーマーを追加する
func (e *Event) AddPerformer(performerID string) error {
	// 重複チェック
	for _, id := range e.performerIDs {
		if id == performerID {
			return errors.New("既に追加されています")
		}
	}
	e.performerIDs = append(e.performerIDs, performerID)
	e.updatedAt = time.Now()
	return nil
}

// RemovePerformer はパフォーマーを削除する
func (e *Event) RemovePerformer(performerID string) {
	for i, id := range e.performerIDs {
		if id == performerID {
			e.performerIDs = append(e.performerIDs[:i], e.performerIDs[i+1:]...)
			break
		}
	}
	e.updatedAt = time.Now()
}

// AddTag はタグを追加する
func (e *Event) AddTag(tag string) error {
	// 重複チェック
	for _, t := range e.tags {
		if t == tag {
			return errors.New("既に追加されています")
		}
	}
	e.tags = append(e.tags, tag)
	e.updatedAt = time.Now()
	return nil
}

// RemoveTag はタグを削除する
func (e *Event) RemoveTag(tag string) {
	for i, t := range e.tags {
		if t == tag {
			e.tags = append(e.tags[:i], e.tags[i+1:]...)
			break
		}
	}
	e.updatedAt = time.Now()
}

// IsUpcoming はイベントが今後開催されるかチェックする
func (e *Event) IsUpcoming() bool {
	return e.startDateTime.After(time.Now())
}

// IsPast はイベントが過去に開催されたかチェックする
func (e *Event) IsPast() bool {
	if e.endDateTime != nil {
		return e.endDateTime.Before(time.Now())
	}
	return e.startDateTime.Before(time.Now())
}

// IsOngoing はイベントが現在開催中かチェックする
func (e *Event) IsOngoing() bool {
	now := time.Now()
	if e.endDateTime != nil {
		return e.startDateTime.Before(now) && e.endDateTime.After(now)
	}
	// endDateTimeがない場合は、startDateTimeの当日のみ開催中とみなす
	return e.startDateTime.Year() == now.Year() &&
		e.startDateTime.YearDay() == now.YearDay()
}

// Validate はイベントの状態が有効かを検証する
func (e *Event) Validate() error {
	if e.title.Value() == "" {
		return errors.New("タイトルは必須です")
	}
	if e.endDateTime != nil && e.endDateTime.Before(e.startDateTime) {
		return errors.New("終了日時は開始日時より後にしてください")
	}
	return nil
}
