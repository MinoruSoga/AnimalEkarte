# BUG-219: StaffSettings の権限グループカラー表示にハードコード hex フォールバック

## 概要

`StaffSettings.tsx:358` と `StaffSettings.tsx:524-530` で、権限グループの色（ユーザー定義カラー）のフォールバック値として `#6B7280`・`#37352F` をハードコードしている。デザイントークン規約違反。

## 再現手順

1. マスタ設定 → スタッフ設定を開く（`/master/staff`）
2. スタッフの権限グループが未設定（`group.color === null`）の場合
3. グループ色インジケーターが `#6B7280`（gray-500 相当）で表示される
4. **結果**: style prop に hex ハードコード値が使われる

## 期待する動作

- フォールバック色は `PALETTE.defaultGray`（`#6B7280`）・`PALETTE.primary`（`#37352F`）等のトークン定数を使用すること

## 現状コード

### `frontend/src/features/master/routes/StaffSettings.tsx:356-359`
```tsx
<div
  className="w-2.5 h-2.5 rounded-full flex-shrink-0"
  style={{ backgroundColor: group.color ?? "#6B7280" }}
/>
```

### `frontend/src/features/master/routes/StaffSettings.tsx:521-531`
```tsx
<span
  style={{
    backgroundColor: g.color ? `${g.color}18` : "rgba(55,53,47,0.06)",
    color: g.color ?? "#37352F",
  }}
>
  <span
    className="size-1.5 rounded-full shrink-0"
    style={{ backgroundColor: g.color ?? "#6B7280" }}
  />
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// features/reservations/components/WeekView.tsx — PALETTE 使用の正しいパターン
style={{ borderColor: PALETTE.danger ?? "#C0392B" }}
// ↑ PALETTE 参照ならばハードコードは ?? のフォールバックとして許容されるが、
//   定義済みトークンがある場合は直接参照すべき
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `features/master/routes/StaffSettings.tsx:358` | `#6B7280` (group.color フォールバック) | 未修正 |
| `features/master/routes/StaffSettings.tsx:524` | `"rgba(55,53,47,0.06)"` (bg フォールバック) | 未修正 |
| `features/master/routes/StaffSettings.tsx:525` | `#37352F` (text color フォールバック) | 未修正 |
| `features/master/routes/StaffSettings.tsx:530` | `#6B7280` (dot color フォールバック) | 未修正 |

## 修正方針

`PALETTE` を import し、ハードコード hex を定数参照に変更する:

### `frontend/src/features/master/routes/StaffSettings.tsx:358`
```tsx
// Before
style={{ backgroundColor: group.color ?? "#6B7280" }}

// After
style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}
```

### `frontend/src/features/master/routes/StaffSettings.tsx:521-531`
```tsx
// Before
style={{
  backgroundColor: g.color ? `${g.color}18` : "rgba(55,53,47,0.06)",
  color: g.color ?? "#37352F",
}}
// ...
style={{ backgroundColor: g.color ?? "#6B7280" }}

// After
style={{
  backgroundColor: g.color ? `${g.color}18` : `${PALETTE.primary}0f`,
  color: g.color ?? PALETTE.primary,
}}
// ...
style={{ backgroundColor: g.color ?? PALETTE.defaultGray }}
```

対応トークン:
- `PALETTE.defaultGray` = `"#6B7280"` (design-tokens.ts:174)
- `PALETTE.primary` = `"#37352F"` (design-tokens.ts:19)
- `PALETTE.primary + "0f"` ≈ `rgba(55,53,47,0.06)` (hex 0f ≈ 6%)

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/typescript-react.md` §4 デザイントークン
> スタイリングには必ず `src/lib/design-tokens.ts` の定数を使用する。Hexカラーの直接指定は厳禁。

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用（`#37352F`等ハードコード禁止）

### プロジェクト内参照実装
- `frontend/src/lib/design-tokens.ts:174` — `PALETTE.defaultGray: "#6B7280"` 存在確認済み
- `frontend/src/lib/design-tokens.ts:19` — `PALETTE.primary: "#37352F"` 存在確認済み

## 優先度
**Low** — 機能的問題なし。色の見た目は変わらない（同値を定数経由に変えるのみ）。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/features/master/routes/StaffSettings.tsx:356-359,521-531`
- `frontend/src/lib/design-tokens.ts:17-19,172-174`
