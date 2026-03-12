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
	"github.com/kuro48/idol-api/internal/usecase/group"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGroupUseCase はGroupUseCaseのモック
type MockGroupUseCase struct {
	mock.Mock
}

func (m *MockGroupUseCase) CreateGroup(ctx context.Context, cmd group.CreateGroupCommand) (*group.GroupDTO, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.GroupDTO), args.Error(1)
}

func (m *MockGroupUseCase) GetGroup(ctx context.Context, query group.GetGroupQuery) (*group.GroupDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*group.GroupDTO), args.Error(1)
}

func (m *MockGroupUseCase) ListGroup(ctx context.Context, query group.ListGroupQuery) ([]*group.GroupDTO, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*group.GroupDTO), args.Error(1)
}

func (m *MockGroupUseCase) UpdateGroup(ctx context.Context, cmd group.UpdateGroupCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockGroupUseCase) DeleteGroup(ctx context.Context, cmd group.DeleteGroupCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func setupGroupRouter(usecase group.GroupUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.AuditContext())
	h := handlers.NewGroupHandler(usecase)
	router.POST("/groups", h.CreateGroup)
	router.GET("/groups/:id", h.GetGroup)
	router.GET("/groups", h.ListGroup)
	router.PUT("/groups/:id", h.UpdateGroup)
	router.DELETE("/groups/:id", h.DeleteGroup)
	return router
}

func TestCreateGroup_ValidInput(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	dto := &group.GroupDTO{ID: "abc123", Name: "テストグループ"}
	mockUC.On("CreateGroup", mock.Anything, mock.Anything).Return(dto, nil)

	router := setupGroupRouter(mockUC)
	body := `{"name": "テストグループ"}`
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCreateGroup_InvalidInput_MissingName(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	router := setupGroupRouter(mockUC)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateGroup")
}

func TestCreateGroup_InvalidInput_InvalidJSON(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	router := setupGroupRouter(mockUC)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateGroup")
}

func TestCreateGroup_UsecaseError(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	mockUC.On("CreateGroup", mock.Anything, mock.Anything).Return(nil, errors.New("データベースエラー"))

	router := setupGroupRouter(mockUC)
	body := `{"name": "テストグループ"}`
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetGroup_Found(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	dto := &group.GroupDTO{ID: "abc123", Name: "テストグループ"}
	mockUC.On("GetGroup", mock.Anything, group.GetGroupQuery{ID: "abc123"}).Return(dto, nil)

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/groups/abc123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetGroup_NotFound(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	mockUC.On("GetGroup", mock.Anything, group.GetGroupQuery{ID: "nonexistent"}).Return(nil, errors.New("グループが見つかりません"))

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/groups/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListGroup_Success(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	dtos := []*group.GroupDTO{
		{ID: "abc123", Name: "グループA"},
		{ID: "def456", Name: "グループB"},
	}
	mockUC.On("ListGroup", mock.Anything, mock.Anything).Return(dtos, nil)

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListGroup_Empty(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	dtos := []*group.GroupDTO{}
	mockUC.On("ListGroup", mock.Anything, mock.Anything).Return(dtos, nil)

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListGroup_Error(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	mockUC.On("ListGroup", mock.Anything, mock.Anything).Return(nil, errors.New("データベースエラー"))

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUpdateGroup_ValidInput(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	mockUC.On("UpdateGroup", mock.Anything, mock.Anything).Return(nil)

	router := setupGroupRouter(mockUC)
	body := `{"name": "更新されたグループ"}`
	req := httptest.NewRequest(http.MethodPut, "/groups/abc123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUpdateGroup_InvalidJSON(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	router := setupGroupRouter(mockUC)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/groups/abc123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "UpdateGroup")
}

func TestDeleteGroup_Success(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	mockUC.On("DeleteGroup", mock.Anything, group.DeleteGroupCommand{ID: "abc123"}).Return(nil)

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodDelete, "/groups/abc123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUC.AssertExpectations(t)
}

func TestDeleteGroup_NotFound(t *testing.T) {
	mockUC := new(MockGroupUseCase)
	mockUC.On("DeleteGroup", mock.Anything, group.DeleteGroupCommand{ID: "nonexistent"}).Return(errors.New("グループが見つかりません"))

	router := setupGroupRouter(mockUC)
	req := httptest.NewRequest(http.MethodDelete, "/groups/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}
