# 実装ロードマップ

## 開発フェーズ概要

```
Phase 1 (MVP) → Phase 2 (申請フロー) → Phase 3 (ユーザー管理)
   4-6週間          3-4週間              4-5週間
```

---

## Phase 1: MVP（管理者のみ運用）

**目標:** 管理者が直接データを追加・編集・削除できる基本APIの構築

**期間:** 4-6週間

### Week 1-2: 基盤構築

#### タスク1: プロジェクト構造の整理（3日）
- [ ] ディレクトリ構造の作成
  ```bash
  mkdir -p cmd/api
  mkdir -p internal/{domain/{model,repository},usecase,interface/{handler,middleware,validator},infrastructure/{database,repository},config}
  mkdir -p pkg/{utils,errors}
  mkdir -p scripts
  ```
- [ ] `main.go` を `cmd/api/main.go` に移行
- [ ] `.env.example` 作成
- [ ] `.gitignore` 更新（.envを追加）
- [ ] README.md 作成

#### タスク2: 設定管理とDB接続（2日）
- [ ] `internal/config/` 実装
  - `config.go`: 設定構造体
  - `env.go`: 環境変数読み込み
- [ ] `internal/infrastructure/database/mongodb.go` 実装
  - MongoDB接続プール管理
  - 接続文字列を環境変数化
  - Pingテスト
- [ ] 環境変数の設定（.env）
  ```env
  PORT=8081
  GIN_MODE=debug
  MONGODB_URI=mongodb://admin:password@localhost:27017
  MONGODB_DATABASE=idol_api
  API_KEY=generate-with-openssl-rand
  ```

#### タスク3: ドメインモデル実装（2日）
- [ ] `internal/domain/model/idol.go`
  - Idol構造体
  - GroupMembership構造体
  - ドメインロジック（CalculateAge, UpdateIsActive）
- [ ] `internal/domain/model/group.go`
  - Group構造体
- [ ] `internal/domain/repository/` インターフェース定義
  - `idol_repository.go`
  - `group_repository.go`

#### タスク4: エラーハンドリング（1日）
- [ ] `pkg/errors/errors.go`
  - カスタムエラー型定義
  - 標準エラー（NotFound, Validation, Unauthorized等）
- [ ] `internal/interface/handler/response.go`
  - エラーレスポンスヘルパー

---

### Week 3-4: Idol API実装

#### タスク5: Repositoryレイヤー（3日）
- [ ] `internal/infrastructure/repository/idol_repository_impl.go`
  - Create
  - GetByID
  - Find（フィルタリング、ソート、ページネーション）
  - Update
  - Delete
  - FindDuplicate
  - Count
- [ ] MongoDBインデックス作成スクリプト
  ```bash
  # scripts/setup_indexes.sh
  ```

#### タスク6: Serviceレイヤー（3日）
- [ ] `internal/usecase/idol_service.go`
  - GetIdols（一覧取得）
  - GetIdolByID（詳細取得）
  - CreateIdol（登録）
    - グループ存在確認
    - 重複チェック
    - is_active自動計算
  - UpdateIdol（更新）
  - DeleteIdol（削除）
- [ ] ビジネスロジックのテスト

#### タスク7: Handlerレイヤー（2日）
- [ ] `internal/interface/handler/idol_handler.go`
  - GetIdols
  - GetIdolByID
  - SearchIdols（全文検索）
  - CreateIdol
  - UpdateIdol
  - PatchIdol
  - DeleteIdol
- [ ] `internal/interface/validator/idol_validator.go`
  - バリデーションルール実装

---

### Week 5: Group API実装

#### タスク8: Group Repository（2日）
- [ ] `internal/infrastructure/repository/group_repository_impl.go`
  - CRUD操作
  - メンバー一覧取得

#### タスク9: Group Service & Handler（2日）
- [ ] `internal/usecase/group_service.go`
- [ ] `internal/interface/handler/group_handler.go`
- [ ] `internal/interface/validator/group_validator.go`

#### タスク10: グループ削除時のカスケード処理（1日）
- [ ] グループ削除時にアイドルのGroupMembershipから該当エントリを削除

---

### Week 6: 認証・セキュリティ・最終調整

#### タスク11: 認証ミドルウェア（1日）
- [ ] `internal/interface/middleware/auth.go`
  - API Key検証
- [ ] API Key生成スクリプト
  ```bash
  # scripts/generate_api_key.sh
  openssl rand -hex 32
  ```

#### タスク12: セキュリティ強化（2日）
- [ ] `internal/interface/middleware/cors.go`
- [ ] `internal/interface/middleware/rate_limit.go`（簡易実装）
- [ ] `internal/interface/middleware/logger.go`
- [ ] 入力サニタイゼーション

#### タスク13: ルーティング統合（1日）
- [ ] `cmd/api/main.go` 完成
  - すべてのエンドポイント統合
  - ミドルウェア適用
  - エラーハンドリング

#### タスク14: テストとドキュメント（2日）
- [ ] 統合テスト
- [ ] Postmanコレクション作成（API動作確認用）
- [ ] README.md更新
- [ ] CLAUDE.md更新

---

## Phase 2: 申請・承認フロー

**目標:** 誰でもデータ申請可能、管理者が承認して反映

**期間:** 3-4週間

### Week 7-8: Submission API実装

#### タスク15: Submissionモデル（1日）
- [ ] `internal/domain/model/submission.go`
- [ ] `internal/domain/repository/submission_repository.go`

#### タスク16: Submission Repository & Service（3日）
- [ ] `internal/infrastructure/repository/submission_repository_impl.go`
- [ ] `internal/usecase/submission_service.go`
  - 申請作成
  - 申請一覧取得（管理者）
  - 申請承認
  - 申請却下

#### タスク17: Submission Handler（2日）
- [ ] `internal/interface/handler/submission_handler.go`
  - CreateSubmission（誰でも可能）
  - GetSubmissions（管理者のみ）
  - ApproveSubmission（管理者のみ）
  - RejectSubmission（管理者のみ）

#### タスク18: レート制限強化（2日）
- [ ] 申請APIのレート制限（10申請/日）
- [ ] IPベースの制限実装

---

### Week 9-10: スパム対策と最終調整

#### タスク19: スパム対策（3日）
- [ ] Email検証（オプション）
- [ ] 不適切コンテンツフィルタリング
- [ ] URLバリデーション

#### タスク20: 通知機能（オプション）（2日）
- [ ] 申請受付メール送信
- [ ] 承認/却下通知メール

#### タスク21: テストとドキュメント（2日）
- [ ] Submission APIのテスト
- [ ] ドキュメント更新

---

## Phase 3: ユーザー管理とJWT認証

**目標:** 信頼できるユーザーに編集権限を付与

**期間:** 4-5週間

### Week 11-12: 認証システム実装

#### タスク22: Adminモデルと認証機能（3日）
- [ ] `internal/domain/model/admin.go`
- [ ] `internal/domain/repository/admin_repository.go`
- [ ] パスワードハッシュ化（bcrypt）

#### タスク23: JWT実装（3日）
- [ ] JWT生成・検証ユーティリティ
- [ ] `internal/interface/middleware/jwt_auth.go`
- [ ] リフレッシュトークン機能

#### タスク24: 認証エンドポイント（2日）
- [ ] POST /api/v1/auth/register
- [ ] POST /api/v1/auth/login
- [ ] POST /api/v1/auth/refresh
- [ ] POST /api/v1/auth/logout

---

### Week 13-14: RBAC実装

#### タスク25: 権限管理（3日）
- [ ] Role定義（admin, trusted_editor, viewer）
- [ ] `internal/interface/middleware/rbac.go`
- [ ] 権限チェック機能

#### タスク26: ユーザー管理API（2日）
- [ ] GET /api/v1/admin/users（管理者のみ）
- [ ] PATCH /api/v1/admin/users/:id（権限変更）
- [ ] DELETE /api/v1/admin/users/:id（ユーザー削除）

#### タスク27: エンドポイント権限設定（2日）
- [ ] 各エンドポイントに適切な権限設定
- [ ] trusted_editorの直接編集を許可

---

### Week 15: 最終調整とリリース準備

#### タスク28: セキュリティ強化（2日）
- [ ] トークンブラックリスト
- [ ] 監査ログ実装
- [ ] セキュリティテスト

#### タスク29: ドキュメントと本番準備（3日）
- [ ] API仕様書最終版
- [ ] デプロイガイド作成
- [ ] 本番環境設定

---

## 継続的改善（Phase 3以降）

### 将来的な機能追加

#### 統計・分析機能
- [ ] アイドル統計API
- [ ] グループ統計API
- [ ] ダッシュボード機能

#### 検索・フィルタリング強化
- [ ] Elasticsearchエラスティック全文検索
- [ ] ファセット検索
- [ ] 関連アイドル推薦

#### 画像管理
- [ ] 画像アップロード機能
- [ ] サムネイル生成
- [ ] CDN統合

#### 監視・運用
- [ ] Prometheus + Grafana
- [ ] アラート設定
- [ ] パフォーマンス監視

#### 2段階認証
- [ ] TOTP実装
- [ ] バックアップコード

---

## 開発チェックリスト

### 各タスク完了時の確認項目

#### コード品質
- [ ] コードレビュー実施
- [ ] `go fmt` 実行
- [ ] `golangci-lint` 通過
- [ ] テストカバレッジ ≥ 70%

#### 機能
- [ ] 要件を満たしている
- [ ] エラーハンドリングが適切
- [ ] バリデーションが実装されている

#### セキュリティ
- [ ] 入力バリデーション
- [ ] 認証・認可チェック
- [ ] SQLインジェクション、XSS対策

#### ドキュメント
- [ ] コードコメント
- [ ] API仕様書更新
- [ ] CLAUDE.md更新

---

## マイルストーン

### M1: Phase 1完了（Week 6終了時）
- ✅ Idol & Group CRUD API動作
- ✅ API Key認証実装
- ✅ 基本的なセキュリティ対策完了
- ✅ ドキュメント整備

### M2: Phase 2完了（Week 10終了時）
- ✅ Submission API動作
- ✅ 申請・承認フロー実装
- ✅ スパム対策実装

### M3: Phase 3完了（Week 15終了時）
- ✅ JWT認証実装
- ✅ RBAC実装
- ✅ 本番環境デプロイ準備完了

---

## リスクと対策

### 技術的リスク
| リスク | 対策 |
|--------|------|
| MongoDB v2の学習曲線 | 公式ドキュメント参照、サンプルコード活用 |
| 複雑なクエリのパフォーマンス | インデックス最適化、explain()での分析 |
| 認証実装の脆弱性 | 既存ライブラリ活用、セキュリティレビュー |

### スケジュールリスク
| リスク | 対策 |
|--------|------|
| 想定より時間がかかる | Phase 2, 3を後回しにしてPhase 1を優先 |
| 仕様変更 | MVPに集中、拡張性を考慮した設計 |

---

## 次のステップ

### 今すぐ始められること

1. **環境構築**
   ```bash
   # ディレクトリ作成
   mkdir -p cmd/api internal/domain/model docs scripts

   # 依存関係追加
   go get github.com/joho/godotenv
   go get github.com/go-playground/validator/v10
   ```

2. **最初のタスク実行**
   - プロジェクト構造の整理（Week 1のタスク1）
   - 設定管理の実装（Week 1のタスク2）

3. **開発の進め方**
   - 1タスクずつ順番に実装
   - 各タスク完了後にコミット
   - 週末にレビューと調整

---

## サポートとリソース

### 参考ドキュメント
- [MongoDB Go Driver公式ドキュメント](https://www.mongodb.com/docs/drivers/go/current/)
- [Gin Webフレームワーク](https://gin-gonic.com/docs/)
- [Go Clean Architecture](https://github.com/bxcodec/go-clean-arch)

### コミュニティリソース
- Go公式フォーラム
- Stack Overflow
- GitHub Issues（依存ライブラリ）
