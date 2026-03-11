# 検索API標準仕様

## 概要

本APIのすべての一覧・検索エンドポイントは、以下の統一クエリパラメータ仕様に従います。

---

## 標準クエリパラメータ

### フィルタ

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `name` | string | 名前による部分一致検索（大文字小文字無視） |
| ※リソース固有 | - | 各リソースのフィルタパラメータは下記参照 |

### ソート

| パラメータ | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `sort` | string | `created_at` | ソートフィールド（リソース別に有効な値が異なる） |
| `order` | string | `desc` | ソート順: `asc`（昇順）または `desc`（降順） |

### ページネーション

| パラメータ | 型 | デフォルト | 最大値 | 説明 |
|-----------|-----|-----------|--------|------|
| `page` | integer | `1` | - | ページ番号（1始まり） |
| `limit` | integer | `20` | `100` | 1ページあたりの件数 |

---

## レスポンス形式

### 一覧レスポンス

```json
{
  "data": [...],
  "meta": {
    "total": 150,
    "page": 2,
    "per_page": 20,
    "total_pages": 8
  },
  "links": {
    "self":  "https://api.example.com/api/v1/idols?page=2&limit=20",
    "first": "https://api.example.com/api/v1/idols?page=1&limit=20",
    "prev":  "https://api.example.com/api/v1/idols?page=1&limit=20",
    "next":  "https://api.example.com/api/v1/idols?page=3&limit=20",
    "last":  "https://api.example.com/api/v1/idols?page=8&limit=20"
  }
}
```

`links` はページネーションが必要な場合のみ含まれます。

---

## リソース別パラメータ一覧

### `/api/v1/idols`

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `name` | string | 名前部分一致 |
| `nationality` | string | 国コード完全一致（例: `JP`） |
| `agency_id` | string | 所属事務所ID |
| `group_id` | string | 所属グループID |
| `age_min` | integer | 最小年齢 |
| `age_max` | integer | 最大年齢 |
| `birthdate_from` | string (YYYY-MM-DD) | 誕生日範囲（開始） |
| `birthdate_to` | string (YYYY-MM-DD) | 誕生日範囲（終了） |
| `sort` | string | `name`, `birthdate`, `created_at` |

### `/api/v1/groups`

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `name` | string | 名前部分一致 |
| `sort` | string | `name`, `formation_date`, `created_at` |

### `/api/v1/agencies`

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `name` | string | 名前部分一致 |
| `country` | string | 国コード完全一致（例: `JP`） |
| `sort` | string | `name`, `founded_date`, `created_at` |

### `/api/v1/events`

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `event_type` | string | イベント種別 |
| `start_date_from` | string (YYYY-MM-DD) | 開催日範囲（開始） |
| `start_date_to` | string (YYYY-MM-DD) | 開催日範囲（終了） |
| `venue_id` | string | 会場ID |
| `performer_id` | string | パフォーマーID |
| `sort` | string | `start_date_time`, `created_at` |
| `order` | string | デフォルト: `asc`（イベントは昇順が自然） |

### `/api/v1/tags`

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `name` | string | 名前部分一致 |
| `category` | string | カテゴリ完全一致 |
| `sort` | string | `name`, `created_at` |

---

## エラーレスポンス

不正なクエリパラメータの場合:

```json
{
  "code": "BAD_REQUEST",
  "message": "sort パラメータは name, birthdate, created_at のいずれかである必要があります"
}
```

---

## クライアント実装ガイドライン

- 未知のフィールドがレスポンスに含まれる場合は無視してください（後方互換変更）
- `meta.total` は全件数です。ページの最後かどうかは `meta.page >= meta.total_pages` で判断できます
- `limit` パラメータは最大 100 です。大量取得が必要な場合はページネーションを繰り返してください
