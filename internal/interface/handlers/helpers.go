package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// getPathID はパスパラメータ "id" を取得する。
// 空文字の場合は 400 を返して false を返す。
func getPathID(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("IDは必須です"))
		return "", false
	}
	return id, true
}
