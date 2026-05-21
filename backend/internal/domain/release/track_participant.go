package release

import (
	"errors"
	"fmt"
)

// ParticipationStatus は楽曲への参加状態を表す。
type ParticipationStatus string

const (
	ParticipationStatusParticipating    ParticipationStatus = "participating"
	ParticipationStatusNotParticipating ParticipationStatus = "not_participating"
)

func NewParticipationStatus(value string) (ParticipationStatus, error) {
	status := ParticipationStatus(value)
	switch status {
	case ParticipationStatusParticipating, ParticipationStatusNotParticipating:
		return status, nil
	default:
		return "", fmt.Errorf("無効な参加状態です: %s", value)
	}
}

func (s ParticipationStatus) Value() string {
	return string(s)
}

// TrackParticipant は楽曲単位のアイドル参加情報を表す値オブジェクト。
type TrackParticipant struct {
	idolID   string
	status   ParticipationStatus
	position *string
}

func NewTrackParticipant(idolID string, status ParticipationStatus, position *string) (TrackParticipant, error) {
	if idolID == "" {
		return TrackParticipant{}, errors.New("アイドルIDは空にできません")
	}
	if _, err := NewParticipationStatus(status.Value()); err != nil {
		return TrackParticipant{}, err
	}
	if position != nil && len([]rune(*position)) > 100 {
		return TrackParticipant{}, errors.New("立ち位置は100文字以内である必要があります")
	}
	if status == ParticipationStatusNotParticipating && position != nil && *position != "" {
		return TrackParticipant{}, errors.New("不参加の場合は立ち位置を設定できません")
	}
	return TrackParticipant{idolID: idolID, status: status, position: position}, nil
}

func (p TrackParticipant) IdolID() string {
	return p.idolID
}

func (p TrackParticipant) Status() ParticipationStatus {
	return p.status
}

func (p TrackParticipant) Position() *string {
	return p.position
}
