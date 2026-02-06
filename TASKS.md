# 現状タスク一覧（2026-02-05）

## 最優先（バグ/破綻）
- [ ] タグID戦略の統一（UUID vs ObjectID）とMongoDB実装の整合
- [ ] `IdolRepository.Search` のデコード修正（`idolDocument`→`domain` 変換）
- [ ] `Birthdate` の空値を永続化で `nil` 扱いに統一（復元エラー回避）
- [ ] `Idol`/`Removal` 作成時に生成IDをエンティティへ反映
- [ ] `Removal.Requester` の型不整合修正（メール vs requester_type）
- [ ] Dockerイメージに `static/terms` を含める（利用規約APIが動かない問題）
- [ ] `event` のコマンドDTOに `json` タグ追加（bind不具合防止）
- [ ] `idol` の検索条件とスキーマ整合（`nationality`/`group_id`）

## 重要（整合性/設計/運用）
- [ ] ID体系の全体方針決定（ObjectID/UUID/文字列）
- [ ] `usecase`導入に伴うドキュメント更新（`docs/*` の参照修正）
- [ ] Swagger再生成（`swag init`）と `@host`/port の整合
- [ ] `config` テストの環境依存除去（`t.Setenv` 等）
- [ ] レートリミッタのメモリ増大対策（TTL/定期cleanup）
- [ ] エラーレスポンスの統一（400/404/409/500 の分類）
- [ ] `agency/group` のページネーション方針整理
- [ ] `include` の実装範囲と仕様整理

## 低優先（品質/開発体験）
- [ ] CI整備（`go test`/`go vet`/`golangci-lint`）
- [ ] 生成物整理（`main`/`coverage.out` の扱い）
- [ ] 監視・メトリクス導入（ログ構造化の強化含む）
- [ ] READMEの最新反映（運用フロー/依存/構成図の更新継続）

## 参考（既知のテスト失敗）
- `go test ./...` で `MONGODB_URI` 未設定により `config` テストが失敗
- `go-build` キャッシュの権限エラー（ローカル環境依存）
