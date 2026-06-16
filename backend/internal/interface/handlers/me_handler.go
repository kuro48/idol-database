package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

type meOshiEntry struct {
	IdolID   string `json:"idol_id"`
	FanSince string `json:"fan_since,omitempty"`
}

type meResponse struct {
	Sub         string        `json:"sub"`
	Email       string        `json:"email"`
	DisplayName string        `json:"display_name"`
	OshiColor   string        `json:"oshi_color"`
	Oshis       []meOshiEntry `json:"oshis"`
	Scopes      []string      `json:"scopes"`
	CanWrite    bool          `json:"can_write"`
	CanAdmin    bool          `json:"can_admin"`
}

func (h *MeHandler) GetMe(c *gin.Context) {
	principal, ok := domainAuth.PrincipalFromContext(c.Request.Context())
	if !ok || principal.SubjectID == "" {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return
	}

	oshis := make([]meOshiEntry, len(principal.Oshis))
	for i, o := range principal.Oshis {
		oshis[i] = meOshiEntry{IdolID: o.IdolID, FanSince: o.FanSince}
	}

	c.JSON(http.StatusOK, meResponse{
		Sub:         principal.SubjectID,
		Email:       principal.Email,
		DisplayName: principal.DisplayName,
		OshiColor:   principal.OshiColor,
		Oshis:       oshis,
		Scopes:      principal.Scopes,
		CanWrite:    principal.CanWrite(),
		CanAdmin:    principal.CanAdmin(),
	})
}
