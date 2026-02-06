package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kuro48/idol-api/internal/domain/event"
)

// ApplicationService はイベントアプリケーションサービス
type ApplicationService struct {
	repository event.Repository
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(repository event.Repository) *ApplicationService {
	return &ApplicationService{
		repository: repository,
	}
}

// CreateEvent はイベントを作成する
func (s *ApplicationService) CreateEvent(ctx context.Context, input CreateInput) (*event.Event, error) {
	// 値オブジェクトの生成
	title, err := event.NewEventTitle(input.Title)
	if err != nil {
		return nil, fmt.Errorf("タイトルの生成エラー: %w", err)
	}

	eventType, err := event.NewEventType(input.EventType)
	if err != nil {
		return nil, fmt.Errorf("イベントタイプの生成エラー: %w", err)
	}

	startDateTime, err := time.Parse(time.RFC3339, input.StartDateTime)
	if err != nil {
		return nil, fmt.Errorf("開始日時のパースエラー: %w", err)
	}

	// エンティティの生成
	newEvent, err := event.NewEvent(title, eventType, startDateTime)
	if err != nil {
		return nil, fmt.Errorf("イベントの生成エラー: %w", err)
	}

	// IDを生成
	id, err := event.NewEventID(generateID())
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}
	newEvent.SetID(id)

	// 終了日時の設定
	var endDateTime *time.Time
	if input.EndDateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *input.EndDateTime)
		if err == nil {
			endDateTime = &parsed
		}
	}

	// 詳細情報の設定
	newEvent.UpdateDetails(
		nil,
		nil,
		endDateTime,
		input.VenueID,
		input.TicketURL,
		input.OfficialURL,
		input.Description,
	)

	// パフォーマーの追加
	for _, performerID := range input.PerformerIDs {
		if err := newEvent.AddPerformer(performerID); err != nil {
			return nil, fmt.Errorf("パフォーマー追加エラー: %w", err)
		}
	}

	// タグの追加
	for _, tag := range input.Tags {
		if err := newEvent.AddTag(tag); err != nil {
			return nil, fmt.Errorf("タグ追加エラー: %w", err)
		}
	}

	// 保存
	if err := s.repository.Save(ctx, newEvent); err != nil {
		return nil, fmt.Errorf("イベントの保存エラー: %w", err)
	}

	return newEvent, nil
}

// GetEvent はイベントを取得する
func (s *ApplicationService) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	eventID, err := event.NewEventID(id)
	if err != nil {
		return nil, fmt.Errorf("IDの生成エラー: %w", err)
	}

	foundEvent, err := s.repository.FindByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("イベントの取得エラー: %w", err)
	}

	return foundEvent, nil
}

// SearchEvents は条件を指定してイベントを検索する（並行処理版）
func (s *ApplicationService) SearchEvents(ctx context.Context, criteria event.SearchCriteria) ([]*event.Event, int64, error) {
	var events []*event.Event
	var total int64
	var errSearch, errCount error

	var wg sync.WaitGroup
	wg.Add(2)

	// データ取得
	go func() {
		defer wg.Done()
		events, errSearch = s.repository.Search(ctx, criteria)
	}()

	// 総件数取得
	go func() {
		defer wg.Done()
		total, errCount = s.repository.Count(ctx, criteria)
	}()

	wg.Wait()

	if errSearch != nil {
		return nil, 0, fmt.Errorf("検索エラー: %w", errSearch)
	}
	if errCount != nil {
		return nil, 0, fmt.Errorf("件数取得エラー: %w", errCount)
	}

	return events, total, nil
}

// UpdateEvent はイベントを更新する
func (s *ApplicationService) UpdateEvent(ctx context.Context, input UpdateInput) error {
	id, err := event.NewEventID(input.ID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("イベントの取得エラー: %w", err)
	}

	// タイトルの更新
	var newTitle *event.EventTitle
	if input.Title != nil {
		title, err := event.NewEventTitle(*input.Title)
		if err != nil {
			return fmt.Errorf("タイトルの生成エラー: %w", err)
		}
		newTitle = &title
	}

	// 開始日時の更新
	var startDateTime *time.Time
	if input.StartDateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *input.StartDateTime)
		if err != nil {
			return fmt.Errorf("開始日時のパースエラー: %w", err)
		}
		startDateTime = &parsed
	}

	// 終了日時の更新
	var endDateTime *time.Time
	if input.EndDateTime != nil {
		parsed, err := time.Parse(time.RFC3339, *input.EndDateTime)
		if err != nil {
			return fmt.Errorf("終了日時のパースエラー: %w", err)
		}
		endDateTime = &parsed
	}

	// 更新
	existingEvent.UpdateDetails(
		newTitle,
		startDateTime,
		endDateTime,
		input.VenueID,
		input.TicketURL,
		input.OfficialURL,
		input.Description,
	)

	// 保存
	if err := s.repository.Update(ctx, existingEvent); err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}

	return nil
}

// DeleteEvent はイベントを削除する
func (s *ApplicationService) DeleteEvent(ctx context.Context, id string) error {
	eventID, err := event.NewEventID(id)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	if err := s.repository.Delete(ctx, eventID); err != nil {
		return fmt.Errorf("イベントの削除エラー: %w", err)
	}

	return nil
}

// AddPerformer はパフォーマーを追加する
func (s *ApplicationService) AddPerformer(ctx context.Context, input AddPerformerInput) error {
	id, err := event.NewEventID(input.EventID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("イベントの取得エラー: %w", err)
	}

	if err := existingEvent.AddPerformer(input.PerformerID); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, existingEvent); err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}

	return nil
}

// RemovePerformer はパフォーマーを削除する
func (s *ApplicationService) RemovePerformer(ctx context.Context, input RemovePerformerInput) error {
	id, err := event.NewEventID(input.EventID)
	if err != nil {
		return fmt.Errorf("IDの生成エラー: %w", err)
	}

	existingEvent, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("イベントの取得エラー: %w", err)
	}

	existingEvent.RemovePerformer(input.PerformerID)

	if err := s.repository.Update(ctx, existingEvent); err != nil {
		return fmt.Errorf("イベントの更新エラー: %w", err)
	}

	return nil
}

// FindUpcoming は今後開催されるイベントを取得する
func (s *ApplicationService) FindUpcoming(ctx context.Context, limit int) ([]*event.Event, error) {
	events, err := s.repository.FindUpcoming(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("今後のイベント取得エラー: %w", err)
	}

	return events, nil
}

// generateID はIDを生成する（簡易実装）
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
