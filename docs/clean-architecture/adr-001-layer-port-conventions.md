# ADR-001: レイヤ定義・Port方針・命名規約

- Status: Proposed
- Date: 2026-02-14
- Related Issues: #11, #12, #13, #14

## Context

現構成は `domain / application / usecase / interface / infrastructure` だが、以下の課題がある。

- `application` と `usecase` の責務が重複し、境界が曖昧
- Domain層にDB実装依存が混入
- エラー分類とHTTP変換規約が統一されていない

## Decision

### 1. ターゲットレイヤ（依存方向）

依存方向は次に統一する。

`Entity <- UseCase <- Interface Adapter <- Framework/DB`

- `Entity` (Domain): ビジネスルール、値オブジェクト、エンティティ
- `UseCase` (Application Business Rules): 入出力ポートを介したユースケース実行
- `Interface Adapter`: HTTPハンドラ、Presenter、Repository Adapter
- `Framework/DB`: Gin、MongoDBドライバ、外部APIクライアント

### 2. Port方針

UseCase層に入出力ポートを定義し、外側実装はAdapter層で行う。

- Input Port: ハンドラから呼ばれるユースケースAPI
- Output Port: 永続化や外部サービス呼び出しの抽象

規約:
- UseCase層は `gin` / `mongo` / `bson` をimportしない
- Domain層は外部ライブラリ依存を持たない（標準ライブラリのみ）
- Adapter層のみがFramework/DBの型を扱う

### 3. エラー分類規約

エラーは3分類とし、境界で明示的に変換する。

- Domain Error: ビジネス不変条件違反（例: 無効な値）
- UseCase Error: ユースケース実行上の業務エラー（例: 権限不足、前提未充足）
- Infra Error: DB/ネットワーク等の技術エラー

変換ルール:
- Domain/UseCase/Infra -> AdapterでHTTPステータスへマッピング
- 文字列判定ではなく型またはコードで判定する

### 4. ディレクトリ・命名規約

`internal` 配下の将来レイアウト:

```text
internal/
  domain/
    <context>/
      entity.go
      value_object.go
      error.go

  usecase/
    <context>/
      port_in.go
      port_out.go
      service.go
      dto.go
      error.go

  adapter/
    http/
      handler/
      presenter/
    persistence/
      mongodb/

  infrastructure/
    mongodb/
    logger/
```

命名規約:
- Input Port: `XXXUseCase` interface
- Output Port: `XXXRepository` / `XXXGateway` interface
- UseCase実装: `XXXService`
- Adapter実装: `MongoXXXRepository`, `HTTPXXXHandler`

### 5. 既存 `application` 層の扱い

移行期間中は `application` を暫定層として許容するが、新規機能は `usecase` に直接実装する。

- 既存 `application` は段階的に `usecase` または `adapter` へ吸収
- 吸収完了したコンテキストから `internal/application/<context>` を削除

## Consequences

- 利点: 依存方向が固定され、テスト容易性と差し替え容易性が上がる
- 欠点: 移行期間に一時的な重複実装が発生する
- 対策: コンテキスト単位で完了させ、完了後に旧層を即時削除する

## Migration Plan (Summary)

1. #15 で `tag/removal` をパイロット移行しテンプレート化
2. #16 で `idol/group/agency/event` へ水平展開
3. #17-#18 でテスト/CI境界チェックを導入
4. #19 でドキュメント最終更新と旧構成クローズ
