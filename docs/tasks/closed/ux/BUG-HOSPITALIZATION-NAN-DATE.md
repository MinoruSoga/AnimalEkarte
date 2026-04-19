# BUG-HOSPITALIZATION-NAN-DATE: 入院詳細デイリーカルテの日付が NaN 表示

## ステータス
✅ **修正済み** (2026-04-08)

## 再現手順
1. `/hospitalization/:id` を開く（退院済み・入院中いずれも）
2. デイリーカルテセクションの日付ナビゲーションを確認

## 症状
`NaN年NaN月NaN日（undefined）` と表示された。

## 根本原因
バックエンド API が `start_date` / `end_date` を ISO 8601 タイムスタンプ形式
（`"2026-03-10T00:00:00Z"`）で返却していたが、`transforms.ts` がそのまま
`startDate` / `endDate` に格納していた。

`DailyDateNav` は `YYYY-MM-DD` 形式の文字列を期待し、
`new Date(selectedDate + "T00:00:00")` と結合するため
`"2026-03-14T00:00:00ZT00:00:00"` → `Invalid Date` → NaN となっていた。

## 修正内容
`frontend/src/features/hospitalization/api/transforms.ts`:
```ts
// Before
startDate: hosp.start_date ?? "",
endDate: hosp.end_date ?? "",

// After
startDate: hosp.start_date ? hosp.start_date.split("T")[0] : "",
endDate: hosp.end_date ? hosp.end_date.split("T")[0] : "",
```

## 影響範囲
- `HospitalizationExpandedView` (PC表示)
- `HospitalizationTabbedView` (スマホ表示)
- `DailyRecordsTab` → `DailyDateNav`
