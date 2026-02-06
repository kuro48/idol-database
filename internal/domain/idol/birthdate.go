package idol

import (
	"errors"
	"fmt"
	"time"
)

type Birthdate struct {
	value time.Time
}

func NewBirthdate(year, month, day int) (Birthdate, error) {
	if year < 1900 || year > time.Now().Year() {
		return Birthdate{}, errors.New("無効な年です")
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	if date.After(time.Now()) {
		return Birthdate{}, errors.New("生年月日は未来の日付にできません")
	}

	return Birthdate{value: date}, nil
}

func NewBirthdateFromString(dateStr string) (Birthdate, error) {
	if dateStr == "" {
		return Birthdate{}, nil // 空は許可
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Birthdate{}, fmt.Errorf("無効な日付形式です: %w", err)
	}

	if date.After(time.Now()) {
		return Birthdate{}, errors.New("生年月日は未来の日付にできません")
	}

	return Birthdate{value: date}, nil
}

func (b Birthdate) Value() time.Time {
	return b.value
}

func (b Birthdate) String() string {
	if b.value.IsZero() {
		return ""
	}
	return b.value.Format("2006-01-02")
}

func (b Birthdate) Age() int {
	if b.value.IsZero() {
		return 0
	}
	now := time.Now()
	age := now.Year() - b.value.Year()
	if now.YearDay() < b.value.YearDay() {
		age--
	}
	return age
}

func (b Birthdate) IsEmpty() bool {
	return b.value.IsZero()
}