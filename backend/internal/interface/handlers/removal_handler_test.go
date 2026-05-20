package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	domainAuth "github.com/kuro48/idol-api/internal/domain/auth"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/usecase/removal"
	"github.com/stretchr/testify/assert"
)

type stubRemovalUseCase struct {
	overdue       []*removal.RemovalRequestDTO
	my            []*removal.PublicRemovalRequestDTO
	createCmd     removal.CreateRemovalRequestCommand
	listMySubject string
}

func (s *stubRemovalUseCase) CreateRemovalRequest(_ context.Context, cmd removal.CreateRemovalRequestCommand) (*removal.CreateRemovalRequestResult, error) {
	s.createCmd = cmd
	return &removal.CreateRemovalRequestResult{
		RemovalRequest: &removal.RemovalRequestDTO{ID: "507f1f77bcf86cd799439012"},
		AccessToken:    "removal-token",
	}, nil
}

func (s *stubRemovalUseCase) GetRemovalRequest(context.Context, string) (*removal.RemovalRequestDTO, error) {
	return nil, nil
}

func (s *stubRemovalUseCase) GetRemovalRequestPublic(context.Context, string, string) (*removal.PublicRemovalRequestDTO, error) {
	return nil, nil
}

func (s *stubRemovalUseCase) ListAllRemovalRequests(context.Context) ([]*removal.RemovalRequestDTO, error) {
	return nil, nil
}

func (s *stubRemovalUseCase) ListPendingRemovalRequests(context.Context) ([]*removal.RemovalRequestDTO, error) {
	return nil, nil
}

func (s *stubRemovalUseCase) ListOverdueRemovalRequests(context.Context) ([]*removal.RemovalRequestDTO, error) {
	return s.overdue, nil
}

func (s *stubRemovalUseCase) ListMyRemovalRequests(_ context.Context, subjectID string) ([]*removal.PublicRemovalRequestDTO, error) {
	s.listMySubject = subjectID
	return s.my, nil
}

func (s *stubRemovalUseCase) UpdateStatus(context.Context, removal.UpdateStatusCommand) (*removal.RemovalRequestDTO, error) {
	return nil, nil
}

func authenticatedRemovalRouter(uc *stubRemovalUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.NewRemovalHandler(uc)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		principal := &domainAuth.Principal{
			SubjectID: "identity-123",
			Email:     "user@example.com",
		}
		c.Request = c.Request.WithContext(domainAuth.WithPrincipal(c.Request.Context(), principal))
		c.Next()
	})
	router.POST("/removal-requests", h.CreateRemovalRequest)
	router.GET("/me/removal-requests", h.ListMyRemovalRequests)
	return router
}

func TestCreateRemovalRequest_UsesAuthenticatedPrincipal(t *testing.T) {
	uc := &stubRemovalUseCase{}
	router := authenticatedRemovalRouter(uc)
	body := `{"target_type":"idol","target_id":"idol-1","requester_type":"third_party","reason":"十分な削除理由をここに記載します","contact_info":"attacker@example.com","description":"十分な詳細説明をここに記載します"}`
	req := httptest.NewRequest(http.MethodPost, "/removal-requests", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "user@example.com", uc.createCmd.ContactInfo)
	assert.Equal(t, "identity-123", uc.createCmd.RequesterIdentityID)
}

func TestListMyRemovalRequests_UsesAuthenticatedSubject(t *testing.T) {
	uc := &stubRemovalUseCase{
		my: []*removal.PublicRemovalRequestDTO{{ID: "507f1f77bcf86cd799439012"}},
	}
	router := authenticatedRemovalRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/me/removal-requests", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "identity-123", uc.listMySubject)
	assert.Contains(t, w.Body.String(), `"count":1`)
}

func TestListOverdueRemovalRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &stubRemovalUseCase{
		overdue: []*removal.RemovalRequestDTO{
			{
				ID:         "507f1f77bcf86cd799439012",
				Status:     "pending",
				SLAOverdue: true,
				SLADueAt:   time.Now().Add(-time.Hour),
			},
		},
	}
	h := handlers.NewRemovalHandler(uc)

	router := gin.New()
	router.GET("/removal-requests/overdue", h.ListOverdueRemovalRequests)

	req := httptest.NewRequest(http.MethodGet, "/removal-requests/overdue", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"count":1`)
	assert.Contains(t, w.Body.String(), `"sla_overdue":true`)
}
