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
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/kuro48/idol-api/internal/domain/webhook"
)

// deliveryTimeout はWebhook配信の最大待ち時間
const deliveryTimeout = 30 * time.Second

const webhookSignatureTolerance = 5 * time.Minute

// ApplicationService はWebhookアプリケーションサービス
type ApplicationService struct {
	subRepo        webhook.SubscriptionRepository
	deliveryRepo   webhook.DeliveryRepository
	httpClient     *http.Client
	resolveIPAddrs func(ctx context.Context, host string) ([]net.IPAddr, error)
	replayMu       sync.Mutex
	replayNonces   map[string]time.Time
	now            func() time.Time
	wg             sync.WaitGroup
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(subRepo webhook.SubscriptionRepository, deliveryRepo webhook.DeliveryRepository) *ApplicationService {
	return NewApplicationServiceWithTimeout(subRepo, deliveryRepo, 10*time.Second)
}

// NewApplicationServiceWithTimeout はタイムアウトを指定してアプリケーションサービスを作成する
func NewApplicationServiceWithTimeout(subRepo webhook.SubscriptionRepository, deliveryRepo webhook.DeliveryRepository, timeout time.Duration) *ApplicationService {
	svc := &ApplicationService{
		subRepo:        subRepo,
		deliveryRepo:   deliveryRepo,
		resolveIPAddrs: net.DefaultResolver.LookupIPAddr,
		replayNonces:   make(map[string]time.Time),
		now:            time.Now,
	}
	svc.httpClient = newWebhookHTTPClient(timeout, svc.resolveIPAddrs)
	return svc
}

// Shutdown はインフライトのWebhook配信がすべて完了するまで待機する
func (s *ApplicationService) Shutdown() {
	s.wg.Wait()
}

// StartRetryWorker は失敗したWebhook配信を定期的にリトライするバックグラウンドワーカーを起動する。
// ctx がキャンセルされるとワーカーは停止し、Shutdown() の待機対象に含まれる。
func (s *ApplicationService) StartRetryWorker(ctx context.Context, interval time.Duration) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RetryPendingDeliveries(ctx); err != nil {
					slog.Error("Webhookリトライワーカーエラー", "error", err)
				}
			}
		}
	}()
}

// CreateSubscriptionInput はWebhook購読作成入力
type CreateSubscriptionInput struct {
	URL       string
	Events    []webhook.EventType
	CreatedBy string
}

// CreateSubscription はWebhook購読を作成する
func (s *ApplicationService) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*webhook.Subscription, error) {
	if err := validateWebhookURL(context.Background(), input.URL, s.resolveIPAddrs); err != nil {
		return nil, err
	}

	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("サブスクリプションIDの生成エラー: %w", err)
	}
	secret, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("シークレットの生成エラー: %w", err)
	}

	sub := webhook.NewSubscription(id, input.URL, secret, input.Events, input.CreatedBy)
	if err := s.subRepo.Save(ctx, sub); err != nil {
		return nil, fmt.Errorf("Webhook購読の保存エラー: %w", err)
	}
	return sub, nil
}

func newWebhookHTTPClient(timeout time.Duration, resolveIPAddrs func(ctx context.Context, host string) ([]net.IPAddr, error)) *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, fmt.Errorf("Webhook送信先アドレスが不正です: %w", err)
			}

			targets, err := resolveAllowedWebhookDialTargets(ctx, host, port, resolveIPAddrs)
			if err != nil {
				return nil, err
			}

			var lastErr error
			dialer := &net.Dialer{Timeout: timeout}
			for _, target := range targets {
				conn, dialErr := dialer.DialContext(ctx, network, target)
				if dialErr == nil {
					return conn, nil
				}
				lastErr = dialErr
			}

			if lastErr == nil {
				lastErr = fmt.Errorf("Webhook送信先に到達できません")
			}
			return nil, lastErr
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("リダイレクト回数の上限（3回）に達しました")
			}
			return validateWebhookURL(req.Context(), req.URL.String(), resolveIPAddrs)
		},
	}
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
		deliveryID, err := generateID()
		if err != nil {
			slog.Error("配信IDの生成に失敗しました", "subscription_id", sub.ID(), "error", err)
			continue
		}
		delivery := webhook.NewDelivery(deliveryID, sub.ID(), event, payloadBytes)
		if err := s.deliveryRepo.Save(ctx, delivery); err != nil {
			continue
		}

		s.wg.Add(1)
		localSub := sub
		localDelivery := delivery
		baseCtx := context.WithoutCancel(ctx)
		go func() {
			defer s.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Webhook配信パニック回復", "subscription_id", localSub.ID(), "panic", r)
				}
			}()
			deliverCtx, cancel := context.WithTimeout(baseCtx, deliveryTimeout)
			defer cancel()
			s.deliver(deliverCtx, localSub, localDelivery)
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
		s.wg.Add(1)
		localSub := sub
		localDelivery := delivery
		baseCtx := ctx
		go func() {
			defer s.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Webhook配信パニック回復", "subscription_id", localSub.ID(), "panic", r)
				}
			}()
			deliverCtx, cancel := context.WithTimeout(baseCtx, deliveryTimeout)
			defer cancel()
			s.deliver(deliverCtx, localSub, localDelivery)
		}()
	}

	return nil
}

// deliver は実際のHTTPリクエストを送信する
func (s *ApplicationService) deliver(ctx context.Context, sub *webhook.Subscription, delivery *webhook.Delivery) {
	timestamp := strconv.FormatInt(s.now().Unix(), 10)
	signature := computeSignature(sub.Secret(), timestamp, delivery.Payload())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sub.URL(), bytes.NewReader(delivery.Payload()))
	if err != nil {
		delivery.MarkFailed(nil, err.Error())
		if updateErr := s.deliveryRepo.Update(ctx, delivery); updateErr != nil {
			slog.Error("配信失敗状態の更新に失敗しました", "delivery_id", delivery.ID(), "error", updateErr)
		}
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "sha256="+signature)
	req.Header.Set("X-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Webhook-Nonce", delivery.ID())
	req.Header.Set("X-Webhook-Event", string(delivery.Event()))
	req.Header.Set("X-Delivery-ID", delivery.ID())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		delivery.MarkFailed(nil, err.Error())
		if updateErr := s.deliveryRepo.Update(ctx, delivery); updateErr != nil {
			slog.Error("配信失敗状態の更新に失敗しました", "delivery_id", delivery.ID(), "error", updateErr)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.MarkSuccess(resp.StatusCode)
	} else {
		delivery.MarkFailed(&resp.StatusCode, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	if updateErr := s.deliveryRepo.Update(ctx, delivery); updateErr != nil {
		slog.Error("配信状態の更新に失敗しました", "delivery_id", delivery.ID(), "error", updateErr)
	}
}

// validateWebhookURL はWebhookURLのSSRF対策バリデーションを行う
func validateWebhookURL(ctx context.Context, rawURL string, resolveIPAddrs func(ctx context.Context, host string) ([]net.IPAddr, error)) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("無効なURL形式です: %w", err)
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("WebhookURLはhttpsスキームのみ使用できます")
	}

	// 認証情報の埋め込みを禁止（SSRF拡大リスク）
	if parsed.User != nil {
		return fmt.Errorf("WebhookURLに認証情報を含めることはできません")
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("URLにホスト名が必要です")
	}

	_, err = resolveAllowedWebhookDialTargets(ctx, host, parsed.Port(), resolveIPAddrs)
	return err
}

func resolveAllowedWebhookDialTargets(ctx context.Context, host, port string, resolveIPAddrs func(ctx context.Context, host string) ([]net.IPAddr, error)) ([]string, error) {
	if host == "localhost" {
		return nil, fmt.Errorf("ループバックアドレスへのWebhookは許可されていません")
	}

	if port == "" {
		port = "443"
	}

	if ip := net.ParseIP(host); ip != nil {
		if err := validateWebhookIP(ip); err != nil {
			return nil, err
		}
		return []string{net.JoinHostPort(ip.String(), port)}, nil
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	addrs, err := resolveIPAddrs(resolveCtx, host)
	if err != nil {
		return nil, fmt.Errorf("Webhook送信先のDNS解決に失敗しました: %w", err)
	}

	targets := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if err := validateWebhookIP(addr.IP); err != nil {
			return nil, err
		}
		targets = append(targets, net.JoinHostPort(addr.IP.String(), port))
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("Webhook送信先のIPアドレスが解決できませんでした")
	}
	return targets, nil
}

func validateWebhookIP(ip net.IP) error {
	if ip == nil {
		return fmt.Errorf("Webhook送信先のIPアドレスが不正です")
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || isPrivateIP(ip) {
		return fmt.Errorf("プライベート/ループバックIPアドレスへのWebhookは許可されていません")
	}
	return nil
}

// isPrivateIP はIPアドレスがプライベート範囲かチェックする
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"100.64.0.0/10",
		"198.18.0.0/15",
		"fc00::/7",
		"fe80::/10",
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
func computeSignature(secret string, timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature はWebhookリクエストの署名を検証する
func VerifySignature(secret, signature string, timestamp string, payload []byte) bool {
	expected := "sha256=" + computeSignature(secret, timestamp, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifyWebhookRequestInput は受信Webhook検証に必要な入力。
type VerifyWebhookRequestInput struct {
	SubscriptionID string
	Signature      string
	Timestamp      string
	Nonce          string
	Payload        []byte
}

// VerifyWebhookRequest はサブスクリプションIDと署名を検証する
func (s *ApplicationService) VerifyWebhookRequest(ctx context.Context, input VerifyWebhookRequestInput) error {
	sub, err := s.subRepo.FindByID(ctx, input.SubscriptionID)
	if err != nil {
		return fmt.Errorf("サブスクリプションが見つかりません: %w", err)
	}
	if !sub.Active() {
		return fmt.Errorf("サブスクリプションが無効です")
	}
	if err := s.validateTimestamp(input.Timestamp); err != nil {
		return err
	}
	if !VerifySignature(sub.Secret(), input.Signature, input.Timestamp, input.Payload) {
		return fmt.Errorf("署名が無効です")
	}
	if err := s.claimReplayNonce(input.SubscriptionID, input.Nonce); err != nil {
		return err
	}
	return nil
}

func (s *ApplicationService) validateTimestamp(raw string) error {
	ts, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fmt.Errorf("Webhook timestamp が不正です")
	}
	signedAt := time.Unix(ts, 0)
	now := s.now()
	if signedAt.Before(now.Add(-webhookSignatureTolerance)) || signedAt.After(now.Add(webhookSignatureTolerance)) {
		return fmt.Errorf("Webhook timestamp が許容範囲外です")
	}
	return nil
}

func (s *ApplicationService) claimReplayNonce(subscriptionID, nonce string) error {
	if nonce == "" {
		return fmt.Errorf("Webhook nonce が未設定です")
	}

	s.replayMu.Lock()
	defer s.replayMu.Unlock()

	now := s.now()
	cutoff := now.Add(-webhookSignatureTolerance)
	for key, seenAt := range s.replayNonces {
		if seenAt.Before(cutoff) {
			delete(s.replayNonces, key)
		}
	}

	key := subscriptionID + ":" + nonce
	if _, exists := s.replayNonces[key]; exists {
		return fmt.Errorf("Webhook nonce が再利用されています")
	}
	s.replayNonces[key] = now
	return nil
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("乱数生成エラー: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("乱数生成エラー: %w", err)
	}
	return hex.EncodeToString(b), nil
}
