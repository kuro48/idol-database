// Package usage はAPIキーの月次使用量カウンターを定義する
package usage

import "time"

// MonthlyUsage はAPIキーの月次リクエスト使用量
type MonthlyUsage struct {
	keyPrefix string // APIキーのルックアップ用プレフィックス
	yearMonth string // "YYYY-MM" 形式（例: "2026-04"）
	count     int    // 当月のリクエスト数
	limit     int    // プランの月間上限（0は無制限）
	updatedAt time.Time
}

// New は MonthlyUsage を新規作成する
func New(keyPrefix, yearMonth string, limit int) *MonthlyUsage {
	return &MonthlyUsage{
		keyPrefix: keyPrefix,
		yearMonth: yearMonth,
		count:     0,
		limit:     limit,
		updatedAt: time.Now(),
	}
}

// Reconstruct はDBから取得したデータで MonthlyUsage を再構築する
func Reconstruct(keyPrefix, yearMonth string, count, limit int, updatedAt time.Time) *MonthlyUsage {
	return &MonthlyUsage{
		keyPrefix: keyPrefix,
		yearMonth: yearMonth,
		count:     count,
		limit:     limit,
		updatedAt: updatedAt,
	}
}

// YearMonthOf は time.Time から "YYYY-MM" 文字列を返す
func YearMonthOf(t time.Time) string {
	return t.UTC().Format("2006-01")
}

// ExceedsLimit は使用量が上限に達しているかを返す
// limit == 0 は無制限とみなし、常に false を返す
func (u *MonthlyUsage) ExceedsLimit() bool {
	if u.limit == 0 {
		return false
	}
	return u.count >= u.limit
}

// Getters

func (u *MonthlyUsage) KeyPrefix() string    { return u.keyPrefix }
func (u *MonthlyUsage) YearMonth() string    { return u.yearMonth }
func (u *MonthlyUsage) Count() int           { return u.count }
func (u *MonthlyUsage) Limit() int           { return u.limit }
func (u *MonthlyUsage) UpdatedAt() time.Time { return u.updatedAt }
