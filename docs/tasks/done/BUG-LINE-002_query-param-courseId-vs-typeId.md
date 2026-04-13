# BUG-LINE-002: クエリパラメータ名不一致 courseId vs typeId（3エンドポイント）

## 概要

LIFF App が `courseId` クエリパラメータを送信しているが、Backend は `typeId` を期待している。3つのエンドポイントに影響。

## 影響エンドポイント

| エンドポイント | Frontend パラメータ | Backend パラメータ | 結果 |
|---|---|---|---|
| `GET /api/liff/:clinicId/staffs` | `courseId` | `typeId` | `invalid typeId` 400 エラー |
| `GET /api/liff/:clinicId/available-dates` | `courseId, staffId` | `typeId, staffId` | `invalid typeId` 400 エラー |
| `GET /api/liff/:clinicId/available-times` | `courseId, staffId, date` | `typeId, staffId, date` | `invalid typeId` 400 エラー |

## 該当コード

**Frontend** (`frontend/line-reserve/src/api/liff-api.ts`):
- L46: `params: { courseId }`
- L61: `params: { courseId, staffId }`
- L76: `params: { courseId, staffId, date }`

**Backend** (`backend/internal/handler/liff_handler.go`):
- L76: `c.Query("typeId")`
- L101: `c.Query("typeId")`
- L135: `c.Query("typeId")`

## 影響

ステップ3（スタッフ選択）、ステップ4（日付選択）、ステップ5（時間選択）がすべて失敗する。BUG-LINE-001 の修正後にブロックされる。

## 修正案

Backend 側を Frontend に合わせる（`typeId` → `courseId`）。LIFF の公開 API はエンドユーザー向けなので「コース」の用語が適切。

## 優先度

**CRITICAL** — 予約フロー ステップ3〜5 ブロック
