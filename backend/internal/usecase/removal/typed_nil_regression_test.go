package removal

// D-01 typed-nil バグの回帰テスト。
//
// Go では typed-nil ポインタをインターフェースに代入すると、
// インターフェース値は非 nil（型情報は持つが値は nil）になる。
// これにより u.notifier == nil ガードが false を返し、
// nil レシーバーでメソッドが呼ばれて panic が発生する。
//
// 正しい修正: main.go の配線で
//   var removalNotifier usecaseRemoval.RemovalNotifier
//   if cfg.SMTPHost != "" { removalNotifier = smtpNotifier }
// とすることで、SMTP 未設定時はインターフェース値が真の nil になる。

import (
	"context"
	"testing"
)

// typedNilTestNotifier は RemovalNotifier を実装するテスト専用構造体
type typedNilTestNotifier struct{}

func (n *typedNilTestNotifier) NotifyReceived(_ context.Context, _ ReceivedNotification) error {
	return nil
}
func (n *typedNilTestNotifier) NotifyResolved(_ context.Context, _ ResolvedNotification) error {
	return nil
}

func TestTypedNilInterfaceIsNotNil(t *testing.T) {
	t.Parallel()
	// typed-nil ポインタをインターフェースに代入すると non-nil になる（Go 仕様）
	var ptr *typedNilTestNotifier // typed-nil
	var iface RemovalNotifier = ptr

	if iface == nil {
		t.Fatal("typed-nil ポインタをインターフェースに代入した場合、iface == nil は false のはず")
	}
	// これが D-01 バグの根本原因: main.go で typed-nil の *email.SMTPNotifier を渡すと
	// u.notifier == nil チェックが通り抜け、nil レシーバーが呼ばれる
}

func TestProperNilInterfaceIsNil(t *testing.T) {
	t.Parallel()
	// インターフェース型の変数をゼロ値のまま使えば真の nil になる
	var iface RemovalNotifier // インターフェース型のゼロ値 = 真の nil

	if iface != nil {
		t.Fatal("インターフェース型のゼロ値は nil のはず")
	}
	// 修正後の main.go はこのパターンを使う:
	//   var removalNotifier usecaseRemoval.RemovalNotifier
	//   if cfg.SMTPHost != "" { removalNotifier = smtpNotifier }
}
