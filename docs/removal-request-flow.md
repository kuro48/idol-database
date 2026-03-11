# 削除申請対応フロー

## 概要

本サービスはユーザー投稿型のアイドル情報データベースです。肖像権・プライバシー権に配慮し、申請から24時間以内の対応を目標とします。

## 申請から完了までのフロー

```
申請者
  |
  | POST /api/v1/removal-requests
  |   { target_type, target_id, reason, requester_email }
  v
[status: pending]
  |
  | 運営者が確認
  | GET /api/v1/removal-requests/pending
  | Authorization: Bearer <ADMIN_API_KEY>
  |
  +-- 承認 --> PUT /api/v1/removal-requests/:id
  |            { status: "approved" }
  |            → 対象コンテンツを削除
  |
  +-- 却下 --> PUT /api/v1/removal-requests/:id
               { status: "rejected", rejection_reason: "..." }
               → 申請者へ理由を通知（メール等）
```

## 申請 API

### 削除申請作成
```bash
POST /api/v1/removal-requests
Content-Type: application/json

{
  "target_type": "idol",        # "idol" or "group"
  "target_id": "507f1f77...",   # 対象のID
  "reason": "本人です。削除を希望します",
  "requester_email": "contact@example.com"
}
```

### 申請確認（管理者）
```bash
GET /api/v1/removal-requests/pending
Authorization: Bearer <ADMIN_API_KEY>
```

### ステータス更新（管理者）
```bash
PUT /api/v1/removal-requests/:id
Authorization: Bearer <ADMIN_API_KEY>
Content-Type: application/json

{
  "status": "approved"
}
```

## 対応基準

| 理由 | 推奨対応 |
|---|---|
| 本人からの削除希望 | 原則承認（24時間以内） |
| 著作権侵害の申告 | 証拠確認後承認 |
| プライバシー侵害 | 緊急性に応じて優先対応 |
| 明らかに虚偽 | 却下し理由を記録 |

## 問い合わせ窓口

- **削除申請**: `POST /api/v1/removal-requests`（APIから直接申請）
- **その他**: GitHub Issues（https://github.com/kuro48/idol-database/issues）
