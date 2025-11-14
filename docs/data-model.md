# データモデル設計書

## MongoDB スキーマ設計

### 1. Idol（アイドル）コレクション

```go
type Idol struct {
    ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`

    // 基本情報
    Name            string              `json:"name" bson:"name" binding:"required"`
    NameKana        string              `json:"name_kana" bson:"name_kana" binding:"required"`
    BirthDate       time.Time           `json:"birth_date" bson:"birth_date"`
    Birthplace      string              `json:"birthplace" bson:"birthplace"`

    // 身体情報
    Height          *int                `json:"height,omitempty" bson:"height,omitempty"` // cm
    BloodType       string              `json:"blood_type,omitempty" bson:"blood_type,omitempty"` // A, B, O, AB

    // 活動期間
    DebutDate       time.Time           `json:"debut_date" bson:"debut_date"`
    GraduationDate  *time.Time          `json:"graduation_date,omitempty" bson:"graduation_date,omitempty"`
    IsActive        bool                `json:"is_active" bson:"is_active"` // 自動計算: graduation_date == nil

    // 所属グループ（複数対応）
    GroupMemberships []GroupMembership  `json:"group_memberships" bson:"group_memberships"`

    // メディア
    ProfileImageURL string              `json:"profile_image_url,omitempty" bson:"profile_image_url,omitempty"`

    // 外部リンク
    OfficialURL     string              `json:"official_url,omitempty" bson:"official_url,omitempty"`
    TwitterHandle   string              `json:"twitter_handle,omitempty" bson:"twitter_handle,omitempty"`
    InstagramHandle string              `json:"instagram_handle,omitempty" bson:"instagram_handle,omitempty"`

    // メタデータ
    CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
    CreatedBy       string              `json:"created_by" bson:"created_by"` // 作成者のユーザーID
}

// グループ所属情報（埋め込みドキュメント）
type GroupMembership struct {
    GroupID         primitive.ObjectID  `json:"group_id" bson:"group_id" binding:"required"`
    GroupName       string              `json:"group_name" bson:"group_name"` // キャッシュ用
    JoinDate        time.Time           `json:"join_date" bson:"join_date"`
    LeaveDate       *time.Time          `json:"leave_date,omitempty" bson:"leave_date,omitempty"`
    Role            string              `json:"role,omitempty" bson:"role,omitempty"` // リーダー、センター等
    Generation      *int                `json:"generation,omitempty" bson:"generation,omitempty"` // 期生
}
```

### 2. Group（グループ）コレクション

```go
type Group struct {
    ID              primitive.ObjectID  `json:"id" bson:"_id,omitempty"`

    // 基本情報
    Name            string              `json:"name" bson:"name" binding:"required"`
    NameKana        string              `json:"name_kana" bson:"name_kana" binding:"required"`

    // 活動情報
    FormationDate   time.Time           `json:"formation_date" bson:"formation_date"`
    DisbandDate     *time.Time          `json:"disband_date,omitempty" bson:"disband_date,omitempty"`
    IsActive        bool                `json:"is_active" bson:"is_active"`

    // 所属情報
    Agency          string              `json:"agency,omitempty" bson:"agency,omitempty"` // 事務所
    Label           string              `json:"label,omitempty" bson:"label,omitempty"`   // レーベル

    // メディア
    LogoImageURL    string              `json:"logo_image_url,omitempty" bson:"logo_image_url,omitempty"`

    // 外部リンク
    OfficialURL     string              `json:"official_url,omitempty" bson:"official_url,omitempty"`
    TwitterHandle   string              `json:"twitter_handle,omitempty" bson:"twitter_handle,omitempty"`
    InstagramHandle string              `json:"instagram_handle,omitempty" bson:"instagram_handle,omitempty"`
    YouTubeChannel  string              `json:"youtube_channel,omitempty" bson:"youtube_channel,omitempty"`

    // メタデータ
    CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
    CreatedBy       string              `json:"created_by" bson:"created_by"`
}
```

### 3. Submission（申請）コレクション（Phase 2で実装）

```go
type Submission struct {
    ID              primitive.ObjectID  `json:"id" bson:"_id,omitempty"`

    // 申請内容
    Type            string              `json:"type" bson:"type"` // "idol" or "group"
    Action          string              `json:"action" bson:"action"` // "create", "update", "delete"
    Data            interface{}         `json:"data" bson:"data"` // 申請されたIdolまたはGroup
    TargetID        *primitive.ObjectID `json:"target_id,omitempty" bson:"target_id,omitempty"` // update/deleteの場合

    // 申請者情報
    SubmitterEmail  string              `json:"submitter_email" bson:"submitter_email"`
    SubmitterName   string              `json:"submitter_name,omitempty" bson:"submitter_name,omitempty"`
    SourceURL       string              `json:"source_url,omitempty" bson:"source_url,omitempty"` // 情報源
    Notes           string              `json:"notes,omitempty" bson:"notes,omitempty"`

    // ステータス管理
    Status          string              `json:"status" bson:"status"` // "pending", "approved", "rejected"
    ReviewedBy      *string             `json:"reviewed_by,omitempty" bson:"reviewed_by,omitempty"`
    ReviewedAt      *time.Time          `json:"reviewed_at,omitempty" bson:"reviewed_at,omitempty"`
    ReviewNotes     string              `json:"review_notes,omitempty" bson:"review_notes,omitempty"`

    // メタデータ
    CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
}
```

### 4. Admin（管理者）コレクション（Phase 3で実装）

```go
type Admin struct {
    ID              primitive.ObjectID  `json:"id" bson:"_id,omitempty"`

    // 認証情報
    Email           string              `json:"email" bson:"email" binding:"required"`
    PasswordHash    string              `json:"-" bson:"password_hash"`

    // プロフィール
    Name            string              `json:"name" bson:"name"`

    // 権限
    Role            string              `json:"role" bson:"role"` // "admin", "trusted_editor", "viewer"
    Permissions     []string            `json:"permissions" bson:"permissions"`

    // API Key（Phase 1-2で使用）
    APIKey          string              `json:"-" bson:"api_key"`

    // メタデータ
    IsActive        bool                `json:"is_active" bson:"is_active"`
    CreatedAt       time.Time           `json:"created_at" bson:"created_at"`
    UpdatedAt       time.Time           `json:"updated_at" bson:"updated_at"`
    LastLoginAt     *time.Time          `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
}
```

## インデックス設計

### Idolコレクション
```javascript
db.idols.createIndex({ "name": 1 })
db.idols.createIndex({ "name_kana": 1 })
db.idols.createIndex({ "is_active": 1 })
db.idols.createIndex({ "group_memberships.group_id": 1 })
db.idols.createIndex({ "debut_date": -1 })
db.idols.createIndex({ "birth_date": 1 })

// テキスト検索用
db.idols.createIndex({
    "name": "text",
    "name_kana": "text"
})
```

### Groupコレクション
```javascript
db.groups.createIndex({ "name": 1 })
db.groups.createIndex({ "name_kana": 1 })
db.groups.createIndex({ "is_active": 1 })
db.groups.createIndex({ "formation_date": -1 })

// テキスト検索用
db.groups.createIndex({
    "name": "text",
    "name_kana": "text"
})
```

### Submissionコレクション
```javascript
db.submissions.createIndex({ "status": 1 })
db.submissions.createIndex({ "type": 1 })
db.submissions.createIndex({ "created_at": -1 })
db.submissions.createIndex({ "submitter_email": 1 })
```

### Adminコレクション
```javascript
db.admins.createIndex({ "email": 1 }, { unique: true })
db.admins.createIndex({ "api_key": 1 }, { unique: true, sparse: true })
```

## バリデーションルール

### 必須項目チェック
- **Idol**: name, name_kana, debut_date
- **Group**: name, name_kana, formation_date

### データ形式チェック
- 日付: RFC3339形式（例: "2020-01-15T00:00:00Z"）
- 血液型: "A", "B", "O", "AB"のいずれか
- URL: 有効なURL形式
- SNSハンドル: @を除いたユーザー名

### 重複チェック
- **Idol**: 同じ name + birth_date の組み合わせは警告
- **Group**: 同じ name は警告（同名グループが存在する可能性を考慮）

### ビジネスロジック検証
- `graduation_date` > `debut_date`
- `disband_date` > `formation_date`
- GroupMembership の `leave_date` > `join_date`
- `is_active` は `graduation_date` または `disband_date` から自動計算

## データの整合性

### アイドルとグループの関係性
- `GroupMembership.group_id` は `Group._id` を参照
- `GroupMembership.group_name` はキャッシュとして保持（検索高速化）
- グループ名変更時は関連するアイドルのキャッシュも更新

### カスケード処理
- グループ削除時: 関連するアイドルの `GroupMembership` から該当エントリを削除
- アイドル削除時: 特別な処理なし（グループ側には参照を持たない）
