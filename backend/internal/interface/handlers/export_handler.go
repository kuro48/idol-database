package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	domainExport "github.com/kuro48/idol-api/internal/domain/export"
	domainIdol "github.com/kuro48/idol-api/internal/domain/idol"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// exportService は ExportHandler が依存するサービス契約
type exportService interface {
	ExportIdols(ctx context.Context, format domainExport.ExportFormat, actor string) (*domainExport.ExportIdolsResult, error)
	ListExportLogs(ctx context.Context, limit int) ([]*domainExport.ExportLog, error)
}

// ExportHandler はエクスポートハンドラー
type ExportHandler struct {
	appService exportService
}

// NewExportHandler はエクスポートハンドラーを作成する
func NewExportHandler(appService exportService) *ExportHandler {
	return &ExportHandler{appService: appService}
}

// ExportIdols はアイドル一覧をエクスポートする
// @Summary      アイドルエクスポート
// @Description  全アイドルデータをJSON/JSONL形式でエクスポートする（管理者専用、レート制限あり）
// @Tags         export
// @Produce      application/json
// @Param        format query string false "出力形式 (json|jsonl)" default(json)
// @Success      200
// @Failure      429 {object} middleware.ErrorResponse
// @Failure      500 {object} middleware.ErrorResponse
// @Router       /admin/export/idols [get]
func (h *ExportHandler) ExportIdols(c *gin.Context) {
	format := domainExport.ExportFormat(c.DefaultQuery("format", "json"))
	actor := middleware.GetActor(c)

	result, err := h.appService.ExportIdols(middleware.AuditContextFor(c), format, actor)
	if err != nil {
		// レート制限エラーは 429 で返す
		if isRateLimitError(err) {
			c.JSON(http.StatusTooManyRequests, middleware.ErrorResponse{
				Code:    "RATE_LIMITED",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("エクスポートに失敗しました"))
		return
	}

	switch result.Format {
	case domainExport.ExportFormatJSONL:
		h.writeJSONL(c, result)
	default:
		h.writeJSON(c, result)
	}
}

// ListExportLogs はエクスポート実行履歴を返す
// @Summary      エクスポート実行履歴
// @Description  エクスポートの実行履歴を返す（管理者専用）
// @Tags         export
// @Produce      json
// @Param        limit query int false "件数上限（最大100、デフォルト50）"
// @Success      200
// @Router       /admin/export/logs [get]
func (h *ExportHandler) ListExportLogs(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}

	logs, err := h.appService.ListExportLogs(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, middleware.NewInternalError("実行履歴の取得に失敗しました"))
		return
	}

	type logResponse struct {
		ID          string `json:"id"`
		Resource    string `json:"resource"`
		Format      string `json:"format"`
		RecordCount int    `json:"record_count"`
		ExecutedBy  string `json:"executed_by"`
		Status      string `json:"status"`
		ErrorMsg    string `json:"error_msg,omitempty"`
		ExecutedAt  string `json:"executed_at"`
	}

	responses := make([]logResponse, 0, len(logs))
	for _, l := range logs {
		responses = append(responses, logResponse{
			ID:          l.ID(),
			Resource:    string(l.Resource()),
			Format:      string(l.Format()),
			RecordCount: l.RecordCount(),
			ExecutedBy:  l.ExecutedBy(),
			Status:      string(l.Status()),
			ErrorMsg:    l.ErrorMsg(),
			ExecutedAt:  l.ExecutedAt().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": responses, "count": len(responses)})
}

func (h *ExportHandler) writeJSON(c *gin.Context, result *domainExport.ExportIdolsResult) {
	records := idolsToExportRecords(result.Idols)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="idols_%s.json"`, result.LogID))
	c.Header("X-Export-Log-ID", result.LogID)
	c.Header("X-Record-Count", strconv.Itoa(len(records)))
	c.JSON(http.StatusOK, gin.H{
		"exported_at":  time.Now().UTC().Format(time.RFC3339),
		"record_count": len(records),
		"log_id":       result.LogID,
		"data":         records,
	})
}

func (h *ExportHandler) writeJSONL(c *gin.Context, result *domainExport.ExportIdolsResult) {
	records := idolsToExportRecords(result.Idols)

	var buf bytes.Buffer
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			continue
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="idols_%s.jsonl"`, result.LogID))
	c.Header("X-Export-Log-ID", result.LogID)
	c.Header("X-Record-Count", strconv.Itoa(len(records)))
	c.Data(http.StatusOK, "application/x-ndjson", buf.Bytes())
}

type idolExportRecord struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Birthdate   string            `json:"birthdate,omitempty"`
	AgencyID    *string           `json:"agency_id,omitempty"`
	SocialLinks map[string]string `json:"social_links,omitempty"`
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
	TagIDs      []string          `json:"tag_ids,omitempty"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

func idolsToExportRecords(idols []*domainIdol.Idol) []idolExportRecord {
	records := make([]idolExportRecord, 0, len(idols))
	for _, i := range idols {
		rec := idolExportRecord{
			ID:        i.ID().Value(),
			Name:      i.Name().Value(),
			AgencyID:  i.AgencyID(),
			TagIDs:    i.TagIDs(),
			CreatedAt: i.CreatedAt().Format(time.RFC3339),
			UpdatedAt: i.UpdatedAt().Format(time.RFC3339),
		}
		if i.Birthdate() != nil {
			rec.Birthdate = i.Birthdate().String()
		}
		if sl := i.SocialLinks(); sl != nil {
			m := make(map[string]string)
			if sl.Twitter() != nil {
				m["twitter"] = *sl.Twitter()
			}
			if sl.Instagram() != nil {
				m["instagram"] = *sl.Instagram()
			}
			if sl.TikTok() != nil {
				m["tiktok"] = *sl.TikTok()
			}
			if sl.YouTube() != nil {
				m["youtube"] = *sl.YouTube()
			}
			if sl.Facebook() != nil {
				m["facebook"] = *sl.Facebook()
			}
			if sl.Official() != nil {
				m["official"] = *sl.Official()
			}
			if sl.FanClub() != nil {
				m["fan_club"] = *sl.FanClub()
			}
			if len(m) > 0 {
				rec.SocialLinks = m
			}
		}
		if extIDs := i.ExternalIDs(); !extIDs.IsEmpty() {
			rawIDs := extIDs.All()
			m := make(map[string]string, len(rawIDs))
			for k, v := range rawIDs {
				m[string(k)] = v
			}
			rec.ExternalIDs = m
		}
		records = append(records, rec)
	}
	return records
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "レート制限")
}
