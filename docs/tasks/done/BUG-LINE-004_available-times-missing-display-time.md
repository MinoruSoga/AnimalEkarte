# BUG-LINE-004: available-times レスポンスに display_time フィールドがない

## 概要

`GET /api/liff/:clinicId/available-times` のレスポンスに `display_time` フィールドが含まれていない。

## 詳細

**Frontend 型定義** (`frontend/line-reserve/src/types/models.ts:65-69`):
```typescript
export interface AvailableTime {
  start_time: string; // "HHMM"
  end_time: string;   // "HHMM"
  display_time: string; // "HH:MM"  ← 期待
}
```

**Backend 実際のレスポンス**:
```json
{"start_time": "0900", "end_time": "0915"}
// display_time フィールドなし
```

## 影響

**MEDIUM** — `TimeSelectPage.tsx:80` で `time.display_time || formatTime(time.start_time)` とフォールバック処理があるため、表示は壊れない。ただし型定義との乖離が存在する。

## 修正案

A. Backend の liff_response に `display_time` フィールドを追加（`"09:00"` 形式）
B. Frontend の型定義から `display_time` を削除し、`formatTime` のみに統一

案Bが軽量。表示ロジックは Frontend で完結すべき。

## 優先度

**LOW** — フォールバックあり、表示に影響なし
