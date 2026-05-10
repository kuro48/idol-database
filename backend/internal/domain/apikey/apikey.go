// Package apikey はDB-backed APIキーのドメインモデルを定義する
package apikey

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"time"

	"github.com/kuro48/idol-api/internal/domain/plan"
)

var hexColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

const (
	// keyBodyLen はキー本体のランダムバイト数（24バイト = 48文字の16進数）
	keyBodyLen = 24
	// KeyPrefix はAPIキーのプレフィックス
	KeyPrefix = "ik_live_"
	// lookupPrefixLen はルックアップに使用するプレフィックス長（キー全体の先頭N文字）
	lookupPrefixLen = 16
)

var objectIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

// APIKey はDB-backed APIキーエンティティ
type APIKey struct {
	id        string
	prefix    string // ルックアップ用プレフィックス（最初の16文字）
	keyHash   string // SHA-256(rawKey) の16進数文字列
	maskedKey string // 表示用（例: "ik_live_a1b2****cdef"）
	email     string // 所有者メールアドレス
	name      string // キーの説明（例: "My Production App"）
	planType  plan.Type
	isActive  bool
	createdAt time.Time
	oshiColor string // 推しメンカラー（CSS hex: "#FF69B4"）、空文字はデフォルト
}

// GenerateRawKey は新しい生のAPIキー文字列を生成する
// 形式: "ik_live_" + 48文字の16進数 = 56文字
// この値は生成時に一度だけ表示し、その後は保存しない
func GenerateRawKey() (string, error) {
	b := make([]byte, keyBodyLen)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("APIキーの生成に失敗しました")
	}
	return KeyPrefix + hex.EncodeToString(b), nil
}

// GenerateRawKeyFromSeed は seed と secret から決定的なAPIキーを生成する。
func GenerateRawKeyFromSeed(secret, seed string) (string, error) {
	if secret == "" || seed == "" {
		return "", errors.New("決定的なAPIキー生成に必要な secret または seed が不足しています")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(seed)); err != nil {
		return "", errors.New("APIキーの生成に失敗しました")
	}
	sum := mac.Sum(nil)
	return KeyPrefix + hex.EncodeToString(sum[:keyBodyLen]), nil
}

// HashKey は生のAPIキーをSHA-256でハッシュ化して16進数文字列で返す
func HashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}

// PrefixOf は生のAPIキーからルックアップ用プレフィックスを取り出す
func PrefixOf(rawKey string) string {
	if len(rawKey) < lookupPrefixLen {
		return rawKey
	}
	return rawKey[:lookupPrefixLen]
}

// MaskKey は生のAPIキーをマスクした表示用文字列を返す
// 例: "ik_live_a1b2****ef12"
func MaskKey(rawKey string) string {
	if len(rawKey) <= 12 {
		return rawKey
	}
	last4 := rawKey[len(rawKey)-4:]
	return rawKey[:12] + "****" + last4
}

// New はAPIキーエンティティを新規作成する
// id: MongoDB ObjectID hex (24文字)
// rawKey: 生のAPIキー（ハッシュ化して内部に保持し、rawKey自体は返さない）
func New(id, rawKey, email, name string, planType plan.Type) (*APIKey, error) {
	if !objectIDPattern.MatchString(id) {
		return nil, errors.New("無効なAPIキーIDです")
	}
	if rawKey == "" {
		return nil, errors.New("APIキーは必須です")
	}
	if email == "" {
		return nil, errors.New("メールアドレスは必須です")
	}
	if !plan.IsValid(planType) {
		return nil, errors.New("無効なプラン種別です")
	}

	return &APIKey{
		id:        id,
		prefix:    PrefixOf(rawKey),
		keyHash:   HashKey(rawKey),
		maskedKey: MaskKey(rawKey),
		email:     email,
		name:      name,
		planType:  planType,
		isActive:  true,
		createdAt: time.Now(),
		oshiColor: "",
	}, nil
}

// Reconstruct はDBから取得したデータでAPIKeyを再構築する
func Reconstruct(id, prefix, keyHash, maskedKey, email, name string, planType plan.Type, isActive bool, createdAt time.Time, oshiColor string) (*APIKey, error) {
	if !objectIDPattern.MatchString(id) {
		return nil, errors.New("無効なAPIキーIDです")
	}
	return &APIKey{
		id:        id,
		prefix:    prefix,
		keyHash:   keyHash,
		maskedKey: maskedKey,
		email:     email,
		name:      name,
		planType:  planType,
		isActive:  isActive,
		createdAt: createdAt,
		oshiColor: oshiColor,
	}, nil
}

// VerifyKey は生のAPIキーがこのエンティティのものか検証する
func (k *APIKey) VerifyKey(rawKey string) bool {
	return HashKey(rawKey) == k.keyHash
}

// Deactivate はAPIキーを無効化する
func (k *APIKey) Deactivate() {
	k.isActive = false
}

// Activate はAPIキーを有効化する
func (k *APIKey) Activate() {
	k.isActive = true
}

// ChangePlan はAPIキーのプラン種別を変更する
func (k *APIKey) ChangePlan(planType plan.Type) error {
	if !plan.IsValid(planType) {
		return errors.New("無効なプラン種別です")
	}
	k.planType = planType
	return nil
}

// UpdateOshiColor は推しメンカラーを更新する（空文字はクリア扱い）
func (k *APIKey) UpdateOshiColor(color string) error {
	if color != "" && !hexColorPattern.MatchString(color) {
		return errors.New("推しメンカラーは #RGB または #RRGGBB 形式で指定してください")
	}
	k.oshiColor = color
	return nil
}

// Getters

func (k *APIKey) ID() string           { return k.id }
func (k *APIKey) Prefix() string       { return k.prefix }
func (k *APIKey) KeyHash() string      { return k.keyHash }
func (k *APIKey) MaskedKey() string    { return k.maskedKey }
func (k *APIKey) Email() string        { return k.email }
func (k *APIKey) Name() string         { return k.name }
func (k *APIKey) PlanType() plan.Type  { return k.planType }
func (k *APIKey) IsActive() bool       { return k.isActive }
func (k *APIKey) CreatedAt() time.Time { return k.createdAt }
func (k *APIKey) OshiColor() string    { return k.oshiColor }
