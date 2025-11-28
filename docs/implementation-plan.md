# Idol API 実装計画 - Phase 1 MVP完成

**作成日**: 2025-11-17
**期間**: 2-3週間
**ゴール**: 法的保護を備えた安全なMVPの完成

---

## 🎯 Phase 1 目標

法的保護（利用規約、プライバシーポリシー、モデレーション機能）を備えた、安全に運用可能なMVPを完成させる。

---

## 🔴 Week 1: セキュリティと法的保護（5-7日）

### Task 1.1: 環境変数管理の改善 ⚠️ 緊急

**優先度**: 🔴 Critical
**工数**: 2-3時間
**担当者**: Backend
**期限**: Day 1

#### 目的
MongoDB接続文字列のハードコードを解消し、セキュリティリスクを排除

#### 実装内容
- [ ] `internal/config/config.go`の拡張
  - MongoDB URI、Database名、Server Port等の環境変数読み込み
  - デフォルト値の設定
  - 必須パラメータのバリデーション

- [ ] `.env.example`の整備
  ```env
  # MongoDB Configuration
  MONGODB_URI=mongodb://admin:password@localhost:27017/?authSource=admin
  MONGODB_DATABASE=idol_database

  # Server Configuration
  SERVER_PORT=8081
  GIN_MODE=debug

  # Security
  JWT_SECRET=your-secret-key-here
  ```

- [ ] `cmd/api/main.go`の修正
  - ハードコードされた接続文字列を削除
  - config経由での設定読み込み

#### 完了条件
- ✅ 環境変数からの設定読み込みが動作
- ✅ `.env.local`なしで起動するとエラーが出る
- ✅ README.mdのセットアップ手順更新

#### 技術的詳細
```go
// internal/config/config.go に追加
type MongoDBConfig struct {
    URI      string
    Database string
}

type ServerConfig struct {
    Port string
    Mode string
}

type Config struct {
    MongoDB MongoDBConfig
    Server  ServerConfig
}

func LoadConfig() (*Config, error) {
    // godotenvでの環境変数読み込み
    // 必須パラメータの検証
    // Config構造体の生成
}
```

#### 依存関係
なし（独立して実装可能）

---

### Task 1.2: 利用規約・プライバシーポリシーの作成

**優先度**: 🔴 High
**工数**: 4-6時間
**担当者**: Legal + Backend
**期限**: Day 2

#### 目的
プロバイダ責任制限法に基づく法的保護の確立

#### 実装内容

##### 1.2.1 利用規約の作成（2-3時間）
- [ ] 利用規約テキストの作成
  - サービス概要
  - ユーザーの責任（投稿内容の責任）
  - 禁止事項（肖像権侵害、虚偽情報など）
  - 削除申請手続き
  - 免責事項
  - 準拠法と管轄裁判所

- [ ] 利用規約エンドポイントの実装
  ```
  GET /api/v1/terms
  ```

##### 1.2.2 プライバシーポリシーの作成（2-3時間）
- [ ] プライバシーポリシーテキストの作成
  - 収集する個人情報の種類
  - 利用目的
  - 第三者提供の有無
  - Cookie利用について
  - お問い合わせ窓口

- [ ] プライバシーポリシーエンドポイントの実装
  ```
  GET /api/v1/privacy
  ```

##### 1.2.3 実装方法
**Option A**: 静的HTMLファイル（推奨）
```
internal/interface/static/
  ├── terms.html
  └── privacy.html
```

**Option B**: DBに保存してAPI経由
```go
// 更新履歴も管理可能
type Policy struct {
    ID        string
    Type      string // "terms" or "privacy"
    Content   string
    Version   string
    CreatedAt time.Time
}
```

#### 完了条件
- ✅ 利用規約が閲覧可能
- ✅ プライバシーポリシーが閲覧可能
- ✅ 削除申請フローへの利用規約リンク追加

#### 技術的詳細
```go
// Ginでの静的ファイル配信
router.StaticFile("/terms", "./internal/interface/static/terms.html")
router.StaticFile("/privacy", "./internal/interface/static/privacy.html")

// またはAPIエンドポイント
router.GET("/api/v1/terms", handlers.GetTerms)
router.GET("/api/v1/privacy", handlers.GetPrivacy)
```

#### 依存関係
Task 1.1の完了（環境設定が整った状態で実装）

---

### Task 1.3: モデレーション機能の実装

**優先度**: 🔴 High
**工数**: 1-2日
**担当者**: Backend
**期限**: Day 3-5

#### 目的
違法コンテンツの早期発見と削除対応の迅速化

#### 実装内容

##### 1.3.1 禁止キーワードフィルタリング（4-6時間）
- [ ] ドメイン層: 禁止キーワードチェッカー
  ```go
  // internal/domain/moderation/keyword_filter.go
  type KeywordFilter struct {
      prohibitedWords []string
  }

  func (f *KeywordFilter) ContainsProhibited(text string) bool
  func (f *KeywordFilter) GetMatchedWords(text string) []string
  ```

- [ ] 設定ファイル: 禁止キーワードリスト
  ```yaml
  # config/prohibited_keywords.yml
  keywords:
    - "差別用語1"
    - "差別用語2"
    # ...
  ```

- [ ] アプリケーション層: アイドル/グループ作成時のチェック
  ```go
  // CreateIdol内で呼び出し
  if filter.ContainsProhibited(cmd.Name) {
      return nil, errors.New("禁止されたキーワードが含まれています")
  }
  ```

##### 1.3.2 通報機能の実装（4-6時間）
- [ ] ドメインモデル: Report
  ```go
  // internal/domain/report/report.go
  type Report struct {
      id           ReportID
      targetType   TargetType  // "idol" or "group"
      targetID     string
      reportType   ReportType  // "inappropriate", "copyright", "privacy"
      reason       string
      reporterEmail string
      status       ReportStatus // "pending", "reviewing", "resolved", "rejected"
      createdAt    time.Time
      updatedAt    time.Time
  }
  ```

- [ ] APIエンドポイント
  ```
  POST   /api/v1/reports           # 通報作成
  GET    /api/v1/reports           # 通報一覧（管理者のみ）
  GET    /api/v1/reports/:id       # 通報詳細（管理者のみ）
  PUT    /api/v1/reports/:id       # ステータス更新（管理者のみ）
  ```

##### 1.3.3 簡易管理ダッシュボード（4-6時間）
- [ ] 管理用エンドポイント
  ```
  GET /api/v1/admin/dashboard/stats    # 統計情報
  GET /api/v1/admin/pending-reviews    # 要確認コンテンツ一覧
  ```

- [ ] ダッシュボードデータ
  ```json
  {
    "pending_removal_requests": 5,
    "pending_reports": 3,
    "total_idols": 120,
    "total_groups": 15,
    "recent_activities": [...]
  }
  ```

#### 完了条件
- ✅ 禁止キーワードを含む投稿が拒否される
- ✅ 通報機能が動作する
- ✅ 管理者が未処理の通報を確認できる

#### 技術的詳細
```go
// 禁止キーワードチェックの組み込み
func (s *ApplicationService) CreateIdol(ctx context.Context, cmd CreateIdolCommand) (*IdolDTO, error) {
    // 既存のバリデーション
    name, err := idol.NewIdolName(cmd.Name)
    if err != nil {
        return nil, err
    }

    // 禁止キーワードチェック追加
    if s.keywordFilter.ContainsProhibited(cmd.Name) {
        return nil, errors.New("投稿内容に問題が検出されました")
    }

    // 以降の処理...
}
```

#### 依存関係
Task 1.1, 1.2の完了（設定とポリシーが整った状態）

---

## 🟡 Week 2: データ整合性と品質向上（6-9日）

### Task 2.1: エラーハンドリングの統一

**優先度**: 🟡 Medium
**工数**: 1日
**担当者**: Backend
**期限**: Day 6-7

#### 目的
エラーレスポンスの一貫性確保とデバッグ性向上

#### 実装内容

##### 2.1.1 カスタムエラー型の定義（2-3時間）
- [ ] 共通エラー型の作成
  ```go
  // internal/domain/errors/errors.go
  type DomainError struct {
      Code    string
      Message string
      Cause   error
  }

  var (
      ErrNotFound       = &DomainError{Code: "NOT_FOUND", Message: "リソースが見つかりません"}
      ErrAlreadyExists  = &DomainError{Code: "ALREADY_EXISTS", Message: "既に存在します"}
      ErrInvalidInput   = &DomainError{Code: "INVALID_INPUT", Message: "入力が不正です"}
      ErrUnauthorized   = &DomainError{Code: "UNAUTHORIZED", Message: "認証が必要です"}
      ErrForbidden      = &DomainError{Code: "FORBIDDEN", Message: "権限がありません"}
  )
  ```

##### 2.1.2 HTTPステータスコードのマッピング（2-3時間）
- [ ] エラーハンドラーミドルウェア
  ```go
  // internal/interface/middleware/error_handler.go
  func ErrorHandler() gin.HandlerFunc {
      return func(c *gin.Context) {
          c.Next()

          if len(c.Errors) > 0 {
              err := c.Errors.Last().Err

              var domainErr *errors.DomainError
              if errors.As(err, &domainErr) {
                  statusCode := mapErrorToStatus(domainErr)
                  c.JSON(statusCode, gin.H{
                      "error": domainErr.Code,
                      "message": domainErr.Message,
                  })
              } else {
                  c.JSON(500, gin.H{
                      "error": "INTERNAL_ERROR",
                      "message": "内部エラーが発生しました",
                  })
              }
          }
      }
  }
  ```

##### 2.1.3 エラーレスポンス形式の統一（1-2時間）
- [ ] 標準エラーレスポンス
  ```json
  {
    "error": "NOT_FOUND",
    "message": "指定されたアイドルが見つかりません",
    "details": {
      "field": "id",
      "value": "invalid_id"
    }
  }
  ```

#### 完了条件
- ✅ 全エンドポイントで統一されたエラーレスポンス
- ✅ 適切なHTTPステータスコードの返却
- ✅ ログ出力の構造化

#### 依存関係
なし（独立して実装可能）

---

### Task 2.2: バリデーション強化

**優先度**: 🟡 Medium
**工数**: 1日
**担当者**: Backend
**期限**: Day 6-7

#### 実装内容

##### 2.2.1 リクエストバリデーション（3-4時間）
- [ ] バリデータの導入（`go-playground/validator`）
  ```go
  type CreateIdolCommand struct {
      Name        string  `json:"name" validate:"required,min=1,max=100"`
      Group       string  `json:"group" validate:"omitempty,max=100"`
      Birthdate   *string `json:"birthdate" validate:"omitempty,datetime=2006-01-02"`
      Nationality string  `json:"nationality" validate:"required,min=1,max=50"`
      ImageURL    string  `json:"image_url" validate:"omitempty,url"`
  }
  ```

##### 2.2.2 ドメインバリデーションの強化（2-3時間）
- [ ] 値オブジェクトのバリデーション強化
  ```go
  // 既存のバリデーションを強化
  func NewIdolName(value string) (IdolName, error) {
      trimmed := strings.TrimSpace(value)
      if trimmed == "" {
          return IdolName{}, errors.New("名前は空にできません")
      }
      if len(trimmed) > 100 {
          return IdolName{}, errors.New("名前は100文字以内にしてください")
      }
      // 禁止文字チェック
      if containsInvalidChars(trimmed) {
          return IdolName{}, errors.New("使用できない文字が含まれています")
      }
      return IdolName{value: trimmed}, nil
  }
  ```

##### 2.2.3 バリデーションエラーメッセージの改善（1-2時間）
- [ ] 多言語対応の準備
- [ ] フィールドごとのエラーメッセージ

#### 完了条件
- ✅ 不正なリクエストが適切に拒否される
- ✅ わかりやすいエラーメッセージ
- ✅ バリデーションルールのドキュメント化

---

### Task 2.3: アイドル-グループ関連付け

**優先度**: 🟡 Medium
**工数**: 1-2日
**担当者**: Backend
**期限**: Day 8-9

#### 目的
データ整合性の向上とリレーショナルなデータ管理

#### 実装内容

##### 2.3.1 ドメインモデルの変更（3-4時間）
- [ ] Idolエンティティの修正
  ```go
  // 変更前
  type Idol struct {
      // ...
      group string  // 文字列
  }

  // 変更後
  type Idol struct {
      // ...
      groupID *group.GroupID  // GroupIDへの参照（オプショナル）
  }
  ```

- [ ] マイグレーション戦略
  - 既存データの移行スクリプト
  - 文字列グループ名からGroupIDへの変換

##### 2.3.2 新規APIエンドポイント（2-3時間）
- [ ] グループに所属するアイドル一覧
  ```
  GET /api/v1/groups/:id/idols
  ```

- [ ] アイドルの所属グループ変更
  ```
  PUT /api/v1/idols/:id/group
  {
    "group_id": "507f1f77bcf86cd799439011"
  }
  ```

##### 2.3.3 整合性チェック（2-3時間）
- [ ] グループ削除時のチェック
  ```go
  func (s *GroupService) Delete(ctx context.Context, id GroupID) error {
      // 所属アイドルがいるか確認
      idols, err := s.idolRepo.FindByGroupID(ctx, id)
      if err != nil {
          return err
      }
      if len(idols) > 0 {
          return errors.New("所属アイドルが存在するため削除できません")
      }
      return s.repo.Delete(ctx, id)
  }
  ```

#### 完了条件
- ✅ アイドルとグループが正しく関連付けられる
- ✅ グループから所属アイドルを取得できる
- ✅ データ整合性が保たれる

#### 依存関係
なし（ただしテストデータのバックアップ推奨）

---

### Task 2.4: 削除申請の自動処理

**優先度**: 🟡 Medium
**工数**: 1-2日
**担当者**: Backend
**期限**: Day 10

#### 実装内容

##### 2.4.1 承認時の自動削除（3-4時間）
- [ ] RemovalRequestのステータス更新時の処理
  ```go
  func (s *ApplicationService) ApproveRemovalRequest(ctx context.Context, id string) error {
      req, err := s.repo.FindByID(ctx, id)
      if err != nil {
          return err
      }

      // ステータスを承認に更新
      req.Approve()
      if err := s.repo.Update(ctx, req); err != nil {
          return err
      }

      // 対象コンテンツを削除
      if req.TargetType() == "idol" {
          return s.idolRepo.Delete(ctx, req.TargetID())
      } else if req.TargetType() == "group" {
          return s.groupRepo.Delete(ctx, req.TargetID())
      }

      return nil
  }
  ```

##### 2.4.2 削除履歴の記録（2-3時間）
- [ ] DeletionLogモデル
  ```go
  type DeletionLog struct {
      ID            string
      TargetType    string
      TargetID      string
      Reason        string
      DeletedBy     string  // "system" or user_id
      DeletedAt     time.Time
      OriginalData  string  // JSON形式で元データを保存
  }
  ```

##### 2.4.3 通知機能（オプション、2-3時間）
- [ ] メール通知の実装
  - 申請者への削除完了通知
  - 投稿者への削除通知（将来的に）

#### 完了条件
- ✅ 承認された削除申請が自動で実行される
- ✅ 削除履歴が記録される
- ✅ ロールバック可能な設計

---

### Task 2.5: テストコードの追加

**優先度**: 🟡 Medium
**工数**: 3-4日
**担当者**: Backend
**期限**: Day 11-14

#### 実装内容

##### 2.5.1 ドメイン層のユニットテスト（1日）
- [ ] 値オブジェクトのテスト
  ```go
  // internal/domain/idol/idol_name_test.go
  func TestNewIdolName(t *testing.T) {
      tests := []struct {
          name    string
          input   string
          wantErr bool
      }{
          {"valid name", "山田花子", false},
          {"empty name", "", true},
          {"too long name", strings.Repeat("あ", 101), true},
      }
      // ...
  }
  ```

- [ ] エンティティのテスト
- [ ] ドメインサービスのテスト

##### 2.5.2 アプリケーション層のテスト（1日）
- [ ] モックリポジトリの作成
  ```go
  type MockIdolRepository struct {
      mock.Mock
  }

  func (m *MockIdolRepository) Save(ctx context.Context, idol *idol.Idol) error {
      args := m.Called(ctx, idol)
      return args.Error(0)
  }
  ```

- [ ] ApplicationServiceのテスト

##### 2.5.3 統合テスト（1日）
- [ ] MongoDBテストコンテナの設定
- [ ] リポジトリ層の統合テスト

##### 2.5.4 E2Eテスト（1日）
- [ ] APIエンドポイントのテスト
  ```go
  func TestCreateIdol(t *testing.T) {
      router := setupTestRouter()

      req := httptest.NewRequest("POST", "/api/v1/idols", strings.NewReader(`{
          "name": "Test Idol",
          "nationality": "日本"
      }`))
      req.Header.Set("Content-Type", "application/json")

      w := httptest.NewRecorder()
      router.ServeHTTP(w, req)

      assert.Equal(t, 201, w.Code)
  }
  ```

#### 完了条件
- ✅ テストカバレッジ80%以上
- ✅ CIで自動テスト実行
- ✅ テストドキュメント作成

---

## 📊 進捗管理

### マイルストーン
- **M1**: Week 1完了（Day 5） - セキュリティと法的保護の確立
- **M2**: Week 2前半完了（Day 9） - データ整合性の向上
- **M3**: Week 2完了（Day 14） - MVP完成

### リスク管理
| リスク | 影響度 | 対策 |
|--------|--------|------|
| モデレーション機能の複雑化 | 中 | 最小限の機能から開始、段階的拡張 |
| テスト実装の遅延 | 高 | 並行作業、優先度付け |
| データマイグレーションの失敗 | 高 | バックアップ必須、段階的移行 |

### レビューポイント
- Day 5: Week 1レビュー（法的保護の完成度確認）
- Day 9: データ整合性レビュー（アイドル-グループ関連）
- Day 14: MVP完成レビュー（全機能の動作確認）

---

## 🎯 タスク優先度マトリクス

### P0 (緊急 - 即座に対応)
- Task 1.1: 環境変数管理の改善

### P1 (高 - Week 1で完了)
- Task 1.2: 利用規約・プライバシーポリシー
- Task 1.3: モデレーション機能

### P2 (中 - Week 2で完了)
- Task 2.1: エラーハンドリング統一
- Task 2.2: バリデーション強化
- Task 2.3: アイドル-グループ関連付け
- Task 2.4: 削除申請自動処理
- Task 2.5: テストコード追加

---

## 📝 次のアクション

1. **即座**: Task 1.1（環境変数管理）から着手
2. **Day 1終了時**: 設定の確認、Day 2のタスク準備
3. **Week 1終了時**: 法的保護の完成度レビュー
4. **Week 2終了時**: MVP完成、次フェーズの計画策定

---

## 補足資料

### 参考リンク
- [プロバイダ責任制限法ガイドライン](https://www.soumu.go.jp/)
- [Go DDD実装パターン](https://github.com/bxcodec/go-clean-arch)
- [Gin Framework ドキュメント](https://gin-gonic.com/)

### 関連ドキュメント
- README.md - プロジェクト概要
- CLAUDE.md - 開発ガイドライン
- .env.example - 環境変数設定例
