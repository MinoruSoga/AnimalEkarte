# BUG-220: WeekView のキャンセル予約に decoration-red-500/50 ハードコード

## 概要

`WeekView.tsx:215` でキャンセル済み予約の打ち消し線色に `decoration-red-500/50` をハードコードしている。デザイントークン規約違反。`design-tokens.ts` に `decoration-*` トークンが未定義のため、追加が必要。

## 再現手順

1. 予約管理の週表示（`/reservations` → 週ビュー）を開く
2. キャンセル状態の予約を確認する
3. テキストに打ち消し線が表示される
4. **結果**: `decoration-red-500/50`（Tailwind ハードコード）で打ち消し線色が付く

## 期待する動作

- 打ち消し線色はデザイントークン（`C.decorationDanger` 等）を使用すること

## 現状コード

### `frontend/src/features/reservations/components/WeekView.tsx:215`
```tsx
${isCancelled ? "line-through decoration-red-500/50" : ""}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// WeekView.tsx:445 — PALETTE.danger を正しく使用している例
style={{ borderColor: PALETTE.danger ?? "#C0392B" }}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `features/reservations/components/WeekView.tsx:215` | decoration-red-500/50 | 未修正 |

## 修正方針

### Step 1: `frontend/src/lib/design-tokens.ts` に decoration トークンを追加

```typescript
// C オブジェクトの適切な位置（danger 系トークン付近）に追加
decorationDanger50:  "decoration-[#C0392B]/50",
```

### Step 2: `frontend/src/features/reservations/components/WeekView.tsx:215`
```tsx
// Before
${isCancelled ? "line-through decoration-red-500/50" : ""}

// After
${isCancelled ? `line-through ${C.decorationDanger50}` : ""}
```

または即時修正として arbitrary value を使用:
```tsx
${isCancelled ? "line-through decoration-[#C0392B]/50" : ""}
```
（`PALETTE.danger = "#C0392B"` を参照）

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/typescript-react.md` §4 デザイントークン
> スタイリングには必ず `src/lib/design-tokens.ts` の定数を使用する。

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用（ハードコード禁止）

### プロジェクト内参照実装
- `frontend/src/features/reservations/components/WeekView.tsx:445` — 同ファイルで `PALETTE.danger` を正しく使用

## 優先度
**Low** — 機能的問題なし。視覚的差異も軽微。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/features/reservations/components/WeekView.tsx:215`
- `frontend/src/lib/design-tokens.ts` (decorationDanger50 トークン追加が必要)
