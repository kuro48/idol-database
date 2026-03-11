package webhook

import (
	"time"
)

// EventType はWebhookイベントの種別
type EventType string

const (
	EventIdolCreated   EventType = "idol.created"
	EventIdolUpdated   EventType = "idol.updated"
	EventIdolDeleted   EventType = "idol.deleted"
	EventGroupCreated  EventType = "group.created"
	EventGroupUpdated  EventType = "group.updated"
	EventGroupDeleted  EventType = "group.deleted"
	EventRemovalApproved EventType = "removal.approved"
)

// DeliveryStatus はWebhook配信状態
type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySuccess DeliveryStatus = "success"
	DeliveryFailed  DeliveryStatus = "failed"
)

// Subscription はWebhook購読設定
type Subscription struct {
	id        string
	url       string
	secret    string // HMAC-SHA256署名用シークレット
	events    []EventType
	active    bool
	createdAt time.Time
	createdBy string
}

// NewSubscription は新しいWebhook購読を作成する
func NewSubscription(id, url, secret string, events []EventType, createdBy string) *Subscription {
	return &Subscription{
		id:        id,
		url:       url,
		secret:    secret,
		events:    events,
		active:    true,
		createdAt: time.Now(),
		createdBy: createdBy,
	}
}

func (s *Subscription) ID() string          { return s.id }
func (s *Subscription) URL() string         { return s.url }
func (s *Subscription) Secret() string      { return s.secret }
func (s *Subscription) Events() []EventType { return s.events }
func (s *Subscription) Active() bool        { return s.active }
func (s *Subscription) CreatedAt() time.Time { return s.createdAt }
func (s *Subscription) CreatedBy() string   { return s.createdBy }

// Deactivate はWebhook購読を無効化する
func (s *Subscription) Deactivate() { s.active = false }

// MatchesEvent はイベント種別が購読対象かを判定する
func (s *Subscription) MatchesEvent(event EventType) bool {
	for _, e := range s.events {
		if e == event {
			return true
		}
	}
	return false
}

// Delivery はWebhook配信記録
type Delivery struct {
	id             string
	subscriptionID string
	event          EventType
	payload        []byte
	status         DeliveryStatus
	attempts       int
	maxAttempts    int
	lastAttemptAt  *time.Time
	nextRetryAt    *time.Time
	responseCode   *int
	errorMessage   string
	createdAt      time.Time
}

// NewDelivery は新しい配信記録を作成する
func NewDelivery(id, subscriptionID string, event EventType, payload []byte) *Delivery {
	return &Delivery{
		id:             id,
		subscriptionID: subscriptionID,
		event:          event,
		payload:        payload,
		status:         DeliveryPending,
		attempts:       0,
		maxAttempts:    5,
		createdAt:      time.Now(),
	}
}

func (d *Delivery) ID() string               { return d.id }
func (d *Delivery) SubscriptionID() string   { return d.subscriptionID }
func (d *Delivery) Event() EventType         { return d.event }
func (d *Delivery) Payload() []byte          { return d.payload }
func (d *Delivery) Status() DeliveryStatus   { return d.status }
func (d *Delivery) Attempts() int            { return d.attempts }
func (d *Delivery) MaxAttempts() int         { return d.maxAttempts }
func (d *Delivery) LastAttemptAt() *time.Time { return d.lastAttemptAt }
func (d *Delivery) NextRetryAt() *time.Time  { return d.nextRetryAt }
func (d *Delivery) ResponseCode() *int       { return d.responseCode }
func (d *Delivery) ErrorMessage() string     { return d.errorMessage }
func (d *Delivery) CreatedAt() time.Time     { return d.createdAt }

// CanRetry はリトライ可能かを判定する
func (d *Delivery) CanRetry() bool {
	return d.status == DeliveryFailed && d.attempts < d.maxAttempts
}

// MarkSuccess は配信成功を記録する
func (d *Delivery) MarkSuccess(responseCode int) {
	now := time.Now()
	d.status = DeliverySuccess
	d.attempts++
	d.lastAttemptAt = &now
	d.responseCode = &responseCode
}

// MarkFailed は配信失敗を記録し、次回リトライ時刻を指数バックオフで設定する
func (d *Delivery) MarkFailed(responseCode *int, errMsg string) {
	now := time.Now()
	d.attempts++
	d.lastAttemptAt = &now
	d.responseCode = responseCode
	d.errorMessage = errMsg

	if d.attempts >= d.maxAttempts {
		d.status = DeliveryFailed
		d.nextRetryAt = nil
	} else {
		d.status = DeliveryFailed
		// 指数バックオフ: 1分, 5分, 30分, 2時間, 6時間
		backoffMinutes := []time.Duration{1, 5, 30, 120, 360}
		idx := d.attempts - 1
		if idx >= len(backoffMinutes) {
			idx = len(backoffMinutes) - 1
		}
		next := now.Add(backoffMinutes[idx] * time.Minute)
		d.nextRetryAt = &next
	}
}
