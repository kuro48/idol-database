package errors

// ErrorCode はドメインエラーコード
type ErrorCode int

const (
	// バリデーション系 (400 Bad Request)
	ErrCodeIDGeneration ErrorCode = iota + 1
	ErrCodeNameValidation
	ErrCodeCountryValidation
	ErrCodeInvalidInput
	ErrCodeRequired
	ErrCodeInvalidFormat

	// 競合系 (409 Conflict)
	ErrCodeDuplicate ErrorCode = iota + 100

	// 不在系 (404 Not Found)
	ErrCodeNotFound ErrorCode = iota + 200
)

// IsBadRequest はバリデーション系エラーかチェック
func (c ErrorCode) IsBadRequest() bool {
	return c >= ErrCodeIDGeneration && c < ErrCodeDuplicate
}

// IsConflict は競合系エラーかチェック
func (c ErrorCode) IsConflict() bool {
	return c >= ErrCodeDuplicate && c < ErrCodeNotFound
}

// IsNotFound は不在系エラーかチェック
func (c ErrorCode) IsNotFound() bool {
	return c >= ErrCodeNotFound
}

// DomainError はドメイン層のエラー型
type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// New はDomainErrorを作成する
func New(code ErrorCode, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

// Wrap はエラーをDomainErrorでラップする
func Wrap(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}
