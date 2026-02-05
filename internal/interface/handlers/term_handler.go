package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// TermHandler は利用規約・プライバシーポリシーを扱うハンドラー
type TermHandler struct {
	staticPath string
}

// NewTermHandler はTermHandlerを作成する
func NewTermHandler(staticPath string) *TermHandler {
	return &TermHandler{
		staticPath: staticPath,
	}
}

// TermResponse は利用規約レスポンス
type TermResponse struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Format  string `json:"format"`
}

// ShowTermsOfService は利用規約を返す
func (h *TermHandler) ShowTermsOfService(c *gin.Context) {
	content, err := h.readTermFile("terms_of_service.md")
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("利用規約の読み込みに失敗しました"))
		return
	}

	c.JSON(http.StatusOK, TermResponse{
		Type:    "terms_of_service",
		Content: content,
		Format:  "markdown",
	})
}

// ShowPrivacyPolicy はプライバシーポリシーを返す
func (h *TermHandler) ShowPrivacyPolicy(c *gin.Context) {
	content, err := h.readTermFile("privacy_policy.md")
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("プライバシーポリシーの読み込みに失敗しました"))
		return
	}

	c.JSON(http.StatusOK, TermResponse{
		Type:    "privacy_policy",
		Content: content,
		Format:  "markdown",
	})
}

// readTermFile は利用規約ファイルを読み込む
func (h *TermHandler) readTermFile(filename string) (string, error) {
	filePath := filepath.Join(h.staticPath, "terms", filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
