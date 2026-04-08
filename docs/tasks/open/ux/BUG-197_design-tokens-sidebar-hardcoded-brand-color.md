# BUG-197: design-tokens.ts の STYLE.sidebarItemActive にハードコードブランドカラー

## 概要

`src/lib/design-tokens.ts:777` の `STYLE.sidebarItemActive` に `#038B94`（ブランドカラー）が Tailwind 任意値クラス `bg-[#038B94]/8` および `border-l-[#038B94]` としてハードコードされている。BUG-180 で報告した L855（`STYLE.formInputError`）と同パターン。デザイントークン定義ファイル自身が自分のルールを破っている。

## 現状コード

### `frontend/src/lib/design-tokens.ts:777`
```ts
// ❌ #038B94 が Tailwind 任意値クラスにハードコード
sidebarItemActive: `bg-[#038B94]/8 ${C.text} border-l-2 border-l-[#038B94]`,
```

`#038B94` はプロジェクトのブランドカラー（teal 系）だが、`design-tokens.ts` 内の他の定数（`C.bgBrand`, `C.bgBrand10`, `C.borderBrand` 等）として既に定義されているはずのもので、それを参照すべき。

## 期待する動作

```ts
// ✅ C.* トークンを使用
sidebarItemActive: `${C.bgBrand10} ${C.text} border-l-2 ${C.borderBrand}`,
// または CSS 変数ベース
sidebarItemActive: `bg-[color-mix(in_sRGB,var(--color-brand)_8%,transparent)] ${C.text} border-l-2 border-l-[var(--color-brand)]`,
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 使用箇所 | 状態 |
|---|---|---|---|---|
| `src/lib/design-tokens.ts` | 777 | bg-[#038B94]/8, border-l-[#038B94] | サイドバーのアクティブ項目 | 未修正 |

## 修正方針

### Step 1: `design-tokens.ts` の `C` オブジェクトにブランド色トークンが存在するか確認
```ts
// C.bgBrand10 または C.brandLight 等が定義されているか確認
// なければ追加:
bgBrand10: "bg-[var(--color-brand-10)]",  // 8-10% 透明ブランド背景
borderBrand: "border-l-[var(--color-brand)]",
```

### Step 2: `STYLE.sidebarItemActive` を修正
```ts
// Before
sidebarItemActive: `bg-[#038B94]/8 ${C.text} border-l-2 border-l-[#038B94]`,

// After（C.* トークン使用）
sidebarItemActive: `${C.bgBrand10} ${C.text} border-l-2 ${C.borderBrand}`,
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hexカラー（`#038B94` など）の直接指定は厳禁。

デザイントークン定義ファイル自身がルール違反をしていると、Tailwind Linter によるチェックが困難になる。

## 優先度
**Medium** — サイドバーは全画面で使用される共通 UI。BUG-180（L855 の同パターン）と合わせて対応することで効率化できる。

## 関連チケット
- BUG-180: design-tokens.ts の STYLE.formInputError ハードコード red（同パターン）

## 関連ファイル
- `frontend/src/lib/design-tokens.ts`
