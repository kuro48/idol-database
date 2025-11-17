# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

**ユーザー投稿型アイドル情報データベース**を段階的に構築するプロジェクト。

### 基本情報
- **言語**: Go 1.24.4
- **Webフレームワーク**: Gin
- **データベース**: MongoDB v2
- **アーキテクチャ**: DDD（ドメイン駆動設計）

### プロジェクト戦略
1. **Phase 1-2** (0-12ヶ月): ユーザー投稿型プラットフォームで基盤構築
2. **Phase 3** (6-12ヶ月): 事務所交渉のための実績構築
3. **Phase 4** (12ヶ月〜): 事務所公式データとのパートナーシップ

### 現在のステータス
- ✅ DDD構造での基本CRUD実装完了
- ✅ MongoDB接続・基本インフラ完成
- 🚧 **Phase 1 Week 1-2**: 法的保護機能を実装中

---

## 開発コマンド

### アプリケーション実行
```bash
# 開発環境での実行
go run cmd/api/main.go

# ビルド
go build -o idol-api cmd/api/main.go

# テスト
go test ./...

# 依存関係の更新
go mod tidy
```

### Docker
```bash
# MongoDB起動
docker-compose up -d

# MongoDB停止
docker-compose down

# ボリューム含めて削除
docker-compose down -v
```

---

## アーキテクチャ

### DDD（ドメイン駆動設計）の4層構造

```
internal/
├── domain/              # ドメイン層（ビジネスロジック）
│   └── idol/
│       ├── value_object.go  # 値オブジェクト
│       ├── idol.go          # エンティティ（Aggregate Root）
│       ├── repository.go    # リポジトリインターフェース
│       └── service.go       # ドメインサービス
│
├── application/         # アプリケーション層（ユースケース）
│   └── idol/
│       ├── command.go       # コマンドDTO
│       ├── query.go         # クエリDTO
│       └── service.go       # アプリケーションサービス
│
├── infrastructure/      # インフラ層（技術的詳細）
│   └── persistence/mongodb/
│       └── idol_repository.go
│
└── interface/           # プレゼンテーション層（外部I/F）
    └── handlers/
        └── idol_handler_ddd.go
```

### 設計原則
- **値オブジェクト**: イミュータブル、自己検証
- **エンティティ**: IDで識別、ビジネスルールをカプセル化
- **Aggregate Root**: Idol が集約のルート
- **リポジトリ**: データアクセスを抽象化
- **ドメインサービス**: 複数エンティティにまたがるビジネスロジック

---

## 重要な実装ガイドライン

### 1. 法的コンプライアンス

**プロバイダ責任制限法に基づく設計**:
- ユーザー投稿型プラットフォーム
- 運営者はデータ収集しない
- 削除申請に24時間以内対応
- 画像の直接ホスティング禁止（外部URLのみ）

### 2. DDD実装パターン

**新機能追加の手順**:
1. ドメイン層: エンティティ・値オブジェクト定義
2. ドメイン層: リポジトリインターフェース定義
3. アプリケーション層: コマンド・クエリDTO定義
4. アプリケーション層: アプリケーションサービス実装
5. インフラ層: リポジトリ実装（MongoDB）
6. プレゼンテーション層: HTTPハンドラー実装
7. テスト: 各層のユニットテスト

---

## 次のステップ（Phase 1 Week 1-2）

**実装優先順位**:
1. 削除申請機能（3日）
2. 利用規約・プライバシーポリシー（2日）
3. モデレーション機能（3日）

詳細は `docs/implementation-roadmap.md` を参照。

---

## 重要なリンク

- [README](README.md) - プロジェクト概要
- [実装ロードマップ](docs/implementation-roadmap.md) - 詳細な実装計画
- [法的ガイドライン](docs/legal-guidelines.md) - 法的リスクと対策
- [API仕様](docs/api-specification.md) - API詳細仕様
