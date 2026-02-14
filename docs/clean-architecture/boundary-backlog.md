# CleanArch 責務重複・境界違反バックログ（Issue #13）

最終更新: 2026-02-14
入力資料: `docs/clean-architecture/dependency-map.md`

## 1. 目的

`application` と `usecase` の責務重複、および境界違反候補を優先度付きで整理し、移行順を固定する。

## 2. 優先度付きバックログ

| ID | Priority | テーマ | 根拠 | 対応方針 |
|---|---|---|---|---|
| CA-001 | High | Domain層のMongo依存除去 | `internal/domain/removal/removal_id.go:6`, `internal/domain/tag/value_object.go:7` | ID検証/生成をインフラまたは専用ユーティリティへ移管し、`domain` は文字列規約のみ保持 |
| CA-002 | High | Interface層のInfra依存除去 | `internal/interface/middleware/error.go:10`, `internal/interface/middleware/error.go:173` | DB固有エラー判定をAdapter層へ閉じ込め、HTTP層はアプリ共通エラーコードのみ解釈 |
| CA-003 | High | UseCase/Application責務の二重化解消 | `internal/usecase/idol/service.go:29`, `internal/application/idol/service.go:26` | UseCaseをI/Oポート中心へ再定義し、ApplicationはUseCase実装か薄いトランザクション層へ統合 |
| CA-004 | High | 複数Application横断依存の縮小 | `internal/usecase/removal/service.go:7` | ユースケース単位で必要ポートを定義し、`idol/group` 直参照をOutput Port化 |
| CA-005 | Medium | バリデーション責務の境界統一 | `internal/interface/handlers/idol_handler.go:127`, `internal/usecase/idol/query.go:79` | 入力バリデーション責務を「Adapterで形式」「UseCaseで業務」に分離し規約化 |
| CA-006 | Medium | 検索Query/ページネーション重複排除 | `internal/usecase/idol/query.go:55`, `internal/usecase/event/query.go:46`, `internal/usecase/tag/service.go:57` | 共通Pagination/Sortポリシーを `usecase/shared` に切り出す |
| CA-007 | Medium | エラーマッピングの一貫化 | `internal/interface/handlers/idol_handler.go:63`, `internal/interface/handlers/event_handler.go:43` | `middleware.WriteError` に統一し、HTTPステータス判定ルールを1箇所化 |
| CA-008 | Low | UseCase DTOのHTTPタグ依存解消 | `internal/usecase/event/command.go:5` | `json`/`binding` タグ付きDTOをinterface層へ移し、usecase層は純粋コマンド型に変更 |

## 3. 重複責務の具体例（抜粋）

| ユースケース層 | アプリケーション層 | 重複/曖昧点 |
|---|---|---|
| `CreateIdol` (`internal/usecase/idol/service.go:29`) | `CreateIdol` (`internal/application/idol/service.go:26`) | 両層が同じユースケース名で存在し責務境界が読みにくい |
| `UpdateIdol` (`internal/usecase/idol/service.go:79`) | `UpdateIdol` (`internal/application/idol/service.go:92`) | UseCaseがInput変換のみ、実処理はApplicationに集中 |
| `SearchEvents` (`internal/usecase/event/service.go:57`) | `SearchEvents` (`internal/application/event/service.go:113`) | 検索条件/ページング責務が2層に分散 |
| `CreateRemovalRequest` (`internal/usecase/removal/service.go:29`) | `CreateRemovalRequest` (`internal/application/removal/service.go:23`) | 存在確認・状態遷移・永続化が複層にまたがる |

## 4. 推奨実行順（#13 内）

1. `CA-001` と `CA-002` を先行（境界違反の直接解消）
2. `CA-003` と `CA-004` で UseCase/Application の責務再配置
3. `CA-005` から `CA-007` で入力・エラー・共通処理を統一
4. `CA-008` を最後に実施（影響範囲が広いが緊急度は低い）

## 5. #14 への接続ポイント

ADRで先に合意すべき項目:
- UseCaseを唯一のアプリケーション境界にするか
- Application層を残す場合の責務（トランザクション/オーケストレーション限定など）
- エラー型の正規仕様（Domain/UseCase/Infraの分類と変換地点）
