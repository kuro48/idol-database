package stripe

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	appBilling "github.com/kuro48/idol-api/internal/application/billing"
	"github.com/kuro48/idol-api/internal/domain/plan"
)

const (
	defaultAPIBaseURL       = "https://api.stripe.com"
	webhookToleranceSeconds = 300
)

// Client は Stripe REST API の最小クライアント。
type Client struct {
	secretKey     string
	webhookSecret string
	apiBaseURL    string
	httpClient    *http.Client
}

// NewClient は Stripe クライアントを作成する。
func NewClient(secretKey, webhookSecret string) *Client {
	return &Client{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		apiBaseURL:    defaultAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CreateCheckoutSession は Stripe Checkout Session を作成する。
func (c *Client) CreateCheckoutSession(ctx context.Context, input appBilling.CreateCheckoutSessionInput) (*appBilling.CheckoutSession, error) {
	values := url.Values{}
	values.Set("mode", "subscription")
	values.Set("customer_email", input.Email)
	values.Set("success_url", input.SuccessURL)
	values.Set("cancel_url", input.CancelURL)
	values.Set("line_items[0][price]", input.PriceID)
	values.Set("line_items[0][quantity]", "1")
	values.Set("metadata[name]", input.Name)
	values.Set("metadata[plan_type]", input.PlanType)

	body, err := c.doFormRequest(ctx, http.MethodPost, "/v1/checkout/sessions", values)
	if err != nil {
		return nil, err
	}

	var resp struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Stripe Checkout Session レスポンスの解析に失敗しました: %w", err)
	}
	if resp.ID == "" || resp.URL == "" {
		return nil, fmt.Errorf("Stripe Checkout Session レスポンスが不正です")
	}

	return &appBilling.CheckoutSession{ID: resp.ID, URL: resp.URL}, nil
}

// VerifyWebhookEvent は Stripe Webhook を検証して最小イベントへ変換する。
func (c *Client) VerifyWebhookEvent(payload []byte, signature string) (*appBilling.WebhookEvent, error) {
	if err := c.verifySignature(payload, signature); err != nil {
		return nil, err
	}

	var event struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID              string            `json:"id"`
				Customer        string            `json:"customer"`
				CustomerEmail   string            `json:"customer_email"`
				Metadata        map[string]string `json:"metadata"`
				CustomerDetails *struct {
					Email string `json:"email"`
					Name  string `json:"name"`
				} `json:"customer_details"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("Stripe Webhook のJSON解析に失敗しました: %w", err)
	}

	result := &appBilling.WebhookEvent{Type: event.Type}
	if event.Type != appBilling.WebhookEventTypeCheckoutSessionCompleted {
		return result, nil
	}

	metadata := event.Data.Object.Metadata
	email := event.Data.Object.CustomerEmail
	name := metadata["name"]
	if event.Data.Object.CustomerDetails != nil {
		if email == "" {
			email = event.Data.Object.CustomerDetails.Email
		}
		if name == "" {
			name = event.Data.Object.CustomerDetails.Name
		}
	}
	planType := plan.Type(metadata["plan_type"])
	if !plan.IsValid(planType) || planType == plan.TypeFree {
		return nil, fmt.Errorf("無効なプラン種別です")
	}
	if event.Data.Object.ID == "" || event.Data.Object.Customer == "" || email == "" {
		return nil, fmt.Errorf("Stripe Checkout Session の必須情報が不足しています")
	}

	result.CheckoutSession = &appBilling.CheckoutSessionCompleted{
		SessionID:  event.Data.Object.ID,
		CustomerID: event.Data.Object.Customer,
		Email:      email,
		Name:       name,
		PlanType:   planType,
	}
	return result, nil
}

// CreatePortalSession は Stripe Billing Portal Session を作成する。
func (c *Client) CreatePortalSession(ctx context.Context, input appBilling.CreatePortalSessionInput) (*appBilling.PortalSession, error) {
	values := url.Values{}
	values.Set("customer", input.CustomerID)
	values.Set("return_url", input.ReturnURL)

	body, err := c.doFormRequest(ctx, http.MethodPost, "/v1/billing_portal/sessions", values)
	if err != nil {
		return nil, err
	}

	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Stripe Portal Session レスポンスの解析に失敗しました: %w", err)
	}
	if resp.URL == "" {
		return nil, fmt.Errorf("Stripe Portal Session レスポンスが不正です")
	}
	return &appBilling.PortalSession{URL: resp.URL}, nil
}

func (c *Client) verifySignature(payload []byte, signature string) error {
	if c.webhookSecret == "" {
		return fmt.Errorf("Stripe Webhook secret が未設定です")
	}

	timestamp, signatures, err := parseStripeSignature(signature)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if timestamp < now-webhookToleranceSeconds || timestamp > now+webhookToleranceSeconds {
		return fmt.Errorf("Stripe Webhook の署名タイムスタンプが許容範囲外です")
	}

	signedPayload := strconv.FormatInt(timestamp, 10) + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	if _, err := mac.Write([]byte(signedPayload)); err != nil {
		return fmt.Errorf("Stripe Webhook 署名の計算に失敗しました: %w", err)
	}
	expected := hex.EncodeToString(mac.Sum(nil))
	for _, candidate := range signatures {
		if hmac.Equal([]byte(candidate), []byte(expected)) {
			return nil
		}
	}
	return fmt.Errorf("Stripe Webhook の署名が無効です")
}

func parseStripeSignature(header string) (int64, []string, error) {
	parts := strings.Split(header, ",")
	var (
		timestamp  int64
		foundTS    bool
		signatures []string
	)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			ts, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, nil, fmt.Errorf("Stripe-Signature の timestamp が不正です")
			}
			timestamp = ts
			foundTS = true
		case "v1":
			signatures = append(signatures, value)
		}
	}
	if !foundTS || len(signatures) == 0 {
		return 0, nil, fmt.Errorf("Stripe-Signature ヘッダーが不正です")
	}
	return timestamp, signatures, nil
}

func (c *Client) doFormRequest(ctx context.Context, method, path string, values url.Values) ([]byte, error) {
	if c.secretKey == "" {
		return nil, fmt.Errorf("Stripe secret key が未設定です")
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiBaseURL+path, bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Stripe リクエスト作成に失敗しました: %w", err)
	}
	req.SetBasicAuth(c.secretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Stripe API リクエストに失敗しました: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Stripe API レスポンスの読み取りに失敗しました: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Stripe API エラー: %s", strings.TrimSpace(string(body)))
	}
	return body, nil
}
