# MongoDB でのメンバーシップ実装

## Collection設計

### 1. groups コレクション
```javascript
{
  "_id": ObjectId("..."),
  "name": "Sample Group",
  "formation_date": ISODate("2010-05-15T00:00:00Z"),
  "created_at": ISODate("2025-01-01T00:00:00Z"),
  "updated_at": ISODate("2025-01-01T00:00:00Z")
}
```

### 2. idols コレクション
```javascript
{
  "_id": ObjectId("..."),
  "name": "山田花子",
  "birthdate": ISODate("2000-05-15T00:00:00Z"),
  "nationality": "日本",
  "created_at": ISODate("2025-01-01T00:00:00Z"),
  "updated_at": ISODate("2025-01-01T00:00:00Z")
}
```

### 3. memberships コレクション（新規）
```javascript
{
  "_id": ObjectId("..."),
  "group_id": "675e1234...",
  "idol_id": "675e5678...",
  "joined_at": ISODate("2015-04-01T00:00:00Z"),
  "left_at": ISODate("2020-03-31T00:00:00Z"),  // nullなら現役
  "role": "member",
  "created_at": ISODate("2025-01-01T00:00:00Z"),
  "updated_at": ISODate("2025-01-01T00:00:00Z")
}
```

### インデックス設計

```javascript
// memberships コレクション
db.memberships.createIndex({ "group_id": 1, "left_at": 1 })  // 現役メンバー検索用
db.memberships.createIndex({ "idol_id": 1 })                  // アイドルの所属履歴検索用
db.memberships.createIndex({ "group_id": 1, "joined_at": 1 }) // 時系列検索用
```

---

## Go実装例

### MembershipRepository（MongoDB）

```go
package mongodb

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/kuro48/idol-api/internal/domain/group"
    "github.com/kuro48/idol-api/internal/domain/idol"
    "github.com/kuro48/idol-api/internal/domain/membership"
    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

type MembershipRepository struct {
    collection *mongo.Collection
}

func NewMembershipRepository(db *mongo.Database) *MembershipRepository {
    return &MembershipRepository{
        collection: db.Collection("memberships"),
    }
}

// membershipDocument はMongoDBに保存するドキュメント構造
type membershipDocument struct {
    ID        bson.ObjectID  `bson:"_id,omitempty"`
    GroupID   string         `bson:"group_id"`
    IdolID    string         `bson:"idol_id"`
    JoinedAt  time.Time      `bson:"joined_at"`
    LeftAt    *time.Time     `bson:"left_at,omitempty"`  // nullableにする
    Role      string         `bson:"role"`
    CreatedAt time.Time      `bson:"created_at"`
    UpdatedAt time.Time      `bson:"updated_at"`
}

// toMembershipDocument: ドメインモデル → MongoDB構造
func toMembershipDocument(m *membership.Membership) *membershipDocument {
    var objectID bson.ObjectID
    if m.ID().Value() != "" {
        objectID, _ = bson.ObjectIDFromHex(m.ID().Value())
    }

    var leftAt *time.Time
    if m.LeftAt() != nil {
        t := m.LeftAt().Value()
        leftAt = &t
    }

    return &membershipDocument{
        ID:        objectID,
        GroupID:   m.GroupID().Value(),
        IdolID:    m.IdolID().Value(),
        JoinedAt:  m.JoinedAt().Value(),
        LeftAt:    leftAt,
        Role:      string(m.Role()),
        CreatedAt: m.CreatedAt(),
        UpdatedAt: m.UpdatedAt(),
    }
}

// toMembershipDomain: MongoDB構造 → ドメインモデル
func toMembershipDomain(doc *membershipDocument) (*membership.Membership, error) {
    id, err := membership.NewMembershipID(doc.ID.Hex())
    if err != nil {
        return nil, err
    }

    groupID, err := group.NewGroupID(doc.GroupID)
    if err != nil {
        return nil, err
    }

    idolID, err := idol.NewIdolID(doc.IdolID)
    if err != nil {
        return nil, err
    }

    joinedAt, err := membership.NewJoinDate(
        doc.JoinedAt.Year(),
        int(doc.JoinedAt.Month()),
        doc.JoinedAt.Day(),
    )
    if err != nil {
        return nil, err
    }

    var leftAt *membership.LeaveDate
    if doc.LeftAt != nil {
        ld, err := membership.NewLeaveDate(
            doc.LeftAt.Year(),
            int(doc.LeftAt.Month()),
            doc.LeftAt.Day(),
        )
        if err != nil {
            return nil, err
        }
        leftAt = &ld
    }

    role, err := membership.NewMemberRole(doc.Role)
    if err != nil {
        return nil, err
    }

    return membership.Reconstruct(
        id,
        groupID,
        idolID,
        joinedAt,
        leftAt,
        role,
        doc.CreatedAt,
        doc.UpdatedAt,
    ), nil
}

// Save は新しいメンバーシップを保存する
func (r *MembershipRepository) Save(ctx context.Context, m *membership.Membership) error {
    doc := toMembershipDocument(m)

    if doc.ID.IsZero() {
        doc.ID = bson.NewObjectID()
        doc.CreatedAt = time.Now()
        doc.UpdatedAt = time.Now()
    }

    _, err := r.collection.InsertOne(ctx, doc)
    if err != nil {
        return fmt.Errorf("メンバーシップの保存エラー: %w", err)
    }

    return nil
}

// FindActiveByGroupID は特定グループの現役メンバーを取得
func (r *MembershipRepository) FindActiveByGroupID(
    ctx context.Context,
    groupID group.GroupID,
) ([]*membership.Membership, error) {
    // left_at が null のものを検索
    filter := bson.M{
        "group_id": groupID.Value(),
        "left_at":  nil,
    }

    cursor, err := r.collection.Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("現役メンバー取得エラー: %w", err)
    }
    defer cursor.Close(ctx)

    var docs []membershipDocument
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, fmt.Errorf("データ変換エラー: %w", err)
    }

    memberships := make([]*membership.Membership, 0, len(docs))
    for _, doc := range docs {
        m, err := toMembershipDomain(&doc)
        if err != nil {
            return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
        }
        memberships = append(memberships, m)
    }

    return memberships, nil
}

// CountActiveByGroupID は現役メンバー数を取得
func (r *MembershipRepository) CountActiveByGroupID(
    ctx context.Context,
    groupID group.GroupID,
) (int, error) {
    filter := bson.M{
        "group_id": groupID.Value(),
        "left_at":  nil,
    }

    count, err := r.collection.CountDocuments(ctx, filter)
    if err != nil {
        return 0, fmt.Errorf("メンバー数カウントエラー: %w", err)
    }

    return int(count), nil
}

// FindActiveByGroupIDAt は特定日時点での現役メンバーを取得
func (r *MembershipRepository) FindActiveByGroupIDAt(
    ctx context.Context,
    groupID group.GroupID,
    date time.Time,
) ([]*membership.Membership, error) {
    // joined_at <= date AND (left_at IS NULL OR left_at > date)
    filter := bson.M{
        "group_id":  groupID.Value(),
        "joined_at": bson.M{"$lte": date},
        "$or": []bson.M{
            {"left_at": nil},
            {"left_at": bson.M{"$gt": date}},
        },
    }

    cursor, err := r.collection.Find(ctx, filter)
    if err != nil {
        return nil, fmt.Errorf("特定日時点メンバー取得エラー: %w", err)
    }
    defer cursor.Close(ctx)

    var docs []membershipDocument
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, fmt.Errorf("データ変換エラー: %w", err)
    }

    memberships := make([]*membership.Membership, 0, len(docs))
    for _, doc := range docs {
        m, err := toMembershipDomain(&doc)
        if err != nil {
            return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
        }
        memberships = append(memberships, m)
    }

    return memberships, nil
}

// FindByIdolID はアイドルの所属履歴を取得
func (r *MembershipRepository) FindByIdolID(
    ctx context.Context,
    idolID idol.IdolID,
) ([]*membership.Membership, error) {
    filter := bson.M{
        "idol_id": idolID.Value(),
    }

    // 加入日でソート
    opts := options.Find().SetSort(bson.M{"joined_at": 1})

    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, fmt.Errorf("所属履歴取得エラー: %w", err)
    }
    defer cursor.Close(ctx)

    var docs []membershipDocument
    if err := cursor.All(ctx, &docs); err != nil {
        return nil, fmt.Errorf("データ変換エラー: %w", err)
    }

    memberships := make([]*membership.Membership, 0, len(docs))
    for _, doc := range docs {
        m, err := toMembershipDomain(&doc)
        if err != nil {
            return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
        }
        memberships = append(memberships, m)
    }

    return memberships, nil
}

// Update はメンバーシップを更新する（卒業処理など）
func (r *MembershipRepository) Update(ctx context.Context, m *membership.Membership) error {
    objectID, err := bson.ObjectIDFromHex(m.ID().Value())
    if err != nil {
        return fmt.Errorf("無効なID形式: %w", err)
    }

    doc := toMembershipDocument(m)
    doc.UpdatedAt = time.Now()

    updateDoc := bson.M{
        "$set": bson.M{
            "left_at":    doc.LeftAt,
            "role":       doc.Role,
            "updated_at": doc.UpdatedAt,
        },
    }

    result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateDoc)
    if err != nil {
        return fmt.Errorf("メンバーシップ更新エラー: %w", err)
    }

    if result.MatchedCount == 0 {
        return errors.New("メンバーシップが見つかりません")
    }

    return nil
}
```

---

## クエリパフォーマンス

### インデックスを使った最適化

```javascript
// 現役メンバー検索（高速）
db.memberships.find({ group_id: "...", left_at: null })
// → インデックス: { group_id: 1, left_at: 1 } を使用

// 特定日時点のメンバー（複雑だが最適化可能）
db.memberships.find({
  group_id: "...",
  joined_at: { $lte: ISODate("2020-01-01") },
  $or: [
    { left_at: null },
    { left_at: { $gt: ISODate("2020-01-01") } }
  ]
})
// → インデックス: { group_id: 1, joined_at: 1 } を使用
```

---

## $lookupを使った集約（オプション）

複数回クエリするのではなく、1回の集約クエリで全データ取得も可能：

```javascript
// グループと現役メンバーを一度に取得
db.groups.aggregate([
  { $match: { _id: ObjectId("...") } },
  {
    $lookup: {
      from: "memberships",
      let: { group_id: { $toString: "$_id" } },
      pipeline: [
        {
          $match: {
            $expr: {
              $and: [
                { $eq: ["$group_id", "$$group_id"] },
                { $eq: ["$left_at", null] }
              ]
            }
          }
        }
      ],
      as: "memberships"
    }
  },
  {
    $lookup: {
      from: "idols",
      localField: "memberships.idol_id",
      foreignField: "_id",
      as: "active_members"
    }
  }
])
```

ただし、**Go側で複数回クエリする方がシンプル**で保守性が高いです。

---

## まとめ

### MongoDBでも問題なく実装可能

**推奨設計**:
- 3つのCollection分離（groups, idols, memberships）
- 適切なインデックス設計
- Go側で複数クエリを組み合わせる

**メリット**:
- データ整合性が保たれる
- DDD設計と一致
- 柔軟なクエリが可能

**パフォーマンス**:
- インデックスを適切に設定すれば高速
- 必要に応じて$lookupで集約も可能
- アプリケーション層でのキャッシュも検討可能

### 正規化 vs 埋め込み の判断基準

**正規化（Collection分離）が良いケース**:
- データ更新が頻繁
- データ整合性が重要
- 複数の視点からクエリする（グループ中心、アイドル中心両方）

**埋め込みが良いケース**:
- ほぼ読み取り専用
- 常にグループ単位でしかアクセスしない
- 1回のクエリでの取得が最優先

**このプロジェクトでは正規化を推奨**します！
