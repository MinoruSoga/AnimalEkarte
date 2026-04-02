# FE-069: 一覧ページのフォントサイズ最低 text-base 化

**Status**: Open
**Priority**: High
**Affects**: VaccinationList, TrimmingList, InventoryList, EstimateList, Accounting, ReservationManagement, Dashboard
**Date Created**: 2026-03-18
**Related**: TASK-017, FE-067

## Summary

一覧ページ内でハードコードされた `text-xs` / `text-sm` を `text-base` に置換する。理想的には STYLE トークン（`STYLE.tableCell` 等）への統一も行う。

## 現状のコード

### VaccinationList.tsx

```typescript
// frontend/src/features/vaccinations/routes/VaccinationList.tsx
// :180 Button className="h-10 text-sm gap-2 bg-white"
// :216-220 全 TableCell に text-sm をハードコード
<TableCell className="font-mono text-sm text-[#37352F] py-2">{r.date}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
<TableCell className="text-sm font-medium text-[#37352F] py-2">{r.vaccineName}</TableCell>
<TableCell className="font-mono text-sm text-[#37352F] py-2">{r.nextDate}</TableCell>
```

### TrimmingList.tsx

```typescript
// frontend/src/features/trimming/routes/TrimmingList.tsx
// :56-71 全 TableCell に text-sm をハードコード
<TableCell className="font-mono text-sm text-[#37352F] py-2">{record.date}</TableCell>
<TableCell className="text-sm text-[#37352F] py-2">{record.ownerName}</TableCell>
// ... 他多数
```

### InventoryList.tsx

```typescript
// frontend/src/features/inventory/routes/InventoryList.tsx
// :220 Button className="h-10 text-sm gap-2 bg-white"
// :239 <div className="flex gap-4 text-sm">
// :276-291 全 TableCell に text-sm をハードコード
```

### EstimateList.tsx

```typescript
// frontend/src/features/estimates/routes/EstimateList.tsx
// :171-177 全 TableCell に text-sm をハードコード
```

### Accounting.tsx

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx
// :255-261 全 TableCell に text-sm をハードコード
```

### ReservationManagement.tsx

```typescript
// frontend/src/features/reservations/routes/ReservationManagement.tsx
// :142 className="h-10 px-4 text-sm font-medium"
// :164 SelectTrigger className="... h-10 text-sm"
// :181 SelectTrigger className="... h-10 text-sm"
// :203 <span className="text-xs text-[#37352F]/60"> ← text-xs
// :215 <p className="mt-2 text-[#37352F]/60 text-sm">
```

### Dashboard.tsx

```typescript
// frontend/src/features/dashboard/routes/Dashboard.tsx
// :303, 313, 321, 323, 338, 346 — text-sm をハードコード
```

## 必要な変更

### 方針

1. TableCell の `text-sm text-[#37352F] py-2` → `${STYLE.tableCell}` に統一（可能な箇所）
2. STYLE トークンに合致しないカスタム指定は `text-sm` → `text-base` に個別置換
3. `text-xs` → `text-base` に置換
4. Button/Select 等の `text-sm` → `text-base` に置換

### ファイル別変更

| ファイル | text-sm 箇所数 | text-xs 箇所数 |
|---------|---------------|---------------|
| VaccinationList.tsx | 6 | 0 |
| TrimmingList.tsx | 8 | 0 |
| InventoryList.tsx | 8 | 0 |
| EstimateList.tsx | 5 | 0 |
| Accounting.tsx | 5 | 0 |
| ReservationManagement.tsx | 5 | 1 |
| Dashboard.tsx | 6 | 0 |

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] 型は `models.ts` から導出
- [ ] STYLE トークン活用（可能な箇所）

## 依存関係

- FE-067 が先に完了していること（STYLE.tableHeaderCell 等が text-base になっている前提）

## 完了条件

- [ ] 上記 7 ファイル内に `text-xs` / `text-sm`（テキストサイズ）が残っていない
- [ ] 可能な箇所は `STYLE.tableCell` / `STYLE.tableCellMono` に統一済み
- [ ] `npm run build` パス
- [ ] `npm run lint` パス
- [ ] 各一覧ページのテーブルレイアウトが崩れていない
