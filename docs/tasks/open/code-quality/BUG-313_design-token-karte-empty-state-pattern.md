# BUG-313: STYLE.tableEmptySm トークン欠如 — カルテ3タブで同一パターン重複

## 概要
カルテ（電子カルテ）の検査・処置・バイタルの3タブにおいて、空状態表示のクラス文字列 `text-center py-12 text-sm ${C.text40}` が同一パターンで繰り返されているが、対応するデザイントークンが存在しない。`STYLE.tableEmpty`（`text-base C.text70`）とは意図的にサイズ・色が異なるため、新トークンの追加が必要。

## 現状コード

### `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx:285`
```tsx
<td colSpan={6} className={`text-center py-12 text-sm ${C.text40}`}>
  データがありません
</td>
```

### `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:270`
```tsx
className={`text-center py-12 text-sm ${C.text40}`}
```

### `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:461`
```tsx
<td colSpan={7} className={`text-center py-12 text-sm ${C.text40}`}>
  データがありません
</td>
```

### 比較: 既存の `STYLE.tableEmpty`
```ts
// frontend/src/lib/design-tokens.ts
tableEmpty: `text-center py-12 ${C.text70} text-base`,
```

**差分**: カルテ内タブは `text-sm`（12px相当、コンパクト）かつ `C.text40`（より薄い）。`STYLE.tableEmpty`（`text-base C.text70`）とは明確に異なる意図的スタイル。

## 影響範囲

| 対象 | ファイル | 行 | 状態 |
|------|---------|-----|------|
| 検査タブ空状態 | `features/medical-records/components/CheckupsTab/CheckupsTab.tsx` | 285 | 未修正 |
| 処置タブ空状態 | `features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx` | 270 | 未修正 |
| バイタルタブ空状態 | `features/medical-records/components/VitalsTab/VitalsTab.tsx` | 461 | 未修正 |

## 修正方針

### 1. `frontend/src/lib/design-tokens.ts` にトークン追加

```ts
// STYLE オブジェクト内 tableEmpty の直下に追加
tableEmpty:
  `text-center py-12 ${C.text70} text-base`,
tableEmptySm:                          // ← 追加
  `text-center py-12 ${C.text40} text-sm`,
```

### 2. 3ファイルを一斉置換

```tsx
// Before
className={`text-center py-12 text-sm ${C.text40}`}

// After
className={STYLE.tableEmptySm}
```

各ファイルの import に `STYLE` が含まれていることを確認（CheckupsTab は `C` のみの場合あり）。

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 同一パターンが複数箇所に重複する場合はトークン化すべきだ。

## 優先度
**Low** — 機能に影響なし。コード一貫性の改善。3箇所同時に修正可能。

## 関連チケット
- なし（このスキャンで初めて検出）

## 関連ファイル
- `frontend/src/lib/design-tokens.ts` — トークン追加先
- `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx:285`
- `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:270`
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:461`
