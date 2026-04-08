# BUG-164: ShiftCalendarPage — ローディング中に空のカレンダーグリッドが表示される

## 概要

`frontend/src/features/shifts/routes/ShiftCalendarPage.tsx` で、シフトデータ・スタッフデータの
取得中に `isLoading` チェックがなく、空の配列 `[]` を渡したまま `ShiftCalendarGrid` をレンダリングする。
「データなし」状態が一瞬フラッシュし、ユーザーが誤ってシフトが未登録と認識する可能性がある。

## 再現手順

1. シフトカレンダーページを開く
2. ネットワーク速度を低速にシミュレート（DevTools → Network → Slow 3G）
3. ページをリロード
4. **結果**: データ取得完了まで空のカレンダーグリッドが表示される

## 期待する動作

- `shiftsQuery.isLoading || staffsQuery.isLoading` が true の間は `<LoadingFallback />` を表示する
- `shiftsQuery.isError || staffsQuery.isError` の場合は `<ErrorFallback />` を表示する

## 現状コード

### `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx:48-70付近`
```tsx
// Before: isLoading チェックなし
const shifts = shiftsQuery.data ?? [];  // ← ローディング中は []
const staffs = staffsQuery.data ?? [];  // ← ローディング中は []

// shiftsQuery.isError のみ確認（staffsQuery.isError 未確認）
{shiftsQuery.isError ? (
  <div>シフトデータの取得に失敗しました</div>
) : (
  <ShiftCalendarGrid
    shifts={shifts}   // ← 空配列のまま表示される
    staffs={staffs}   // ← 空配列のまま表示される
  />
)}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:196-197
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `ShiftCalendarPage.tsx:48-49` | shiftsQuery.isLoading 未チェック | 未修正 |
| `ShiftCalendarPage.tsx:48-49` | staffsQuery.isLoading 未チェック | 未修正 |
| `ShiftCalendarPage.tsx:53-70` | staffsQuery.isError 未チェック | 未修正 |

## 修正方針

### 1. ローディング・エラー状態の追加 — `ShiftCalendarPage.tsx:48付近`
```tsx
import { LoadingFallback } from "@/components/shared/DataStates/LoadingFallback";
import { ErrorFallback } from "@/components/shared/DataStates/ErrorFallback";

const isLoading = shiftsQuery.isLoading || staffsQuery.isLoading;
const isError = shiftsQuery.isError || staffsQuery.isError;
const shifts = shiftsQuery.data ?? [];
const staffs = staffsQuery.data ?? [];

// JSX内で早期リターン代わりに条件分岐
if (isLoading) return <LoadingFallback />;
if (isError) return <ErrorFallback />;
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Frontend ベストプラクティス
データ取得中は `LoadingFallback`、エラー時は `ErrorFallback` を使用するのがプロジェクト標準。
`features/medical-records/routes/MedicalRecords.tsx` が参照実装。

## 優先度
**Medium** — UX上の問題。シフトデータなしと誤認されるリスクがある。

## 関連チケット
- FE-247: 受付カンバンの初期ローディングスケルトン欠如（同種問題）
- BUG-163: MedicalRecordForm の null リターン問題（同種）

## 関連ファイル
- `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx:48-70`
- `frontend/src/components/shared/DataStates/LoadingFallback.tsx`
- `frontend/src/components/shared/DataStates/ErrorFallback.tsx`
