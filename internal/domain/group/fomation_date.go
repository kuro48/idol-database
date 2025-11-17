package group

import (
	"errors"
	"fmt"
	"time"
)

type FormationDate struct {
	value time.Time
}

func NewFormationDate(year, month, day int) (FormationDate, error) {
	if year < 1900 || year > time.Now().Year() {
		return FormationDate{}, errors.New("無効な年です")
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	if date.After(time.Now()) {
		return FormationDate{}, errors.New("結成日は未来の日付にできません")
	}

	return FormationDate{value: date}, nil
}

func NewFormationDateFromString(dateStr string) (FormationDate, error) {
	if dateStr == "" {
		return FormationDate{}, nil
	}

	date, err := time.Parse("2006-01-02", dateStr)

	if err != nil {
		return FormationDate{}, fmt.Errorf("無効な日付形式です: %w", err)
	}

	if date.After(time.Now()) {
		return FormationDate{}, errors.New("結成日は未来の日付にできません")
	}

	return FormationDate{value: date}, nil
}

func (f FormationDate) Value() time.Time {
	return f.value
}

func (f FormationDate) String() string {
	if f.value.IsZero() {
		return ""
	}
	return f.value.Format("2006-01-02")
}

func (f FormationDate) IsEmpty() bool {
	return f.value.IsZero()
}
