// Package id はシステム全体のID生成を一元管理するパッケージです。
//
// ID体系: MongoDB ObjectID の16進数文字列表現（24文字）を使用する。
// 全aggregateはこのパッケージのGenerate()を使用してIDを生成すること。
//
// 方針:
//   - 新規作成されるすべてのエンティティIDはMongoDB ObjectID hex文字列とする
//   - MongoDBドキュメントの _id フィールドはstring型で保持する（後方互換性維持のため）
//   - 既存のUnixナノ秒形式IDを持つデータの移行はスコープ外（別途移行対応を検討すること）
package id

import "go.mongodb.org/mongo-driver/v2/bson"

// Generate は新しいObjectID hex文字列を生成する。
// 生成されるIDはMongoDB ObjectIDの16進数表現（24文字の英数字）である。
func Generate() string {
	return bson.NewObjectID().Hex()
}

// IsValid はIDが有効なObjectID hex文字列かどうかをチェックする。
// 空文字列はfalseを返す。
func IsValid(id string) bool {
	if id == "" {
		return false
	}
	_, err := bson.ObjectIDFromHex(id)
	return err == nil
}
