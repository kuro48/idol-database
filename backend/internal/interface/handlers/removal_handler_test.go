package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/usecase/removal"
	"github.com/stretchr/testify/assert"
)

type stubRemovalUseCase struct {
	overdue []*removal.RemovalRequestDTO
}

func (s *stubRemovalUseCase) CreateRemovalRequest(context.Context, removal.CreateRemovalRequestCommand) (*removal.CreateRemovalRequestResult, error) {
	return nil, nil
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

func (s *stubRemovalUseCase) UpdateStatus(context.Context, removal.UpdateStatusCommand) (*removal.RemovalRequestDTO, error) {
	return nil, nil
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
