package idol


// DomainError はドメイン層のエラー
type DomainError struct {
	message string
}

// NewDomainError は新しいドメインエラーを作成する
func NewDomainError(message string) *DomainError {
	return &DomainError{message: message}
}

// Error はエラーメッセージを返す
func (e *DomainError) Error() string {
	return e.message
}