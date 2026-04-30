package removal

import (
	"context"
	"time"

	groupDomain "github.com/kuro48/idol-api/internal/domain/group"
	idolDomain "github.com/kuro48/idol-api/internal/domain/idol"
	domain "github.com/kuro48/idol-api/internal/domain/removal"
	domainWebhook "github.com/kuro48/idol-api/internal/domain/webhook"
)

// RemovalAppPort は removal.Usecase が removal application サービスに要求する契約
type RemovalAppPort interface {
	CreateRemovalRequest(ctx context.Context, input RemovalCreateInput) (*RemovalCreateResult, error)
	GetRemovalRequest(ctx context.Context, id string) (*domain.RemovalRequest, error)
	ListAllRemovalRequests(ctx context.Context) ([]*domain.RemovalRequest, error)
	ListPendingRemovalRequests(ctx context.Context) ([]*domain.RemovalRequest, error)
	UpdateRemovalRequest(ctx context.Context, request *domain.RemovalRequest) error
}

// RemovalIdolPort は removal.Usecase が idol application サービスに要求する契約
type RemovalIdolPort interface {
	GetIdol(ctx context.Context, id string) (*idolDomain.Idol, error)
	DeleteIdol(ctx context.Context, id string) error
}

// RemovalGroupPort は removal.Usecase が group application サービスに要求する契約
type RemovalGroupPort interface {
	GetGroup(ctx context.Context, id string) (*groupDomain.Group, error)
	DeleteGroup(ctx context.Context, id string) error
}

// RemovalWebhookPublisher は削除申請イベントを通知する契約
type RemovalWebhookPublisher interface {
	Publish(ctx context.Context, event domainWebhook.EventType, payload interface{}) error
}

// RemovalNotifier は削除申請の通知を送る契約。
type RemovalNotifier interface {
	NotifyReceived(ctx context.Context, notification ReceivedNotification) error
	NotifyResolved(ctx context.Context, notification ResolvedNotification) error
}

// ReceivedNotification は削除申請受付通知の内容。
type ReceivedNotification struct {
	To          string
	RequestID   string
	TargetType  string
	AccessToken string
	CreatedAt   time.Time
}

// ResolvedNotification は削除申請完了通知の内容。
type ResolvedNotification struct {
	To         string
	RequestID  string
	TargetType string
	Status     string
	UpdatedAt  time.Time
}

// RemovalCreateInput は削除申請作成の入力
type RemovalCreateInput struct {
	TargetType  string
	TargetID    string
	Requester   string
	Reason      string
	ContactInfo string
	Evidence    string
	Description string
}

// RemovalCreateResult はアプリケーション層の作成結果
type RemovalCreateResult struct {
	Request     *domain.RemovalRequest
	AccessToken string
}
