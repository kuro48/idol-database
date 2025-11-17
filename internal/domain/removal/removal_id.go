package removal

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RemovalID は削除申請の一意識別子
type RemovalID struct {
	value string
}

// NewRemovalID は新しい削除申請IDを生成する
func NewRemovalID(value string) (RemovalID, error) {
	if value == "" {
		return RemovalID{}, errors.New("削除申請IDは空にできません")
	}

	// ObjectIDの形式チェック
	if _, err := bson.ObjectIDFromHex(value); err != nil {
		return RemovalID{}, errors.New("無効な削除申請ID形式です")
	}

	return RemovalID{value: value}, nil
}

// Value はIDの値を返す
func (id RemovalID) Value() string {
	return id.value
}

// Equals は2つのIDが等しいかチェック
func (id RemovalID) Equals(other RemovalID) bool {
	return id.value == other.value
}
