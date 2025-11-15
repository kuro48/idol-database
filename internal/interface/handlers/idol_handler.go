package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/domain/models"
	"github.com/kuro48/idol-api/internal/infrastructure/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IdolHandler struct {
	repo *repository.IdolRepository
}

func NewIdolHandler(repo *repository.IdolRepository) *IdolHandler {
	return &IdolHandler{repo: repo}
}

func (h *IdolHandler) CreateIdol(c *gin.Context) {
	var req models.CreateIdolRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H {
			"error": "リクエスト形式が不正です",
			"details": err.Error(),
		})
		return
	}

	idol := &models.Idol{
		ID:          bson.NewObjectID(),
        Name:        req.Name,
        Group:       req.Group,
        Birthdate:   req.Birthdate,
        Nationality: req.Nationality,
        ImageURL:    req.ImageURL,
    }

	if err := h.repo.Create(idol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H {
			"error": "アイドル作成に失敗しました",
            "details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
        "message": "アイドルを作成しました",
        "idol": idol,
    })
}

func (h *IdolHandler) GetIdols(c *gin.Context) {
	idols, err := h.repo.FindAll()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "アイドル一覧取得に失敗しました",
            "details": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "idols": idols,
        "count": len(idols),
    })
}

func (h *IdolHandler) GetIdol(c *gin.Context) {
	id := c.Param("id")

    idol, err := h.repo.FindByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "アイドルが見つかりません",
            "details": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "idol": idol,
    })
}

func (h *IdolHandler) UpdateIdol(c *gin.Context) {
	id := c.Param("id")

    var req models.UpdateIdolRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "リクエスト形式が不正です",
            "details": err.Error(),
        })
        return
    }

    if err := h.repo.Update(id, &req); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "アイドル更新に失敗しました",
            "details": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "アイドル情報を更新しました",
    })
}

func (h *IdolHandler) DeleteIdol(c *gin.Context) {
	id := c.Param("id")

    if err := h.repo.Delete(id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "アイドル削除に失敗しました",
            "details": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "アイドルを削除しました",
    })
}