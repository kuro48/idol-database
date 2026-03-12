// Package webhook はWebhook管理と配信を担当するアプリケーションサービス
package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kuro48/idol-api/internal/domain/webhook"
)

// ApplicationService はWebhookアプリケーションサービス
type ApplicationService struct {
	subRepo      webhook.SubscriptionRepository
	deliveryRepo webhook.DeliveryRepository
	httpClient   *http.Client
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(subRepo webhook.SubscriptionRepository, deliveryRepo webhook.DeliveryRepository) *ApplicationService {
	return &ApplicationService{
		subRepo:      subRepo,
		deliveryRepo: deliveryRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateSubscriptionInput はWebhook購読作成入力
type CreateSubscriptionInput struct {
	URL       string
	Events    []webhook.EventType
	CreatedBy string
}

// CreateSubscription はWebhook購読を作成する
func (s *ApplicationService) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*webhook.Subscription, error) {
	id := generateID()
	secret := generateSecret()
	sub := webhook.NewSubscription(id, input.URL, secret, input.Events, input.CreatedBy)
	if err := s.subRepo.Save(ctx, sub); err != nil {
		return nil, fmt.Errorf("Webhook購読の保存エラー: %w", err)
	}
	return sub, nil
}

// DeleteSubscription はWebhook購読を削除する
func (s *ApplicationService) DeleteSubscription(ctx context.Context, id string) error {
	return s.subRepo.Delete(ctx, id)
}

// ListSubscriptions はWebhook購読一覧を返す
func (s *ApplicationService) ListSubscriptions(ctx context.Context) ([]*webhook.Subscription, error) {
	return s.subRepo.FindAll(ctx)
}

// Publish はイベントをすべてのアクティブな購読者に配信する（非同期）
func (s *ApplicationService) Publish(ctx context.Context, event webhook.EventType, payload interface{}) error {
	subs, err := s.subRepo.FindActiveByEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("購読者取得エラー: %w", err)
	}

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      payload,
	})
	if err != nil {
		return fmt.Errorf("ペイロードのシリアライズエラー: %w", err)
	}

	for _, sub := range subs {
		delivery := webhook.NewDelivery(generateID(), sub.ID(), event, payloadBytes)
		if err := s.deliveryRepo.Save(ctx, delivery); err != nil {
			continue // 保存エラーでも他の購読者への配信は続ける
		}

		// 非同期で配信（コンテキストから独立させる）
		localSub := sub
		localDelivery := delivery
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Webhook配信パニック回復 [subscription: %s]: %v", localSub.ID(), r)
				}
			}()
			s.deliver(context.Background(), localSub, localDelivery)
		}()
	}

	return nil
}

// RetryPendingDeliveries はリトライ待ちの配信を再実行する
func (s *ApplicationService) RetryPendingDeliveries(ctx context.Context) error {
	deliveries, err := s.deliveryRepo.FindPendingRetries(ctx)
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		sub, err := s.subRepo.FindByID(ctx, delivery.SubscriptionID())
		if err != nil || !sub.Active() {
			continue
		}
		localSub := sub
		localDelivery := delivery
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Webhook配信パニック回復 [subscription: %s]: %v", localSub.ID(), r)
				}
			}()
			s.deliver(context.Background(), localSub, localDelivery)
		}()
	}

	return nil
}

// deliver は実際のHTTPリクエストを送信する
func (s *ApplicationService) deliver(ctx context.Context, sub *webhook.Subscription, delivery *webhook.Delivery) {
	signature := computeSignature(sub.Secret(), delivery.Payload())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL(), bytes.NewReader(delivery.Payload()))
	if err != nil {
		delivery.MarkFailed(nil, err.Error())
		_ = s.deliveryRepo.Update(ctx, delivery)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "sha256="+signature)
	req.Header.Set("X-Webhook-Event", string(delivery.Event()))
	req.Header.Set("X-Delivery-ID", delivery.ID())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		delivery.MarkFailed(nil, err.Error())
		_ = s.deliveryRepo.Update(ctx, delivery)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.MarkSuccess(resp.StatusCode)
	} else {
		delivery.MarkFailed(&resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	_ = s.deliveryRepo.Update(ctx, delivery)
}

// computeSignature はHMAC-SHA256シグネチャを計算する
func computeSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature はWebhookリクエストの署名を検証する
func VerifySignature(secret, signature string, payload []byte) bool {
	expected := "sha256=" + computeSignature(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
