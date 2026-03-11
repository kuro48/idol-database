// Package export はデータエクスポートのドメインモデルを定義する
package export

import "time"

// ExportFormat はエクスポート形式
type ExportFormat string

const (
	ExportFormatJSON  ExportFormat = "json"
	ExportFormatJSONL ExportFormat = "jsonl"
)

// ExportResource はエクスポート対象リソース
type ExportResource string

const (
	ExportResourceIdols    ExportResource = "idols"
	ExportResourceGroups   ExportResource = "groups"
	ExportResourceAgencies ExportResource = "agencies"
	ExportResourceEvents   ExportResource = "events"
)

// ExportStatus はエクスポートの状態
type ExportStatus string

const (
	ExportStatusCompleted ExportStatus = "completed"
	ExportStatusFailed    ExportStatus = "failed"
)

// ExportLog はエクスポート実行履歴
type ExportLog struct {
	id          string
	resource    ExportResource
	format      ExportFormat
	recordCount int
	executedBy  string
	status      ExportStatus
	errorMsg    string
	executedAt  time.Time
}

// NewExportLog は新しいエクスポートログを作成する
func NewExportLog(id string, resource ExportResource, format ExportFormat, executedBy string) *ExportLog {
	return &ExportLog{
		id:         id,
		resource:   resource,
		format:     format,
		executedBy: executedBy,
		status:     ExportStatusCompleted,
		executedAt: time.Now(),
	}
}

func (e *ExportLog) ID() string              { return e.id }
func (e *ExportLog) Resource() ExportResource { return e.resource }
func (e *ExportLog) Format() ExportFormat     { return e.format }
func (e *ExportLog) RecordCount() int         { return e.recordCount }
func (e *ExportLog) ExecutedBy() string       { return e.executedBy }
func (e *ExportLog) Status() ExportStatus     { return e.status }
func (e *ExportLog) ErrorMsg() string         { return e.errorMsg }
func (e *ExportLog) ExecutedAt() time.Time    { return e.executedAt }

func (e *ExportLog) SetRecordCount(n int) { e.recordCount = n }
func (e *ExportLog) MarkFailed(msg string) {
	e.status = ExportStatusFailed
	e.errorMsg = msg
}
