// Package audit は監査コンテキスト（作成者・更新者・ソース）を context.Context 経由で伝搬するユーティリティを提供します。
// infrastructure 層と interface 層の両方がインポートできる shared パッケージです。
package audit

import "context"

type contextKey int

const (
	actorKey  contextKey = iota
	sourceKey contextKey = iota
)

// WithActor はコンテキストにアクター（API キー識別子）をセットする
func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

// ActorFrom はコンテキストからアクターを取得する。未設定の場合は "anonymous" を返す
func ActorFrom(ctx context.Context) string {
	if v, ok := ctx.Value(actorKey).(string); ok && v != "" {
		return v
	}
	return "anonymous"
}

// WithSource はコンテキストにソース（"api", "import" など）をセットする
func WithSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, sourceKey, source)
}

// SourceFrom はコンテキストからソースを取得する。未設定の場合は "api" を返す
func SourceFrom(ctx context.Context) string {
	if v, ok := ctx.Value(sourceKey).(string); ok && v != "" {
		return v
	}
	return "api"
}
