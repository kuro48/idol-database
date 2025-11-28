# グループとアイドルの関係性設計

## 問題定義

- アイドルグループには「現在のメンバー」と「過去のメンバー（卒業生）」がいる
- 「ある時点でのメンバー数」を知りたい
- グループエンティティに全メンバーを持たせると、現役メンバー数の取得が複雑になる

## 解決策: Membershipドメインの導入

グループとアイドルの**所属関係**を独立したドメインとして扱います。

---

## データ構造

### 1. Group（グループ）

```go
// グループはグループ自体の情報のみを持つ
type Group struct {
    id            GroupID
    name          GroupName
    formationDate FormationDate
    // メンバー情報は持たない！
    createdAt     time.Time
    updatedAt     time.Time
}
```

### 2. Idol（アイドル）

```go
// アイドルはアイドル自体の情報のみを持つ
type Idol struct {
    id          IdolID
    name        IdolName
    birthdate   Birthdate
    nationality Nationality
    // グループ情報は持たない！
    createdAt   time.Time
    updatedAt   time.Time
}
```

### 3. Membership（所属関係）← 新しいドメイン

```go
package membership

import (
    "time"
    "github.com/kuro48/idol-api/internal/domain/group"
    "github.com/kuro48/idol-api/internal/domain/idol"
)

// Membership はアイドルのグループ所属関係を表す
type Membership struct {
    id        MembershipID
    groupID   group.GroupID
    idolID    idol.IdolID
    joinedAt  JoinDate      // 加入日
    leftAt    *LeaveDate    // 卒業日（nullなら現役）
    role      MemberRole    // リーダー、メンバーなど
    createdAt time.Time
    updatedAt time.Time
}

// NewMembership は新しい所属関係を作成する（加入時）
func NewMembership(
    groupID group.GroupID,
    idolID idol.IdolID,
    joinedAt JoinDate,
    role MemberRole,
) *Membership {
    now := time.Now()
    return &Membership{
        groupID:   groupID,
        idolID:    idolID,
        joinedAt:  joinedAt,
        leftAt:    nil, // 加入時は現役
        role:      role,
        createdAt: now,
        updatedAt: now,
    }
}

// Leave はメンバーが卒業する
func (m *Membership) Leave(leftAt LeaveDate) error {
    // ビジネスルール: 既に卒業している場合はエラー
    if m.leftAt != nil {
        return NewDomainError("既に卒業済みです")
    }

    // ビジネスルール: 卒業日は加入日より後
    if leftAt.Value().Before(m.joinedAt.Value()) {
        return NewDomainError("卒業日は加入日より後である必要があります")
    }

    m.leftAt = &leftAt
    m.updatedAt = time.Now()
    return nil
}

// IsActive は現在活動中かチェック
func (m *Membership) IsActive() bool {
    return m.leftAt == nil
}

// IsActiveAt は指定日時点で活動中だったかチェック
func (m *Membership) IsActiveAt(date time.Time) bool {
    // 加入日より前
    if date.Before(m.joinedAt.Value()) {
        return false
    }

    // 卒業していない、または卒業日より前
    if m.leftAt == nil {
        return true
    }

    return date.Before(m.leftAt.Value())
}
```

### 値オブジェクト

```go
// MemberRole はメンバーの役割
type MemberRole string

const (
    RoleLeader      MemberRole = "leader"       // リーダー
    RoleSubLeader   MemberRole = "sub_leader"   // サブリーダー
    RoleMember      MemberRole = "member"       // 一般メンバー
    RoleTrainee     MemberRole = "trainee"      // 研修生
)

// JoinDate は加入日
type JoinDate struct {
    value time.Time
}

func NewJoinDate(year, month, day int) (JoinDate, error) {
    date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

    if date.After(time.Now()) {
        return JoinDate{}, errors.New("加入日は未来の日付にできません")
    }

    return JoinDate{value: date}, nil
}

func (d JoinDate) Value() time.Time {
    return d.value
}

// LeaveDate は卒業日
type LeaveDate struct {
    value time.Time
}

func NewLeaveDate(year, month, day int) (LeaveDate, error) {
    date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

    if date.After(time.Now()) {
        return LeaveDate{}, errors.New("卒業日は未来の日付にできません")
    }

    return LeaveDate{value: date}, nil
}

func (d LeaveDate) Value() time.Time {
    return d.value
}
```

---

## リポジトリインターフェース

```go
package membership

import (
    "context"
    "time"
    "github.com/kuro48/idol-api/internal/domain/group"
    "github.com/kuro48/idol-api/internal/domain/idol"
)

type Repository interface {
    // 基本CRUD
    Save(ctx context.Context, membership *Membership) error
    FindByID(ctx context.Context, id MembershipID) (*Membership, error)
    Update(ctx context.Context, membership *Membership) error
    Delete(ctx context.Context, id MembershipID) error

    // カスタムクエリ
    // 特定グループの現役メンバーを取得
    FindActiveByGroupID(ctx context.Context, groupID group.GroupID) ([]*Membership, error)

    // 特定グループの全メンバー（卒業生含む）を取得
    FindAllByGroupID(ctx context.Context, groupID group.GroupID) ([]*Membership, error)

    // 特定アイドルの所属履歴を取得
    FindByIdolID(ctx context.Context, idolID idol.IdolID) ([]*Membership, error)

    // 特定日時点での特定グループのメンバーを取得
    FindActiveByGroupIDAt(ctx context.Context, groupID group.GroupID, date time.Time) ([]*Membership, error)

    // 現役メンバー数を取得
    CountActiveByGroupID(ctx context.Context, groupID group.GroupID) (int, error)
}
```

---

## 使用例

### 1. アイドルがグループに加入

```go
// アプリケーションサービスで実装
func (s *MembershipService) JoinGroup(
    ctx context.Context,
    groupID string,
    idolID string,
    joinedDate time.Time,
    role string,
) error {
    // 1. ドメインオブジェクトに変換
    gid, _ := group.NewGroupID(groupID)
    iid, _ := idol.NewIdolID(idolID)
    joined, _ := membership.NewJoinDate(joinedDate.Year(), int(joinedDate.Month()), joinedDate.Day())
    memberRole, _ := membership.NewMemberRole(role)

    // 2. Membershipエンティティ作成
    m := membership.NewMembership(gid, iid, joined, memberRole)

    // 3. 保存
    return s.membershipRepo.Save(ctx, m)
}
```

### 2. 現在のメンバー数を取得

```go
func (s *GroupService) GetCurrentMemberCount(
    ctx context.Context,
    groupID string,
) (int, error) {
    gid, _ := group.NewGroupID(groupID)

    // Membershipリポジトリで現役メンバー数を取得
    count, err := s.membershipRepo.CountActiveByGroupID(ctx, gid)
    return count, err
}
```

### 3. 2020年時点のメンバー数を取得

```go
func (s *GroupService) GetMemberCountAt(
    ctx context.Context,
    groupID string,
    date time.Time,
) (int, error) {
    gid, _ := group.NewGroupID(groupID)

    // 特定日時点でのメンバーシップを取得
    memberships, err := s.membershipRepo.FindActiveByGroupIDAt(ctx, gid, date)
    if err != nil {
        return 0, err
    }

    return len(memberships), nil
}
```

### 4. グループの現役メンバー一覧を取得

```go
// DTO定義
type GroupWithMembersDTO struct {
    Group        GroupDTO        `json:"group"`
    ActiveIdols  []IdolDTO       `json:"active_members"`
    MemberCount  int             `json:"member_count"`
}

// アプリケーションサービス
func (s *GroupService) GetGroupWithMembers(
    ctx context.Context,
    groupID string,
) (*GroupWithMembersDTO, error) {
    // 1. グループ取得
    gid, _ := group.NewGroupID(groupID)
    grp, err := s.groupRepo.FindByID(ctx, gid)
    if err != nil {
        return nil, err
    }

    // 2. 現役メンバーシップ取得
    memberships, err := s.membershipRepo.FindActiveByGroupID(ctx, gid)
    if err != nil {
        return nil, err
    }

    // 3. 各メンバーシップからアイドル情報を取得
    idols := make([]IdolDTO, 0, len(memberships))
    for _, m := range memberships {
        idol, err := s.idolRepo.FindByID(ctx, m.IdolID())
        if err != nil {
            continue
        }
        idols = append(idols, toIdolDTO(idol))
    }

    return &GroupWithMembersDTO{
        Group:       toGroupDTO(grp),
        ActiveIdols: idols,
        MemberCount: len(idols),
    }, nil
}
```

### 5. アイドルの所属履歴を取得

```go
type IdolWithGroupHistoryDTO struct {
    Idol   IdolDTO              `json:"idol"`
    Groups []GroupHistoryDTO    `json:"group_history"`
}

type GroupHistoryDTO struct {
    Group     GroupDTO   `json:"group"`
    JoinedAt  time.Time  `json:"joined_at"`
    LeftAt    *time.Time `json:"left_at,omitempty"`
    Role      string     `json:"role"`
    IsActive  bool       `json:"is_active"`
}

func (s *IdolService) GetIdolWithGroupHistory(
    ctx context.Context,
    idolID string,
) (*IdolWithGroupHistoryDTO, error) {
    // 1. アイドル取得
    iid, _ := idol.NewIdolID(idolID)
    idl, err := s.idolRepo.FindByID(ctx, iid)
    if err != nil {
        return nil, err
    }

    // 2. 所属履歴取得
    memberships, err := s.membershipRepo.FindByIdolID(ctx, iid)
    if err != nil {
        return nil, err
    }

    // 3. 各グループ情報を取得
    history := make([]GroupHistoryDTO, 0, len(memberships))
    for _, m := range memberships {
        grp, err := s.groupRepo.FindByID(ctx, m.GroupID())
        if err != nil {
            continue
        }

        var leftAt *time.Time
        if !m.IsActive() {
            t := m.LeftAt().Value()
            leftAt = &t
        }

        history = append(history, GroupHistoryDTO{
            Group:    toGroupDTO(grp),
            JoinedAt: m.JoinedAt().Value(),
            LeftAt:   leftAt,
            Role:     string(m.Role()),
            IsActive: m.IsActive(),
        })
    }

    return &IdolWithGroupHistoryDTO{
        Idol:   toIdolDTO(idl),
        Groups: history,
    }, nil
}
```

---

## データベース構造（MongoDB）

### groups コレクション

```json
{
  "_id": ObjectId("..."),
  "name": "Sample Group",
  "formation_date": ISODate("2010-05-15T00:00:00Z"),
  "created_at": ISODate("2025-01-01T00:00:00Z"),
  "updated_at": ISODate("2025-01-01T00:00:00Z")
}
```

### idols コレクション

```json
{
  "_id": ObjectId("..."),
  "name": "山田花子",
  "birthdate": ISODate("2000-05-15T00:00:00Z"),
  "nationality": "日本",
  "created_at": ISODate("2025-01-01T00:00:00Z"),
  "updated_at": ISODate("2025-01-01T00:00:00Z")
}
```

### memberships コレクション（新規）

```json
{
  "_id": ObjectId("..."),
  "group_id": "675e1234...",  // GroupのObjectID
  "idol_id": "675e5678...",   // IdolのObjectID
  "joined_at": ISODate("2015-04-01T00:00:00Z"),
  "left_at": ISODate("2020-03-31T00:00:00Z"),  // nullなら現役
  "role": "member",
  "created_at": ISODate("2025-01-01T00:00:00Z"),
  "updated_at": ISODate("2025-01-01T00:00:00Z")
}
```

### クエリ例

```javascript
// 現役メンバー数を取得
db.memberships.countDocuments({
  group_id: "675e1234...",
  left_at: null
})

// 2020年1月1日時点のメンバーを取得
db.memberships.find({
  group_id: "675e1234...",
  joined_at: { $lte: ISODate("2020-01-01T00:00:00Z") },
  $or: [
    { left_at: null },
    { left_at: { $gte: ISODate("2020-01-01T00:00:00Z") } }
  ]
})

// アイドルの所属履歴を取得
db.memberships.find({
  idol_id: "675e5678..."
}).sort({ joined_at: 1 })
```

---

## メリット

### 1. 明確な責務分離
- `Group`: グループ自体の情報
- `Idol`: アイドル自体の情報
- `Membership`: 所属関係と時系列情報

### 2. 柔軟なクエリ
- 現在のメンバー数
- 過去のある時点のメンバー数
- メンバーの所属履歴
- 卒業生を含む全メンバー

### 3. ビジネスルールの明確化
- 卒業処理のロジックが`Membership`に集約
- 加入・卒業の妥当性チェックが容易

### 4. スケーラビリティ
- 大規模グループでもパフォーマンス維持
- インデックス最適化が容易

---

## API エンドポイント例

```bash
# グループの現役メンバー一覧
GET /api/v1/groups/:id/members

# グループの全メンバー（卒業生含む）
GET /api/v1/groups/:id/members?include_graduated=true

# グループの特定時点のメンバー
GET /api/v1/groups/:id/members?date=2020-01-01

# アイドルの所属履歴
GET /api/v1/idols/:id/group-history

# アイドルをグループに加入させる
POST /api/v1/memberships
{
  "group_id": "...",
  "idol_id": "...",
  "joined_at": "2015-04-01",
  "role": "member"
}

# アイドルを卒業させる
PUT /api/v1/memberships/:id/leave
{
  "left_at": "2020-03-31"
}
```

---

## まとめ

**推奨アプローチ**:
- `Group`、`Idol`、`Membership`の3つのドメインに分離
- 所属関係を独立したエンティティとして扱う
- 時系列データ（加入日・卒業日）を`Membership`で管理

**この設計のメリット**:
- 現役・卒業の区別が明確
- 時系列クエリが容易
- ビジネスロジックが集約される
- 将来の拡張（複数グループ兼任など）に対応しやすい
