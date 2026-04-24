package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	domainapikey "github.com/kuro48/idol-api/internal/domain/apikey"
	"github.com/kuro48/idol-api/internal/domain/plan"
	domainusage "github.com/kuro48/idol-api/internal/domain/usage"
)

const (
	// CtxKeyPlanType はGinコンテキストに格納するプラン種別のキー
	CtxKeyPlanType = "plan_type"
	// CtxKeyWriteEnabled はGinコンテキストに格納するwrite権限フラグのキー
	CtxKeyWriteEnabled = "write_enabled"
	// CtxKeyAPIKeyPrefix はGinコンテキストに格納するAPIキープレフィックスのキー
	CtxKeyAPIKeyPrefix = "api_key_prefix"

	// planAuthTimeout はDB問い合わせのタイムアウト
	planAuthTimeout = 3 * time.Second
)

// PlanAuthMiddleware はプランベースのAPIキー認証と月次使用量制限を行うミドルウェア
type PlanAuthMiddleware struct {
	apikeyRepo domainapikey.Repository
	usageRepo  domainusage.Repository
}

// NewPlanAuth は PlanAuthMiddleware を作成する
func NewPlanAuth(apikeyRepo domainapikey.Repository, usageRepo domainusage.Repository) *PlanAuthMiddleware {
	return &PlanAuthMiddleware{
		apikeyRepo: apikeyRepo,
		usageRepo:  usageRepo,
	}
}

// OptionalAuth はAPIキーが提示された場合のみ認証・使用量カウントを行うミドルウェア関数を返す
// Authorization: Bearer ヘッダーがない場合は匿名アクセスとして通過させる（公開 read ルート用）
func (m *PlanAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := extractBearerToken(c)
		if rawKey == "" {
			// キーなし → 匿名アクセス（既存の IP ベースレート制限のみ適用）
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), planAuthTimeout)
		defer cancel()

		prefix := domainapikey.PrefixOf(rawKey)
		candidates, err := m.apikeyRepo.FindByPrefix(ctx, prefix)
		if err != nil {
			slog.Error("APIキー検索エラー", "error", err)
			c.JSON(http.StatusInternalServerError, NewInternalError("認証処理に失敗しました"))
			c.Abort()
			return
		}

		apiKey := findMatchingKey(candidates, rawKey)
		if apiKey == nil {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		limits := plan.GetLimits(apiKey.PlanType())
		yearMonth := domainusage.YearMonthOf(time.Now())

		usage, err := m.usageRepo.IncrementAndGet(ctx, apiKey.Prefix(), yearMonth, limits.MonthlyRequests)
		if err != nil {
			slog.Error("使用量カウントエラー", "error", err)
			c.JSON(http.StatusInternalServerError, NewInternalError("使用量の記録に失敗しました"))
			c.Abort()
			return
		}

		if usage.ExceedsLimit() {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Code:    "PLAN_LIMIT_EXCEEDED",
				Message: "月間リクエスト上限に達しました。プランのアップグレードをご検討ください。",
			})
			c.Abort()
			return
		}

		c.Set(CtxKeyPlanType, string(apiKey.PlanType()))
		c.Set(CtxKeyWriteEnabled, limits.WriteEnabled)
		c.Set(CtxKeyAPIKeyPrefix, apiKey.Prefix())
		c.Next()
	}
}

// Auth はAPIキー認証 + プラン制限を行うミドルウェア関数を返す
// Authorization: Bearer <api_key> ヘッダーからキーを取得する（キーなしは 401）
func (m *PlanAuthMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := extractBearerToken(c)
		if rawKey == "" {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), planAuthTimeout)
		defer cancel()

		prefix := domainapikey.PrefixOf(rawKey)
		candidates, err := m.apikeyRepo.FindByPrefix(ctx, prefix)
		if err != nil {
			slog.Error("APIキー検索エラー", "error", err)
			c.JSON(http.StatusInternalServerError, NewInternalError("認証処理に失敗しました"))
			c.Abort()
			return
		}

		apiKey := findMatchingKey(candidates, rawKey)
		if apiKey == nil {
			c.JSON(http.StatusUnauthorized, NewUnauthorizedError())
			c.Abort()
			return
		}

		limits := plan.GetLimits(apiKey.PlanType())
		yearMonth := domainusage.YearMonthOf(time.Now())

		usage, err := m.usageRepo.IncrementAndGet(ctx, apiKey.Prefix(), yearMonth, limits.MonthlyRequests)
		if err != nil {
			slog.Error("使用量カウントエラー", "error", err)
			c.JSON(http.StatusInternalServerError, NewInternalError("使用量の記録に失敗しました"))
			c.Abort()
			return
		}

		if usage.ExceedsLimit() {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Code:    "PLAN_LIMIT_EXCEEDED",
				Message: "月間リクエスト上限に達しました。プランのアップグレードをご検討ください。",
			})
			c.Abort()
			return
		}

		c.Set(CtxKeyPlanType, string(apiKey.PlanType()))
		c.Set(CtxKeyWriteEnabled, limits.WriteEnabled)
		c.Set(CtxKeyAPIKeyPrefix, apiKey.Prefix())
		c.Next()
	}
}

// RequireWrite は write スコープが必要なエンドポイント用ミドルウェアを返す
// PlanAuth.Auth() の後に使用する
func RequireWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		writeEnabled, exists := c.Get(CtxKeyWriteEnabled)
		if !exists || writeEnabled == false {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Code:    "WRITE_SCOPE_REQUIRED",
				Message: "このエンドポイントには write スコープが必要です。Developer プラン以上にアップグレードしてください。",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractBearerToken は Authorization: Bearer <token> からトークンを取り出す
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return ""
	}
	return strings.TrimPrefix(authHeader, prefix)
}

// findMatchingKey は候補キーの中から rawKey に一致するものを返す
func findMatchingKey(candidates []*domainapikey.APIKey, rawKey string) *domainapikey.APIKey {
	for _, k := range candidates {
		if k.VerifyKey(rawKey) {
			return k
		}
	}
	return nil
}
