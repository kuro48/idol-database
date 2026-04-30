package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

const accessTokenHeader = "X-Access-Token"

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

// getAccessToken はヘッダーまたは query から公開アクセストークンを取得する。
// 未指定の場合は 401 を返して false を返す。
func getAccessToken(c *gin.Context) (string, bool) {
	token := c.GetHeader(accessTokenHeader)
	if token == "" {
		token = c.Query("access_token")
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, middleware.NewUnauthorizedError())
		return "", false
	}
	return token, true
}
