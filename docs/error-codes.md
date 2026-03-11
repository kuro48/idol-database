# エラーレスポンス仕様

## エラーレスポンス形式

すべてのAPIエラーは以下の統一フォーマットで返します：

```json
{
  "code": "ERROR_CODE",
  "message": "人間が読めるエラーメッセージ",
  "details": [...]
}
```

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `code` | string | ✅ | 機械可読なエラーコード（UPPER_SNAKE_CASE固定） |
| `message` | string | ✅ | 日本語エラーメッセージ |
| `details` | array/object | ❌ | バリデーションエラーの詳細情報など |

---

## エラーコード一覧

| code | HTTP Status | 説明 |
|------|------------|------|
| `BAD_REQUEST` | 400 | リクエストの形式が不正、またはパラメータが無効 |
| `VALIDATION_ERROR` | 400 | バリデーションエラー（`details` に詳細を含む） |
| `UNAUTHORIZED` | 401 | 認証が必要（Authorization ヘッダーなし、または形式不正） |
| `FORBIDDEN` | 403 | 権限がない（トークン不一致） |
| `NOT_FOUND` | 404 | 対象リソースが存在しない |
| `CONFLICT` | 409 | リソースが既に存在する（重複） |
| `TOO_MANY_REQUESTS` | 429 | レートリミット超過 |
| `INTERNAL_ERROR` | 500 | サーバー内部エラー |
| `SERVICE_UNAVAILABLE` | 503 | サービス設定不備（ADMIN_API_KEY 未設定など） |

---

## バリデーションエラーの `details` 形式

```json
{
  "code": "VALIDATION_ERROR",
  "message": "入力値が不正です",
  "details": [
    {
      "field": "name",
      "message": "nameは必須です"
    },
    {
      "field": "birthdate",
      "message": "birthdateはYYYY-MM-DD形式で入力してください"
    }
  ]
}
```

---

## エラー例

### 404 Not Found

```json
{
  "code": "NOT_FOUND",
  "message": "アイドルが見つかりません"
}
```

### 409 Conflict

```json
{
  "code": "CONFLICT",
  "message": "タグ名 'J-POP' は既に存在します"
}
```

### 401 Unauthorized

```json
{
  "code": "UNAUTHORIZED",
  "message": "認証が必要です"
}
```

### 403 Forbidden

```json
{
  "code": "FORBIDDEN",
  "message": "この操作を実行する権限がありません"
}
```

### 429 Too Many Requests

```json
{
  "code": "TOO_MANY_REQUESTS",
  "message": "リクエストが多すぎます。しばらく待ってから再試行してください"
}
```

---

## 実装ガイドライン

### Go ハンドラー内での使用方法

```go
// バリデーションエラー
c.JSON(http.StatusBadRequest, middleware.NewBadRequestError("リクエストが不正です: "+err.Error()))

// リソース不在（usecase のエラーを middleware.WriteError で変換推奨）
middleware.WriteError(c, err, middleware.ErrorContext{Resource: "アイドル"})

// 内部エラー
c.JSON(http.StatusInternalServerError, middleware.NewInternalError("処理中にエラーが発生しました"))
```

### 禁止事項

- `gin.H{"error": "..."}` のようなアドホックなエラーレスポンスは使用禁止
- コードを小文字（`"not_found"`）で返すことは禁止（UPPER_SNAKE_CASE固定）
- HTTPステータスコードのみに依存した実装は禁止（`code` フィールドを必ず含める）
