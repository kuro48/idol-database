# API仕様書

## ベースURL
```
http://localhost:8081/api/v1
```

## 認証

### Phase 1（MVP）
書き込み系エンドポイント（POST/PUT/DELETE）には以下のヘッダーが必須：
```
X-API-Key: your-secret-api-key
```

### Phase 2-3（将来）
```
Authorization: Bearer <JWT-token>
```

---

## エンドポイント一覧

### 🔵 Idol（アイドル）API

#### 1. アイドル一覧取得
```http
GET /api/v1/idols
```

**クエリパラメータ:**
| パラメータ | 型 | 説明 | 例 |
|-----------|-----|------|-----|
| `name` | string | 名前での部分一致検索 | `?name=田中` |
| `group_id` | ObjectID | 特定グループに所属するアイドル | `?group_id=507f1f77bcf86cd799439011` |
| `is_active` | boolean | 活動中/卒業済みフィルター | `?is_active=true` |
| `min_age` | int | 最小年齢 | `?min_age=18` |
| `max_age` | int | 最大年齢 | `?max_age=25` |
| `blood_type` | string | 血液型 | `?blood_type=A` |
| `sort` | string | ソート項目 | `?sort=debut_date` |
| `order` | string | 昇順/降順 | `?order=desc` (asc/desc) |
| `page` | int | ページ番号（1始まり） | `?page=1` |
| `limit` | int | 1ページあたりの件数（デフォルト20、最大100） | `?limit=50` |

**レスポンス例:**
```json
{
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "name": "田中美咲",
      "name_kana": "たなかみさき",
      "birth_date": "2000-05-15T00:00:00Z",
      "birthplace": "東京都",
      "height": 165,
      "blood_type": "A",
      "debut_date": "2018-03-01T00:00:00Z",
      "graduation_date": null,
      "is_active": true,
      "group_memberships": [
        {
          "group_id": "507f191e810c19729de860ea",
          "group_name": "スターライト",
          "join_date": "2018-03-01T00:00:00Z",
          "leave_date": null,
          "role": "リーダー",
          "generation": 1
        }
      ],
      "profile_image_url": "https://example.com/images/tanaka.jpg",
      "official_url": "https://example.com/tanaka",
      "twitter_handle": "tanaka_misaki",
      "instagram_handle": "tanaka.misaki",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 100,
    "items_per_page": 20
  }
}
```

---

#### 2. アイドル詳細取得
```http
GET /api/v1/idols/:id
```

**パスパラメータ:**
- `id`: アイドルのObjectID

**レスポンス例:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "田中美咲",
  "name_kana": "たなかみさき",
  "birth_date": "2000-05-15T00:00:00Z",
  "birthplace": "東京都",
  "height": 165,
  "blood_type": "A",
  "debut_date": "2018-03-01T00:00:00Z",
  "graduation_date": null,
  "is_active": true,
  "group_memberships": [
    {
      "group_id": "507f191e810c19729de860ea",
      "group_name": "スターライト",
      "join_date": "2018-03-01T00:00:00Z",
      "leave_date": null,
      "role": "リーダー",
      "generation": 1
    }
  ],
  "profile_image_url": "https://example.com/images/tanaka.jpg",
  "official_url": "https://example.com/tanaka",
  "twitter_handle": "tanaka_misaki",
  "instagram_handle": "tanaka.misaki",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**エラーレスポンス:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Idol not found"
  }
}
```

---

#### 3. アイドル検索（全文検索）
```http
GET /api/v1/idols/search
```

**クエリパラメータ:**
| パラメータ | 型 | 説明 | 例 |
|-----------|-----|------|-----|
| `q` | string | 検索キーワード（名前、よみがな） | `?q=たなか` |
| `page` | int | ページ番号 | `?page=1` |
| `limit` | int | 1ページあたりの件数 | `?limit=20` |

**レスポンス:** アイドル一覧取得と同じ形式

---

#### 4. アイドル登録
```http
POST /api/v1/idols
```

**認証:** 必須（X-API-Key）

**リクエストボディ:**
```json
{
  "name": "田中美咲",
  "name_kana": "たなかみさき",
  "birth_date": "2000-05-15T00:00:00Z",
  "birthplace": "東京都",
  "height": 165,
  "blood_type": "A",
  "debut_date": "2018-03-01T00:00:00Z",
  "graduation_date": null,
  "group_memberships": [
    {
      "group_id": "507f191e810c19729de860ea",
      "join_date": "2018-03-01T00:00:00Z",
      "role": "リーダー",
      "generation": 1
    }
  ],
  "profile_image_url": "https://example.com/images/tanaka.jpg",
  "official_url": "https://example.com/tanaka",
  "twitter_handle": "tanaka_misaki",
  "instagram_handle": "tanaka.misaki"
}
```

**バリデーション:**
- `name`: 必須、1-100文字
- `name_kana`: 必須、1-100文字、ひらがなのみ
- `birth_date`: 必須、有効な日付
- `debut_date`: 必須、有効な日付
- `graduation_date`: オプション、debut_date より後
- `height`: オプション、50-300の範囲
- `blood_type`: オプション、"A", "B", "O", "AB"のいずれか
- `group_memberships.group_id`: グループが存在すること

**レスポンス（201 Created）:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "message": "Idol created successfully"
}
```

**エラーレスポンス（400 Bad Request）:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "name",
        "message": "Name is required"
      }
    ]
  }
}
```

---

#### 5. アイドル更新
```http
PUT /api/v1/idols/:id
```

**認証:** 必須（X-API-Key）

**リクエストボディ:** 登録と同じ形式（全フィールド必須）

**レスポンス（200 OK）:**
```json
{
  "message": "Idol updated successfully"
}
```

---

#### 6. アイドル部分更新
```http
PATCH /api/v1/idols/:id
```

**認証:** 必須（X-API-Key）

**リクエストボディ:** 更新したいフィールドのみ
```json
{
  "graduation_date": "2024-03-31T00:00:00Z"
}
```

**レスポンス（200 OK）:**
```json
{
  "message": "Idol updated successfully"
}
```

---

#### 7. アイドル削除
```http
DELETE /api/v1/idols/:id
```

**認証:** 必須（X-API-Key）

**レスポンス（200 OK）:**
```json
{
  "message": "Idol deleted successfully"
}
```

---

### 🟢 Group（グループ）API

#### 1. グループ一覧取得
```http
GET /api/v1/groups
```

**クエリパラメータ:**
| パラメータ | 型 | 説明 | 例 |
|-----------|-----|------|-----|
| `name` | string | 名前での部分一致検索 | `?name=スター` |
| `is_active` | boolean | 活動中/解散済みフィルター | `?is_active=true` |
| `agency` | string | 事務所名 | `?agency=ABC事務所` |
| `sort` | string | ソート項目 | `?sort=formation_date` |
| `order` | string | 昇順/降順 | `?order=desc` |
| `page` | int | ページ番号 | `?page=1` |
| `limit` | int | 1ページあたりの件数 | `?limit=20` |

**レスポンス例:**
```json
{
  "data": [
    {
      "id": "507f191e810c19729de860ea",
      "name": "スターライト",
      "name_kana": "すたーらいと",
      "formation_date": "2015-04-01T00:00:00Z",
      "disband_date": null,
      "is_active": true,
      "agency": "ABC事務所",
      "label": "XYZレコード",
      "logo_image_url": "https://example.com/logos/starlight.jpg",
      "official_url": "https://starlight-official.com",
      "twitter_handle": "starlight_official",
      "instagram_handle": "starlight.official",
      "youtube_channel": "UCxxxxxxxxxxxxx",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 3,
    "total_items": 50,
    "items_per_page": 20
  }
}
```

---

#### 2. グループ詳細取得
```http
GET /api/v1/groups/:id
```

**レスポンス:** グループ一覧取得の単一オブジェクト形式

---

#### 3. グループのメンバー一覧取得
```http
GET /api/v1/groups/:id/members
```

**クエリパラメータ:**
| パラメータ | 型 | 説明 | 例 |
|-----------|-----|------|-----|
| `is_active` | boolean | 現役/卒業メンバーフィルター | `?is_active=true` |
| `generation` | int | 期生でフィルター | `?generation=1` |

**レスポンス:**
```json
{
  "data": [
    {
      "idol_id": "507f1f77bcf86cd799439011",
      "name": "田中美咲",
      "name_kana": "たなかみさき",
      "join_date": "2018-03-01T00:00:00Z",
      "leave_date": null,
      "role": "リーダー",
      "generation": 1,
      "is_active_in_group": true
    }
  ]
}
```

---

#### 4. グループ登録
```http
POST /api/v1/groups
```

**認証:** 必須（X-API-Key）

**リクエストボディ:**
```json
{
  "name": "スターライト",
  "name_kana": "すたーらいと",
  "formation_date": "2015-04-01T00:00:00Z",
  "disband_date": null,
  "agency": "ABC事務所",
  "label": "XYZレコード",
  "logo_image_url": "https://example.com/logos/starlight.jpg",
  "official_url": "https://starlight-official.com",
  "twitter_handle": "starlight_official",
  "instagram_handle": "starlight.official",
  "youtube_channel": "UCxxxxxxxxxxxxx"
}
```

**バリデーション:**
- `name`: 必須、1-100文字
- `name_kana`: 必須、1-100文字、ひらがなのみ
- `formation_date`: 必須、有効な日付
- `disband_date`: オプション、formation_date より後

**レスポンス（201 Created）:**
```json
{
  "id": "507f191e810c19729de860ea",
  "message": "Group created successfully"
}
```

---

#### 5. グループ更新
```http
PUT /api/v1/groups/:id
```

**認証:** 必須（X-API-Key）

---

#### 6. グループ部分更新
```http
PATCH /api/v1/groups/:id
```

**認証:** 必須（X-API-Key）

---

#### 7. グループ削除
```http
DELETE /api/v1/groups/:id
```

**認証:** 必須（X-API-Key）

**注意:** グループに所属するアイドルの `group_memberships` から該当エントリを削除

---

### 🟡 Submission（申請）API（Phase 2で実装）

#### 1. 申請一覧取得（管理者のみ）
```http
GET /api/v1/submissions
```

**認証:** 必須（X-API-Key）

**クエリパラメータ:**
| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `status` | string | pending/approved/rejected |
| `type` | string | idol/group |
| `page` | int | ページ番号 |
| `limit` | int | 1ページあたりの件数 |

---

#### 2. 申請作成（誰でも可能）
```http
POST /api/v1/submissions
```

**認証:** 不要

**リクエストボディ:**
```json
{
  "type": "idol",
  "action": "create",
  "data": {
    "name": "田中美咲",
    "name_kana": "たなかみさき",
    ...
  },
  "submitter_email": "user@example.com",
  "submitter_name": "山田太郎",
  "source_url": "https://official-site.com/profile",
  "notes": "公式サイトより転載"
}
```

**レスポンス（201 Created）:**
```json
{
  "id": "507f1f77bcf86cd799439012",
  "message": "Submission created successfully. It will be reviewed by administrators."
}
```

---

#### 3. 申請承認（管理者のみ）
```http
POST /api/v1/submissions/:id/approve
```

**認証:** 必須（X-API-Key）

**リクエストボディ:**
```json
{
  "review_notes": "確認しました。公式情報と一致しています。"
}
```

**処理:**
- 申請内容をIdol/Groupコレクションに反映
- Submissionのstatusを"approved"に更新

---

#### 4. 申請却下（管理者のみ）
```http
POST /api/v1/submissions/:id/reject
```

**認証:** 必須（X-API-Key）

**リクエストボディ:**
```json
{
  "review_notes": "情報源が不明確なため却下"
}
```

---

### 📊 統計API（将来実装）

#### アイドル統計
```http
GET /api/v1/stats/idols
```

**レスポンス:**
```json
{
  "total_idols": 1500,
  "active_idols": 1200,
  "graduated_idols": 300,
  "average_age": 22.5,
  "blood_type_distribution": {
    "A": 450,
    "B": 300,
    "O": 400,
    "AB": 150
  }
}
```

---

## エラーレスポンス形式

### 標準エラーフォーマット
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": []
  }
}
```

### エラーコード一覧
| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| `VALIDATION_ERROR` | 400 | バリデーションエラー |
| `UNAUTHORIZED` | 401 | 認証エラー（API Key不正） |
| `FORBIDDEN` | 403 | 権限エラー |
| `NOT_FOUND` | 404 | リソースが見つからない |
| `DUPLICATE` | 409 | 重複エラー |
| `INTERNAL_ERROR` | 500 | サーバー内部エラー |

---

## レート制限

### Phase 1
- 読み取りAPI: 制限なし
- 書き込みAPI: 100リクエスト/時間

### Phase 2-3
- 読み取りAPI: 1000リクエスト/時間
- 書き込みAPI: 100リクエスト/時間（認証ユーザー）
- 申請API: 10申請/日（未認証ユーザー）

**レート制限超過時:**
```http
HTTP/1.1 429 Too Many Requests
Retry-After: 3600

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later."
  }
}
```

---

## CORS設定

### 開発環境
- すべてのオリジンを許可

### 本番環境
- 許可されたオリジンのみ（設定可能）
