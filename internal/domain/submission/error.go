package submission

import "errors"

// NewDomainError は新しいドメインエラーを作成する
func NewDomainError(msg string) error {
	return errors.New(msg)
}
