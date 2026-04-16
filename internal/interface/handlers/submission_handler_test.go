package handlers_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/usecase/submission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubmissionUseCase は SubmissionUseCase のモック
type MockSubmissionUseCase struct {
	mock.Mock
}

func (m *MockSubmissionUseCase) CreateSubmission(ctx context.Context, cmd submission.CreateSubmissionCommand) (*submission.PublicSubmissionDTO, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*submission.PublicSubmissionDTO), args.Error(1)
}

func (m *MockSubmissionUseCase) GetSubmissionPublic(ctx context.Context, id string) (*submission.PublicSubmissionDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*submission.PublicSubmissionDTO), args.Error(1)
}

func (m *MockSubmissionUseCase) ListAllSubmissions(ctx context.Context) ([]*submission.SubmissionDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*submission.SubmissionDTO), args.Error(1)
}

func (m *MockSubmissionUseCase) ListPendingSubmissions(ctx context.Context) ([]*submission.SubmissionDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*submission.SubmissionDTO), args.Error(1)
}

func (m *MockSubmissionUseCase) UpdateStatus(ctx context.Context, cmd submission.UpdateStatusCommand) (*submission.SubmissionDTO, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*submission.SubmissionDTO), args.Error(1)
}

func (m *MockSubmissionUseCase) ReviseSubmission(ctx context.Context, cmd submission.ReviseSubmissionCommand) (*submission.PublicSubmissionDTO, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*submission.PublicSubmissionDTO), args.Error(1)
}

func setupSubmissionRouter(uc submission.SubmissionUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := handlers.NewSubmissionHandler(uc)
	router.POST("/submissions", h.CreateSubmission)
	router.GET("/submissions/:id", h.GetSubmission)
	router.GET("/submissions", h.ListAllSubmissions)
	router.GET("/submissions/pending", h.ListPendingSubmissions)
	router.PUT("/submissions/:id/status", h.UpdateStatus)
	router.PUT("/submissions/:id/revise", h.ReviseSubmission)
	return router
}

func newPublicDTO() *submission.PublicSubmissionDTO {
	return &submission.PublicSubmissionDTO{
		ID:         "sub-001",
		TargetType: "idol",
		Payload:    `{"name":"テスト"}`,
		SourceURLs: []string{"https://example.com"},
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func newAdminDTO() *submission.SubmissionDTO {
	return &submission.SubmissionDTO{
		ID:               "sub-001",
		TargetType:       "idol",
		Payload:          `{"name":"テスト"}`,
		SourceURLs:       []string{"https://example.com"},
		ContributorEmail: "user@example.com",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// --- CreateSubmission ---

func TestCreateSubmission_ValidInput(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("CreateSubmission", mock.Anything, mock.Anything).Return(newPublicDTO(), nil)

	router := setupSubmissionRouter(mockUC)
	body := `{"target_type":"idol","payload":{"name":"テスト"},"source_urls":["https://example.com"],"contributor_email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCreateSubmission_InvalidTargetType(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	router := setupSubmissionRouter(mockUC)

	body := `{"target_type":"unknown","payload":{"name":"テスト"},"source_urls":["https://example.com"],"contributor_email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateSubmission")
}

func TestCreateSubmission_MissingSourceURLs(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	router := setupSubmissionRouter(mockUC)

	body := `{"target_type":"idol","payload":{"name":"テスト"},"source_urls":[],"contributor_email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateSubmission")
}

func TestCreateSubmission_InvalidEmail(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	router := setupSubmissionRouter(mockUC)

	body := `{"target_type":"idol","payload":{"name":"テスト"},"source_urls":["https://example.com"],"contributor_email":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "CreateSubmission")
}

func TestCreateSubmission_UsecaseError(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("CreateSubmission", mock.Anything, mock.Anything).Return(nil, errors.New("DB error"))

	router := setupSubmissionRouter(mockUC)
	body := `{"target_type":"idol","payload":{"name":"テスト"},"source_urls":["https://example.com"],"contributor_email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/submissions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// --- GetSubmission ---

func TestGetSubmission_Found(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("GetSubmissionPublic", mock.Anything, "sub-001").Return(newPublicDTO(), nil)

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions/sub-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestGetSubmission_UsecaseError(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("GetSubmissionPublic", mock.Anything, "sub-999").Return(nil, errors.New("not found"))

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions/sub-999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// "not found" を含むエラーは middleware.WriteError が 404 にマッピングする
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// --- ListAllSubmissions ---

func TestListAllSubmissions_Success(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	dtos := []*submission.SubmissionDTO{newAdminDTO()}
	mockUC.On("ListAllSubmissions", mock.Anything).Return(dtos, nil)

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListAllSubmissions_Empty(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("ListAllSubmissions", mock.Anything).Return([]*submission.SubmissionDTO{}, nil)

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListAllSubmissions_UsecaseError(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("ListAllSubmissions", mock.Anything).Return(nil, errors.New("DB error"))

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// --- ListPendingSubmissions ---

func TestListPendingSubmissions_Success(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	dtos := []*submission.SubmissionDTO{newAdminDTO()}
	mockUC.On("ListPendingSubmissions", mock.Anything).Return(dtos, nil)

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestListPendingSubmissions_UsecaseError(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("ListPendingSubmissions", mock.Anything).Return(nil, errors.New("DB error"))

	router := setupSubmissionRouter(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/submissions/pending", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// --- UpdateStatus ---

func TestUpdateStatus_Approved(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	dto := newAdminDTO()
	dto.Status = "approved"
	mockUC.On("UpdateStatus", mock.Anything, mock.Anything).Return(dto, nil)

	router := setupSubmissionRouter(mockUC)
	body := `{"status":"approved","reviewed_by":"admin1"}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUpdateStatus_Rejected(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	dto := newAdminDTO()
	dto.Status = "rejected"
	mockUC.On("UpdateStatus", mock.Anything, mock.Anything).Return(dto, nil)

	router := setupSubmissionRouter(mockUC)
	body := `{"status":"rejected","reviewed_by":"admin1"}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUpdateStatus_NeedsRevision_MissingNote(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	router := setupSubmissionRouter(mockUC)

	body := `{"status":"needs_revision","reviewed_by":"admin1"}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "UpdateStatus")
}

func TestUpdateStatus_NeedsRevision_WithNote(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	dto := newAdminDTO()
	dto.Status = "needs_revision"
	dto.RevisionNote = "修正してください"
	mockUC.On("UpdateStatus", mock.Anything, mock.Anything).Return(dto, nil)

	router := setupSubmissionRouter(mockUC)
	body := `{"status":"needs_revision","reviewed_by":"admin1","revision_note":"修正してください"}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUpdateStatus_UsecaseError(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("UpdateStatus", mock.Anything, mock.Anything).Return(nil, errors.New("DB error"))

	router := setupSubmissionRouter(mockUC)
	body := `{"status":"approved","reviewed_by":"admin1"}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// --- ReviseSubmission ---

func TestReviseSubmission_Success(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("ReviseSubmission", mock.Anything, mock.Anything).Return(newPublicDTO(), nil)

	router := setupSubmissionRouter(mockUC)
	body := `{"payload":{"name":"修正後"},"source_urls":["https://example.com/new"]}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/revise", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestReviseSubmission_InvalidInput(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	router := setupSubmissionRouter(mockUC)

	body := `{"source_urls":[]}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/revise", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "ReviseSubmission")
}

func TestReviseSubmission_UsecaseError(t *testing.T) {
	mockUC := new(MockSubmissionUseCase)
	mockUC.On("ReviseSubmission", mock.Anything, mock.Anything).Return(nil, errors.New("not in needs_revision state"))

	router := setupSubmissionRouter(mockUC)
	body := `{"payload":{"name":"修正後"},"source_urls":["https://example.com/new"]}`
	req := httptest.NewRequest(http.MethodPut, "/submissions/sub-001/revise", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}
