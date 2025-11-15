package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/domain/models"
	"github.com/kuro48/idol-api/internal/infrastructure/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestCreateIdol(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*models.Idol")).Return(nil)

		router := setupTestRouter()
		router.POST("/idols", handler.CreateIdol)

		reqBody := models.CreateIdolRequest{
			Name:        "テストアイドル",
			Group:       "テストグループ",
			Birthdate:   "2000-01-01",
			Nationality: "日本",
			ImageURL:    "https://example.com/image.jpg",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "アイドルを作成しました", response["message"])
		assert.NotNil(t, response["idol"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		router := setupTestRouter()
		router.POST("/idols", handler.CreateIdol)

		req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "リクエスト形式が不正です", response["error"])
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		mockRepo.On("Create", mock.AnythingOfType("*models.Idol")).Return(errors.New("database error"))

		router := setupTestRouter()
		router.POST("/idols", handler.CreateIdol)

		reqBody := models.CreateIdolRequest{
			Name: "テストアイドル",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/idols", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "アイドル作成に失敗しました", response["error"])

		mockRepo.AssertExpectations(t)
	})
}

func TestGetIdols(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		idols := []models.Idol{
			{
				ID:          bson.NewObjectID(),
				Name:        "アイドル1",
				Group:       "グループA",
				Birthdate:   "2000-01-01",
				Nationality: "日本",
			},
			{
				ID:          bson.NewObjectID(),
				Name:        "アイドル2",
				Group:       "グループB",
				Birthdate:   "2001-02-02",
				Nationality: "韓国",
			},
		}

		mockRepo.On("FindAll").Return(idols, nil)

		router := setupTestRouter()
		router.GET("/idols", handler.GetIdols)

		req := httptest.NewRequest(http.MethodGet, "/idols", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["count"])
		assert.NotNil(t, response["idols"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		mockRepo.On("FindAll").Return([]models.Idol{}, nil)

		router := setupTestRouter()
		router.GET("/idols", handler.GetIdols)

		req := httptest.NewRequest(http.MethodGet, "/idols", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["count"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		mockRepo.On("FindAll").Return(nil, errors.New("database error"))

		router := setupTestRouter()
		router.GET("/idols", handler.GetIdols)

		req := httptest.NewRequest(http.MethodGet, "/idols", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetIdol(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()
		idol := &models.Idol{
			ID:          id,
			Name:        "テストアイドル",
			Group:       "テストグループ",
			Birthdate:   "2000-01-01",
			Nationality: "日本",
		}

		mockRepo.On("FindByID", id.Hex()).Return(idol, nil)

		router := setupTestRouter()
		router.GET("/idols/:id", handler.GetIdol)

		req := httptest.NewRequest(http.MethodGet, "/idols/"+id.Hex(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response["idol"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()
		mockRepo.On("FindByID", id.Hex()).Return(nil, errors.New("アイドルが見つかりません"))

		router := setupTestRouter()
		router.GET("/idols/:id", handler.GetIdol)

		req := httptest.NewRequest(http.MethodGet, "/idols/"+id.Hex(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateIdol(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()
		updateReq := &models.UpdateIdolRequest{
			Name:  "更新後の名前",
			Group: "更新後のグループ",
		}

		mockRepo.On("Update", id.Hex(), mock.AnythingOfType("*models.UpdateIdolRequest")).Return(nil)

		router := setupTestRouter()
		router.PUT("/idols/:id", handler.UpdateIdol)

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/idols/"+id.Hex(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "アイドル情報を更新しました", response["message"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()

		router := setupTestRouter()
		router.PUT("/idols/:id", handler.UpdateIdol)

		req := httptest.NewRequest(http.MethodPut, "/idols/"+id.Hex(), bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()
		updateReq := &models.UpdateIdolRequest{
			Name: "更新後の名前",
		}

		mockRepo.On("Update", id.Hex(), mock.AnythingOfType("*models.UpdateIdolRequest")).Return(errors.New("database error"))

		router := setupTestRouter()
		router.PUT("/idols/:id", handler.UpdateIdol)

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/idols/"+id.Hex(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteIdol(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()
		mockRepo.On("Delete", id.Hex()).Return(nil)

		router := setupTestRouter()
		router.DELETE("/idols/:id", handler.DeleteIdol)

		req := httptest.NewRequest(http.MethodDelete, "/idols/"+id.Hex(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "アイドルを削除しました", response["message"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(repository.MockIdolRepository)
		handler := NewIdolHandler(mockRepo)

		id := bson.NewObjectID()
		mockRepo.On("Delete", id.Hex()).Return(errors.New("アイドルが見つかりません"))

		router := setupTestRouter()
		router.DELETE("/idols/:id", handler.DeleteIdol)

		req := httptest.NewRequest(http.MethodDelete, "/idols/"+id.Hex(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})
}
