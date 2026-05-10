package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
	appUserPrefs "github.com/kuro48/idol-api/internal/application/userprefs"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// MeHandler は認証済み OIDC ユーザーの自分自身の情報を扱うハンドラー
type MeHandler struct {
	prefsService *appUserPrefs.Service
}

// NewMeHandler は MeHandler を作成する
func NewMeHandler(prefsService *appUserPrefs.Service) *MeHandler {
	return &MeHandler{prefsService: prefsService}
}

type meResponse struct {
	Sub         string   `json:"sub"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	OshiColor   string   `json:"oshi_color"`
	Scopes      []string `json:"scopes"`
	CanWrite    bool     `json:"can_write"`
	CanAdmin    bool     `json:"can_admin"`
}

type updateOshiColorReq struct {
	OshiColor string `json:"oshi_color"`
}

// GetMe は現在の OIDC ユーザー情報と設定を返す
// @Summary     自分の情報取得
// @Tags        me
// @Produce     json
// @Success     200 {object} meResponse
// @Failure     401 {object} middleware.ErrorResponse
// @Router      /me [get]
func (h *MeHandler) GetMe(c *gin.Context) {
	principal, ok := domainAuth.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return
	}

	prefs, err := h.prefsService.GetOrCreate(c.Request.Context(), principal.SubjectID)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "ユーザー設定の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, meResponse{
		Sub:         principal.SubjectID,
		Email:       principal.Email,
		DisplayName: principal.DisplayName,
		OshiColor:   prefs.OshiColor(),
		Scopes:      principal.Scopes,
		CanWrite:    principal.CanWrite(),
		CanAdmin:    principal.CanAdmin(),
	})
}

// UpdateMyOshiColor は推しメンカラーを更新する
// @Summary     推しメンカラーの更新
// @Tags        me
// @Accept      json
// @Produce     json
// @Param       request body updateOshiColorReq true "推しメンカラー"
// @Success     200 {object} meResponse
// @Failure     400 {object} middleware.ErrorResponse
// @Router      /me/oshi-color [patch]
func (h *MeHandler) UpdateMyOshiColor(c *gin.Context) {
	principal, ok := domainAuth.PrincipalFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return
	}

	var req updateOshiColorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "リクエストが不正です"})
		return
	}

	prefs, err := h.prefsService.UpdateOshiColor(c.Request.Context(), principal.SubjectID, req.OshiColor)
	if err != nil {
		middleware.WriteError(c, err, middleware.ErrorContext{Message: "推しメンカラーの更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, meResponse{
		Sub:         principal.SubjectID,
		Email:       principal.Email,
		DisplayName: principal.DisplayName,
		OshiColor:   prefs.OshiColor(),
		Scopes:      principal.Scopes,
		CanWrite:    principal.CanWrite(),
		CanAdmin:    principal.CanAdmin(),
	})
}
