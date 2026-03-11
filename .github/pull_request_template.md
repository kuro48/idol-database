## 概要
<!-- 変更内容を簡潔に説明 -->

## 変更種別
- [ ] バグ修正
- [ ] 新機能追加
- [ ] リファクタリング
- [ ] ドキュメント更新
- [ ] CI/設定変更

## レイヤ境界チェック
<!-- CleanArch 依存方向ルールに従っているか確認 -->

- [ ] `domain` 層が `application/usecase/interface/infrastructure` をインポートしていない
- [ ] `application` 層が `usecase/interface` をインポートしていない
- [ ] `infrastructure` 層が `application/usecase/interface` をインポートしていない
- [ ] ハンドラーが具体型ではなく Input Port インターフェースを使用している

> ローカル確認コマンド:
> ```bash
> go build ./cmd/api
> go test ./...
> ```

## テスト
- [ ] 新規ユースケースに単体テストを追加した
- [ ] `go test ./...` がパスしている

## 関連 Issue
<!-- Closes #xxx -->
