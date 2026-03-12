package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	"github.com/kuro48/idol-api/internal/usecase/idol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIdolUseCase はIdolUseCaseのモック
type MockIdolUseCase struct {
	mock.Mock
}

func (m *MockIdolUseCase) CreateIdol(ctx context.Context, cmd idol.CreateIdolCommand) (*idol.IdolDTO, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idol.IdolDTO), args.Error(1)
}

func (m *MockIdolUseCase) BulkCreateIdols(ctx context.Context, cmds []idol.CreateIdolCommand) (*idol.BulkResult, error) {
	args := m.Called(ctx, cmds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idol.BulkResult), args.Error(1)
}

func (m *MockIdolUseCase) GetIdol(ctx context.Context, query idol.GetIdolQuery) (*idol.IdolDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idol.IdolDTO), args.Error(1)
}

func (m *MockIdolUseCase) ListIdols(ctx context.Context, query idol.ListIdolsQuery) ([]*idol.IdolDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*idol.IdolDTO), args.Error(1)
}

func (m *MockIdolUseCase) SearchIdols(ctx context.Context, query idol.ListIdolsQuery) (*idol.SearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idol.SearchResult), args.Error(1)
}

func (m *MockIdolUseCase) UpdateIdol(ctx context.Context, cmd idol.UpdateIdolCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockIdolUseCase) DeleteIdol(ctx context.Context, cmd idol.DeleteIdolCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockIdolUseCase) RestoreIdol(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIdolUseCase) UpdateSocialLinks(ctx context.Context, cmd idol.UpdateSocialLinksCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockIdolUseCase) FindDuplicateCandidates(ctx context.Context, id string) ([]*idol.DuplicateCandidateDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*idol.DuplicateCandidateDTO), args.Error(1)
}

func (m *MockIdolUseCase) GetExternalIDs(ctx context.Context, id string) (map[string]string, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockIdolUseCase) UpdateExternalIDs(ctx context.Context, cmd idol.UpdateExternalIDsCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func setupIdolRouter(usecase idol.IdolUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.AuditContext())
	h := handlers.NewIdolHandler(usecase)
	router.POST("/idols", h.CreateIdol)
	router.GET("/idols/:id", h.GetIdol)
	router.GET("/idols", h.ListIdols)
	router.PUT("/idols/:id", h.UpdateIdol)
	router.DELETE("/idols/:id", h.DeleteIdol)
	router.PUT("/idols/:id/restore", h.RestoreIdol)
	return router
}

func TestCreateIdol_ValidInput(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	dto := &idol.IdolDTO{ID: "idol-001", Name: "テストアイドル"}
	mockUC.On("CreateIdol", mock.Anything, mock.Anything).Return(dto, nil)

	router := setupIdolRouter(mockUC)
	body := `{"name": "テストアイドル"}`
	req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCreateIdol_MissingName(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	router := setupIdolRouter(mockUC)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateIdol")
}

func TestCreateIdol_InvalidJSON(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	router := setupIdolRouter(mockUC)

	body := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateIdol")
}

func TestCreateIdol_UsecaseError(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("CreateIdol", mock.Anything, mock.Anything).Return(nil, errors.New("データベースエラー"))

	router := setupIdolRouter(mockUC)
	body := `{"name": "テストアイドル"}`
	req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetIdol_Found(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	dto := &idol.IdolDTO{ID: "idol-001", Name: "テストアイドル"}
	mockUC.On("GetIdol", mock.Anything, idol.GetIdolQuery{ID: "idol-001"}).Return(dto, nil)

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/idols/idol-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetIdol_NotFound(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("GetIdol", mock.Anything, idol.GetIdolQuery{ID: "nonexistent"}).Return(nil, errors.New("アイドルが見つかりません"))

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/idols/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListIdols_Success(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	result := &idol.SearchResult{
		Data: []*idol.IdolDTO{
			{ID: "idol-001", Name: "アイドルA"},
			{ID: "idol-002", Name: "アイドルB"},
		},
	}
	mockUC.On("SearchIdols", mock.Anything, mock.Anything).Return(result, nil)

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/idols", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListIdols_InvalidSortParam(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	router := setupIdolRouter(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/idols?sort=invalid_sort", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "SearchIdols")
}

func TestUpdateIdol_Success(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("UpdateIdol", mock.Anything, mock.Anything).Return(nil)

	router := setupIdolRouter(mockUC)
	body := `{"name": "更新されたアイドル"}`
	req := httptest.NewRequest(http.MethodPut, "/idols/idol-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUpdateIdol_InvalidJSON(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	router := setupIdolRouter(mockUC)

	body := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/idols/idol-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "UpdateIdol")
}

func TestDeleteIdol_Success(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("DeleteIdol", mock.Anything, idol.DeleteIdolCommand{ID: "idol-001"}).Return(nil)

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodDelete, "/idols/idol-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUC.AssertExpectations(t)
}

func TestDeleteIdol_NotFound(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("DeleteIdol", mock.Anything, idol.DeleteIdolCommand{ID: "nonexistent"}).Return(errors.New("アイドルが見つかりません"))

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodDelete, "/idols/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestRestoreIdol_Success(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("RestoreIdol", mock.Anything, "idol-001").Return(nil)

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodPut, "/idols/idol-001/restore", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestRestoreIdol_NotFound(t *testing.T) {
	mockUC := new(MockIdolUseCase)
	mockUC.On("RestoreIdol", mock.Anything, "nonexistent").Return(errors.New("アイドルが見つかりません"))

	router := setupIdolRouter(mockUC)
	req := httptest.NewRequest(http.MethodPut, "/idols/nonexistent/restore", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}
