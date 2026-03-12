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
	"net"
	"log"
	"net/http"
	"net/url"
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
	return NewApplicationServiceWithTimeout(subRepo, deliveryRepo, 10*time.Second)
}

// NewApplicationServiceWithTimeout はタイムアウトを指定してアプリケーションサービスを作成する
func NewApplicationServiceWithTimeout(subRepo webhook.SubscriptionRepository, deliveryRepo webhook.DeliveryRepository, timeout time.Duration) *ApplicationService {
	return &ApplicationService{
		subRepo:      subRepo,
		deliveryRepo: deliveryRepo,
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("リダイレクト回数の上限（3回）に達しました")
				}
				// リダイレクト先もSSRFチェック
				return validateWebhookURL(req.URL.String())
			},
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
	// URLバリデーション（SSRF対策）
	if err := validateWebhookURL(input.URL); err != nil {
		return nil, err
	}

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

// validateWebhookURL はWebhookURLのSSRF対策バリデーションを行う
func validateWebhookURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("無効なURL形式です: %w", err)
	}

	// httpsのみ許可
	if parsed.Scheme != "https" {
		return fmt.Errorf("WebhookURLはhttpsスキームのみ使用できます")
	}

	// ホスト名の解決とプライベートIPチェック
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("URLにホスト名が必要です")
	}

	// ループバックアドレスの拒否
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("ループバックアドレスへのWebhookは許可されていません")
	}

	// DNS解決してIPチェック
	addrs, err := net.LookupHost(host)
	if err == nil {
		for _, addr := range addrs {
			ip := net.ParseIP(addr)
			if ip == nil {
				continue
			}
			if isPrivateIP(ip) {
				return fmt.Errorf("プライベートIPアドレスへのWebhookは許可されていません")
			}
		}
	}

	return nil
}

// isPrivateIP はIPアドレスがプライベート範囲かチェックする
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16", // リンクローカル
		"fc00::/7",       // IPv6 ユニークローカル
		"fe80::/10",      // IPv6 リンクローカル
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
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

// VerifyWebhookRequest はサブスクリプションIDと署名を検証する
func (s *ApplicationService) VerifyWebhookRequest(ctx context.Context, subscriptionID, signature string, payload []byte) error {
	sub, err := s.subRepo.FindByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("サブスクリプションが見つかりません: %w", err)
	}
	if !sub.Active() {
		return fmt.Errorf("サブスクリプションが無効です")
	}
	if !VerifySignature(sub.Secret(), signature, payload) {
		return fmt.Errorf("署名が無効です")
	}
	return nil
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
