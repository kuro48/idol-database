package group

import (
	"errors"
	"fmt"
	"time"
)

type DisbandDate struct {
	value time.Time
}

func NewDisbandDate(year, month, day int) (DisbandDate, error) {
	if year < 1900 || year > time.Now().Year() {
		return DisbandDate{}, errors.New("無効な年です")
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	if date.After(time.Now()) {
		return DisbandDate{}, errors.New("解散日は未来の日付にできません")
	}

	return DisbandDate{value: date}, nil
}

func NewDisbandDateFromString(dateStr string) (DisbandDate, error) {
	if dateStr == "" {
		return DisbandDate{}, nil
	}

	date, err := time.Parse("2006-01-02", dateStr)

	if err != nil {
		return DisbandDate{}, fmt.Errorf("無効な日付形式です: %w", err)
	}

	if date.After(time.Now()) {
		return DisbandDate{}, errors.New("解散日は未来の日付にできません")
	}

	return DisbandDate{value: date}, nil
}

func (f DisbandDate) Value() time.Time {
	return f.value
}

func (f DisbandDate) String() string {
	if f.value.IsZero() {
		return ""
	}
	return f.value.Format("2006-01-02")
}

func (f DisbandDate) IsEmpty() bool {
	return f.value.IsZero()
}
