# BUG-LINE-003: 予約作成リクエストボディ course_id vs type_id

## 概要

LIFF App が予約作成時に `course_id` を送信するが、Backend は `type_id` を期待している。

## 該当コード

**Frontend** (`frontend/line-reserve/src/types/models.ts:102-110`):
```typescript
export interface CreateReservationBody {
  course_id: number;  // ← Frontend が送信
  ...
}
```

**Backend** (`backend/internal/handler/liff_request.go:7`):
```go
type liffCreateReservationRequest struct {
  TypeID uint64 `json:"type_id" binding:"required"`  // ← Backend が期待
  ...
}
```

## 影響

ステップ7（確認）→ 予約確定で `type_id` が 0 となり、`binding:"required"` バリデーションで 400 エラー。予約が作成されない。

## 修正案

Backend 側を `json:"course_id"` に変更する。

## 優先度

**CRITICAL** — 予約確定不能
