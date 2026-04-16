# BUG-228: `bg-white` をデザイントークン `C.bgWhite` に未置換（vaccinations・examinations・inventory）

## 概要

`features/vaccinations/`、`features/examinations/`、`features/inventory/` の計 14 箇所で Tailwind クラス `bg-white` を直接使用している。`src/lib/design-tokens.ts` に `C.bgWhite = "bg-white"` トークンが定義されており、すべての色は必ずデザイントークン経由で指定するプロジェクト規約に違反している。

## 再現手順

（ランタイム動作は変わらないが、コードの一貫性違反として確認可能）

1. 各ファイルで `bg-white` を検索する
2. **結果**: `C.bgWhite` ではなく Tailwind クラスを直接記述している

## 期待する動作

- `bg-white` → `${C.bgWhite}` に統一

## 現状コード

### 対象ファイル一覧

```
features/vaccinations/components/VaccinationCard.tsx:28
features/vaccinations/routes/VaccinationList.tsx:233
features/vaccinations/routes/VaccinationForm.tsx:197
features/examinations/components/ExaminationCard.tsx:37
features/examinations/routes/ExaminationsList.tsx:250
features/examinations/routes/ExaminationForm.tsx:78,93,115,145,157
features/inventory/routes/InventoryForm.tsx:60,141,226
features/inventory/routes/InventoryList.tsx:253
```

#### `ExaminationCard.tsx:37`（代表例）
```tsx
// ❌ 現状: bg-white 直接使用
className={`bg-white border ${C.borderLight} shadow-none rounded-[4px] ...`}
```

#### `ExaminationForm.tsx:78`
```tsx
// ❌ 現状
<div className={`bg-white p-4 rounded-lg border ${C.borderMedium} space-y-4 shadow-sm`}>
```

### 比較: 正しい実装

```tsx
// ✅ 正しい: C.bgWhite トークンを使用
className={`${C.bgWhite} border ${C.borderLight} shadow-none rounded-[4px] ...`}
<div className={`${C.bgWhite} p-4 rounded-lg border ${C.borderMedium} space-y-4 shadow-sm`}>
```

### `design-tokens.ts` でのトークン定義

```typescript
// src/lib/design-tokens.ts
bgWhite: "bg-white",  // ← トークンとして定義済み
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `features/vaccinations/components/VaccinationCard.tsx:28` | `bg-white` 直接使用 | 未修正 |
| `features/vaccinations/routes/VaccinationList.tsx:233` | `bg-white` 直接使用 | 未修正 |
| `features/vaccinations/routes/VaccinationForm.tsx:197` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/components/ExaminationCard.tsx:37` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/routes/ExaminationsList.tsx:250` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/routes/ExaminationForm.tsx:78` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/routes/ExaminationForm.tsx:93` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/routes/ExaminationForm.tsx:115` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/routes/ExaminationForm.tsx:145` | `bg-white` 直接使用 | 未修正 |
| `features/examinations/routes/ExaminationForm.tsx:157` | `bg-white` 直接使用 | 未修正 |
| `features/inventory/routes/InventoryForm.tsx:60` | `bg-white` 直接使用 | 未修正 |
| `features/inventory/routes/InventoryForm.tsx:141` | `bg-white` 直接使用 | 未修正 |
| `features/inventory/routes/InventoryForm.tsx:226` | `bg-white` 直接使用 | 未修正 |
| `features/inventory/routes/InventoryList.tsx:253` | `bg-white` 直接使用 | 未修正 |

## 修正方針

各ファイルで `bg-white` を `${C.bgWhite}` に一括置換する。

```tsx
// Before
<div className={`bg-white p-4 rounded-lg border ${C.borderMedium} ...`}>
<Card className={`bg-white border ${C.borderLight} ...`}>

// After
<div className={`${C.bgWhite} p-4 rounded-lg border ${C.borderMedium} ...`}>
<Card className={`${C.bgWhite} border ${C.borderLight} ...`}>
```

注意: `focus:bg-white` や `hover:bg-white` のような疑似クラスは対象外（それらに対応するトークンは未定義）。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用（`#37352F`等ハードコード禁止）

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。

### プロジェクト内参照実装
- `src/lib/design-tokens.ts:239` — `bgWhite: "bg-white"` の定義確認

## 優先度
**Low** — 視覚的変化なし。デザイントークン一貫性の問題。同一ファイルで他の色は `C.*` トークンを使用している点との不統一を解消する。

## 関連チケット
- BUG-217: ReservationFormModal のハードコード色違反
- BUG-218: PetSelectionResultsTable の `text-gray-400` 直接使用
- BUG-219: StaffSettings の hex カラー直接使用
- BUG-220: WeekView の `decoration-red-500/50` 直接使用

## 関連ファイル
- `frontend/src/features/vaccinations/components/VaccinationCard.tsx`
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx`
- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`
- `frontend/src/features/examinations/components/ExaminationCard.tsx`
- `frontend/src/features/examinations/routes/ExaminationsList.tsx`
- `frontend/src/features/examinations/routes/ExaminationForm.tsx`
- `frontend/src/features/inventory/routes/InventoryForm.tsx`
- `frontend/src/features/inventory/routes/InventoryList.tsx`
- `frontend/src/lib/design-tokens.ts:239`
