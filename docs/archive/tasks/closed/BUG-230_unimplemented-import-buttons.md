# BUG-230: 未実装データ取込ボタンが `onClick={() => {}}` プレースホルダのまま残存

## 概要

`vaccinations/routes/VaccinationList.tsx:233`、`examinations/routes/ExaminationsList.tsx:250`、`inventory/routes/InventoryList.tsx:254` に「データ取込」ボタンが存在するが、`onClick={() => {}}` のプレースホルダ実装のまま公開されている。ボタンを押しても何も起きず、ユーザーに混乱を与える。また手動スタイリング（`h-10 text-base gap-2 bg-white`）が各ファイルで微妙に異なり（`text-base` vs `text-sm`）、視覚的一貫性も欠く。

## 再現手順

1. ワクチン一覧（`/vaccinations`）を開く
2. ヘッダー右上の「データ取込」ボタンを押す
3. **結果**: 何も起きない（`onClick={() => {}}` のため）
4. 診察一覧（`/examinations`）・在庫管理（`/inventory`）でも同様

## 期待する動作

- ボタンを押すとデータ取込フローが開始される、または
- 実装されるまでボタン自体を非表示にする（未実装 UI を公開しない）

## 現状コード

### `features/vaccinations/routes/VaccinationList.tsx:233`
```tsx
<Button variant="outline" className="h-10 text-base gap-2 bg-white" onClick={() => {}}>
  <FileSpreadsheet className={ICON.action} />
  データ取込
</Button>
```

### `features/examinations/routes/ExaminationsList.tsx:250`
```tsx
<Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
  <FileSpreadsheet className={ICON.action} />
  検査データ取込
</Button>
```

### `features/inventory/routes/InventoryList.tsx:254`
```tsx
<Button
  variant="outline"
  className="h-10 text-base gap-2 bg-white"
  onClick={() => {}}
>
  <FileSpreadsheet className={ICON.action} />
  データ取込
</Button>
```

### 問題点
1. **`onClick={() => {}}` — 何もしない**: ユーザーが押しても無反応。壊れているように見える。
2. **手動スタイリングの不統一**: `text-base` vs `text-sm` の差異。`bg-white` はデザイントークン非使用（`C.bgPage` または `C.bgWhite` を使うべき）。

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `features/vaccinations/routes/VaccinationList.tsx:233` | `onClick={() => {}}` プレースホルダ | 未修正 |
| `features/examinations/routes/ExaminationsList.tsx:250` | `onClick={() => {}}` プレースホルダ・`text-sm` 不統一 | 未修正 |
| `features/inventory/routes/InventoryList.tsx:254` | `onClick={() => {}}` プレースホルダ | 未修正 |

## 修正方針

### 短期対応（推奨）: ボタンを非表示にする

実装が確定するまで、未実装のボタンを削除または非表示にする。未実装 UI を公開するより、機能がないことを明示するほうが UX 上正しい。

```tsx
// VaccinationList.tsx:233 — ボタン削除
// ExaminationsList.tsx:250 — ボタン削除
// InventoryList.tsx:254 — ボタン削除
```

### 長期対応: 実装する場合のスタイリング修正

実装する場合は、手動スタイリングを統一し `bg-white` を `C.bgPage` に置換する：

```tsx
// Before
<Button variant="outline" className="h-10 text-base gap-2 bg-white" onClick={() => {}}>

// After（デザイントークン使用・サイズ統一・実際のハンドラ）
<Button variant="outline" className={`h-10 text-base gap-2 ${C.bgPage}`} onClick={handleImport}>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用（`#37352F` 等ハードコード禁止）

`bg-white` はハードコードされた Tailwind 色クラスであり、デザイントークンに違反する。

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **PROHIBITED**: Hexカラー（`#37352F` など）の直接指定は厳禁。すべてのスタイリングで `src/lib/design-tokens.ts` の定数を使用する。

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnersList.tsx` — ヘッダーアクションボタンの正しい実装（実際のハンドラ有り）

## 優先度
**Medium** — 押しても何も起きないボタンはユーザーに UI が壊れている印象を与える。日常業務で頻繁に使うページ（ワクチン・診察・在庫）に影響。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx:233`
- `frontend/src/features/examinations/routes/ExaminationsList.tsx:250`
- `frontend/src/features/inventory/routes/InventoryList.tsx:254`
