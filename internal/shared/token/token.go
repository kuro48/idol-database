package token

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

const tokenBytes = 32

// Generate は公開アクセストークンを生成する。
func Generate() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// Hash はトークンの SHA-256 ハッシュを返す。
func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Verify は生トークンとハッシュを定数時間比較で検証する。
func Verify(raw, hashed string) bool {
	if raw == "" || hashed == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(Hash(raw)), []byte(hashed)) == 1
}
