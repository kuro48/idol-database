// Package export はデータエクスポートのアプリケーションサービス
package export

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	domainExport "github.com/kuro48/idol-api/internal/domain/export"
)

// rateLimitDuration はエクスポートのレート制限間隔（同一アクターは1分に1回まで）
const rateLimitDuration = 1 * time.Minute

// ApplicationService はエクスポートアプリケーションサービス
type ApplicationService struct {
	logRepo domainExport.LogRepository
	idolApp *appIdol.ApplicationService
}

// NewApplicationService はアプリケーションサービスを作成する
func NewApplicationService(logRepo domainExport.LogRepository, idolApp *appIdol.ApplicationService) *ApplicationService {
	return &ApplicationService{logRepo: logRepo, idolApp: idolApp}
}

// ExportIdols はアイドル一覧をエクスポートする
func (s *ApplicationService) ExportIdols(ctx context.Context, format domainExport.ExportFormat, actor string) (*domainExport.ExportIdolsResult, error) {
	// レート制限チェック
	since := time.Now().Add(-rateLimitDuration)
	lastLog, err := s.logRepo.FindLastByActor(ctx, actor, since)
	if err != nil {
		return nil, fmt.Errorf("レート制限チェックエラー: %w", err)
	}
	if lastLog != nil {
		remaining := rateLimitDuration - time.Since(lastLog.ExecutedAt())
		return nil, fmt.Errorf("レート制限: あと %.0f 秒後に再試行してください", remaining.Seconds())
	}

	// フォーマット検証
	if format != domainExport.ExportFormatJSON && format != domainExport.ExportFormatJSONL {
		format = domainExport.ExportFormatJSON
	}

	logID := generateExportID()
	exportLog := domainExport.NewExportLog(logID, domainExport.ExportResourceIdols, format, actor)

	// データ取得
	idols, err := s.idolApp.ListIdols(ctx)
	if err != nil {
		exportLog.MarkFailed(err.Error())
		if saveErr := s.logRepo.Save(ctx, exportLog); saveErr != nil {
			slog.Warn("エクスポート失敗ログの保存に失敗しました", "error", saveErr)
		}
		return nil, fmt.Errorf("アイドルデータ取得エラー: %w", err)
	}

	exportLog.SetRecordCount(len(idols))
	if err := s.logRepo.Save(ctx, exportLog); err != nil {
		// ログ保存失敗でもエクスポート自体は続行
		slog.Warn("エクスポートログの保存に失敗しました", "error", err)
	}

	return &domainExport.ExportIdolsResult{
		Idols:  idols,
		Format: format,
		LogID:  logID,
	}, nil
}

// ListExportLogs はエクスポート実行履歴を返す
func (s *ApplicationService) ListExportLogs(ctx context.Context, limit int) ([]*domainExport.ExportLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.logRepo.FindRecent(ctx, limit)
}

func generateExportID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand.Read failed: %v", err))
	}
	return hex.EncodeToString(b)
}
