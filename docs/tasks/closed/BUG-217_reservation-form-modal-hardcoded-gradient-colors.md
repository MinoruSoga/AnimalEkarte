# BUG-217: ReservationFormModal のペット選択済み状態にハードコードグラデーション色

## 概要

`ReservationFormModal.tsx:247` で、ペット選択済み状態を示すハイライト背景に `from-blue-50/50 to-cyan-50/50 border-blue-100` をハードコードしている。デザイントークン規約違反。

## 再現手順

1. 予約フォームモーダルを開く（`/reservations` 等）
2. ペットを選択する
3. 上部の「予約対象（選択中）」セクション背景がブルーグラデーションに変わる
4. **結果**: `from-blue-50/50 to-cyan-50/50 border-blue-100` のハードコード色が使われる

## 期待する動作

- 背景色・ボーダー色はすべてデザイントークン（`C.*` / `PALETTE.*`）を使用すること

## 現状コード

### `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx:247`
```tsx
<div className={`rounded-lg border p-3 transition-colors ${selectedPets.length > 0
  ? "bg-gradient-to-r from-blue-50/50 to-cyan-50/50 border-blue-100"
  : `${C.bgPage} ${C.borderMediumLight}`}`}>
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// features/reservations/components/WeekView.tsx — PALETTE.danger を使用した正しいパターン
style={{ borderColor: PALETTE.danger ?? "#C0392B" }}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `components/shared/ReservationFormModal/ReservationFormModal.tsx:247` | from-blue-50/50, to-cyan-50/50, border-blue-100 | 未修正 |

## 修正方針

### `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx:247`

グラデーションの代わりに `C.bgAccentLight50` フラット背景 + `C.borderAccentLight` を使用する:

```tsx
// Before
? "bg-gradient-to-r from-blue-50/50 to-cyan-50/50 border-blue-100"

// After
? `${C.bgAccentLight50} ${C.borderAccentLight}`
```

`C.bgAccentLight50` = `"bg-[#D3E5EF]/50"` (既存トークン)
`C.borderAccentLight` = `"border-[#2383E2]/30"` (既存トークン)

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用（`#37352F`等ハードコード禁止）

### `.claude/rules/typescript-react.md` §4 デザイントークン
> スタイリングには必ず `src/lib/design-tokens.ts` の定数を使用する。Hexカラーの直接指定は厳禁。

### プロジェクト内参照実装
- `frontend/src/features/reservations/components/WeekView.tsx` — `PALETTE.danger` 使用
- `frontend/src/lib/design-tokens.ts:495` — `bgAccentLight50: "bg-[#D3E5EF]/50"` 存在確認済み
- `frontend/src/lib/design-tokens.ts:291` — `borderAccentLight: "border-[#2383E2]/30"` 存在確認済み

## 優先度
**Low** — 機能的問題はない。UI 上のグラデーションが消えるが視覚的に許容範囲。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx:247`
- `frontend/src/lib/design-tokens.ts:283-291,492-495`
