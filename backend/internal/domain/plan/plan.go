// Package plan はAPIプランの種別と制限値を定義する
package plan

// Type はAPIプランの種別
type Type string

const (
	TypeFree      Type = "free"
	TypeDeveloper Type = "developer"
	TypeBusiness  Type = "business"
)

// Limits はプランごとの制限値
type Limits struct {
	// MonthlyRequests は1ヶ月あたりのリクエスト上限（0は無制限）
	MonthlyRequests int
	// WriteEnabled は write スコープ（POST/PUT/DELETE）が使えるか
	WriteEnabled bool
}

// GetLimits はプランの制限値を返す
func GetLimits(t Type) Limits {
	switch t {
	case TypeDeveloper:
		return Limits{MonthlyRequests: 50_000, WriteEnabled: true}
	case TypeBusiness:
		return Limits{MonthlyRequests: 500_000, WriteEnabled: true}
	default: // TypeFree
		return Limits{MonthlyRequests: 1_000, WriteEnabled: false}
	}
}

// MonthlyPrice は月額料金（円）を返す
func MonthlyPrice(t Type) int {
	switch t {
	case TypeDeveloper:
		return 1500
	case TypeBusiness:
		return 8000
	default:
		return 0
	}
}

// IsValid はプラン種別が有効かを返す
func IsValid(t Type) bool {
	switch t {
	case TypeFree, TypeDeveloper, TypeBusiness:
		return true
	default:
		return false
	}
}

// StripePriceIDKey は各プランに対応する Stripe Price ID の環境変数キーを返す
func StripePriceIDKey(t Type) string {
	switch t {
	case TypeDeveloper:
		return "STRIPE_PRICE_DEVELOPER"
	case TypeBusiness:
		return "STRIPE_PRICE_BUSINESS"
	default:
		return ""
	}
}
