# セキュリティ設計書

## フェーズ別セキュリティ戦略

### Phase 1: MVP（管理者のみ運用）

#### 認証方式: API Key

**実装方針:**
- 環境変数でAPI Keyを管理
- 書き込み系エンドポイントでAPI Key検証
- 読み取り系エンドポイントは認証不要

**API Key生成:**
```bash
# scripts/generate_api_key.sh
openssl rand -hex 32
```

**セキュリティ対策:**
1. **HTTPS必須**（本番環境）
2. **環境変数管理**（.envファイルは.gitignoreに追加）
3. **レート制限**（書き込みAPI: 100リクエスト/時間）
4. **入力バリデーション**（SQLインジェクション、XSS対策）
5. **CORS設定**（許可するオリジンを制限）

**リスク:**
- ✅ API Key漏洩時の影響が大きい → Key定期ローテーション
- ✅ 単一Keyのため権限分離不可 → Phase 2で解決

---

### Phase 2: 申請・承認フロー追加

#### 新機能: Submission API

**認証不要エンドポイント:**
- `POST /api/v1/submissions` - 誰でも申請可能

**セキュリティ対策:**
1. **レート制限強化**（未認証: 10申請/日）
2. **スパム対策**
   - reCAPTCHA統合（フロントエンド）
   - Email検証（申請時にメールアドレス必須）
3. **コンテンツフィルタリング**
   - 不適切な内容の検出
   - URLバリデーション（情報源の確認）
4. **申請データの一時保存**
   - 承認されるまで公開データに反映しない
   - 管理者による手動レビュー

**承認フロー:**
```
ユーザー申請
    ↓
申請データ保存（status: pending）
    ↓
管理者レビュー
    ↓
承認 → 公開データに反映
却下 → 申請者に通知（オプション）
```

---

### Phase 3: 信頼ユーザー拡大（JWT認証）

#### 認証方式: JWT + Role-Based Access Control (RBAC)

**ユーザー権限レベル:**
| Role | 権限 | 説明 |
|------|------|------|
| `admin` | すべての操作 | 管理者（申請の承認/却下、ユーザー管理） |
| `trusted_editor` | CRUD（即座に反映） | 信頼できるユーザー（直接編集可能） |
| `viewer` | 読み取りのみ | 一般ユーザー |

**JWT実装:**
```go
type JWTClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

func GenerateToken(user *model.Admin) (string, error) {
    claims := JWTClaims{
        UserID: user.ID.Hex(),
        Email:  user.Email,
        Role:   user.Role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            Issuer:    "idol-api",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(jwtSecret))
}
```

**認証エンドポイント:**
```
POST /api/v1/auth/register   - ユーザー登録（初期Role: viewer）
POST /api/v1/auth/login      - ログイン（JWT発行）
POST /api/v1/auth/refresh    - トークンリフレッシュ
POST /api/v1/auth/logout     - ログアウト
```

**セキュリティ対策:**
1. **パスワードハッシュ化**（bcrypt）
2. **JWT有効期限**（24時間）
3. **リフレッシュトークン**（7日間）
4. **トークンブラックリスト**（ログアウト時）
5. **2段階認証**（オプション、将来実装）

---

## 共通セキュリティ対策

### 1. 入力バリデーション

**実装箇所:** `internal/interface/validator/`

**検証項目:**
- 必須フィールドチェック
- データ型チェック
- 長さ制限（文字列、配列）
- 形式チェック（Email、URL、日付）
- 範囲チェック（数値）
- ホワイトリスト検証（血液型、ステータス等）

**サニタイゼーション:**
```go
import "github.com/microcosm-cc/bluemonday"

func SanitizeHTML(input string) string {
    p := bluemonday.StrictPolicy()
    return p.Sanitize(input)
}
```

---

### 2. レート制限

**実装方式:**

**Phase 1:**
- インメモリマップ（IP別カウンター）
- 簡易実装、再起動でリセット

**Phase 2-3:**
- Redis使用（永続化、分散対応）
- スライディングウィンドウ方式

**制限値:**
| エンドポイント | 認証状態 | 制限 |
|--------------|---------|-----|
| 読み取りAPI | なし | 1000リクエスト/時間 |
| 書き込みAPI | API Key | 100リクエスト/時間 |
| 申請API | なし | 10申請/日 |
| 認証API | なし | 5リクエスト/分（ログイン） |

**レート制限超過時の対応:**
```
HTTP/1.1 429 Too Many Requests
Retry-After: 3600

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "retry_after": 3600
  }
}
```

---

### 3. CORS設定

**開発環境:**
```go
config := cors.DefaultConfig()
config.AllowAllOrigins = true
```

**本番環境:**
```go
config := cors.Config{
    AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}
```

---

### 4. ロギングと監査

**ログレベル:**
- **DEBUG**: 開発環境のみ
- **INFO**: 通常の操作（API呼び出し、DB操作）
- **WARN**: 予期しない状況（バリデーションエラー、レート制限）
- **ERROR**: エラー（DB接続失敗、サーバーエラー）

**監査ログ:**
```go
type AuditLog struct {
    Timestamp   time.Time
    UserID      string
    Action      string  // "create", "update", "delete", "approve", "reject"
    ResourceType string // "idol", "group", "submission"
    ResourceID  string
    IPAddress   string
    UserAgent   string
    Changes     interface{} // 変更内容（JSON）
}
```

**ログ保存先:**
- 開発環境: 標準出力
- 本番環境: ファイル + クラウドロギングサービス（CloudWatch, Stackdriver等）

---

### 5. データ保護

#### 個人情報の扱い
- 生年月日、出身地などは公開情報として扱う
- 管理者情報（Email、パスワード）は暗号化

#### バックアップ戦略
- MongoDBの自動バックアップ（日次）
- ポイントインタイムリカバリー対応

#### データ削除ポリシー
- 論理削除（deleted_atフィールド追加）
- 物理削除は管理者が明示的に実行

---

### 6. HTTPS/TLS

**本番環境:**
- Let's Encrypt証明書使用
- TLS 1.2以上
- 強力な暗号スイートのみ許可

**HTTP → HTTPS リダイレクト:**
```go
r.Use(middleware.TLSRedirect())
```

---

### 7. 依存関係の脆弱性管理

**定期チェック:**
```bash
# Go依存関係の脆弱性スキャン
go list -json -m all | nancy sleuth
```

**自動更新:**
- Dependabot有効化（GitHub）
- 週次で依存関係の更新確認

---

## セキュリティチェックリスト

### Phase 1（MVP）
- [ ] API Key環境変数管理
- [ ] .envファイルを.gitignoreに追加
- [ ] HTTPS設定（本番環境）
- [ ] 入力バリデーション実装
- [ ] レート制限実装
- [ ] CORS設定
- [ ] エラーメッセージからの情報漏洩防止
- [ ] MongoDB接続文字列の環境変数化
- [ ] ログ実装

### Phase 2（申請フロー）
- [ ] 申請APIのレート制限強化
- [ ] reCAPTCHA統合
- [ ] スパムフィルタリング
- [ ] 不適切コンテンツ検出
- [ ] Email検証（オプション）

### Phase 3（JWT認証）
- [ ] パスワードハッシュ化（bcrypt）
- [ ] JWT実装（有効期限、リフレッシュトークン）
- [ ] RBAC実装
- [ ] トークンブラックリスト
- [ ] 2段階認証（オプション）
- [ ] 監査ログ強化

---

## インシデント対応計画

### API Key漏洩時
1. 即座に新しいAPI Keyを生成
2. .envファイルを更新
3. サーバー再起動
4. 影響範囲の調査（ログ確認）

### 不正アクセス検知時
1. 該当IPアドレスをブロック
2. レート制限を一時的に強化
3. ログ分析による影響範囲の特定
4. 必要に応じてデータのロールバック

### データ改ざん検知時
1. 影響を受けたデータの特定
2. バックアップからの復元
3. 改ざんの原因調査
4. 再発防止策の実施

---

## セキュリティベストプラクティス

1. **最小権限の原則**: 必要最小限の権限のみ付与
2. **防御の多層化**: 複数のセキュリティ対策を組み合わせる
3. **セキュアなデフォルト**: セキュアな設定をデフォルトとする
4. **情報の最小化**: エラーメッセージで内部情報を漏らさない
5. **定期的なレビュー**: セキュリティ設定の定期的な見直し
6. **依存関係の管理**: 脆弱性のある依存関係の迅速な更新
7. **ログと監視**: 異常検知のための適切なログと監視
