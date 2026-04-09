# BUG-174: AccountingDocument.tsx 印刷ビューのハードコードカラー違反

## 概要

`features/accounting/routes/AccountingDocument.tsx`（印刷・PDF出力用ビュー）で Tailwind プリセットカラーをハードコード使用している。印刷ビューは特殊な用途であるが、将来のテーマ変更・ブランドカラー変更時に漏れが発生しやすいため、他のファイルと同様にトークンを使用すること。

## 再現手順

1. 会計詳細画面でレシート印刷 / PDF出力ボタンをクリック
2. 印刷プレビューを表示する
3. ソースコードで `AccountingDocument.tsx` を開き `bg-` / `text-` カラークラスを検索
4. **結果**: Tailwind プリセットカラーがハードコードされている

## 期待する動作

- 印刷ビューであっても `C.*` / `STYLE.*` トークンを使用する
- ただし印刷専用の `@media print` スタイルが必要な場合は Tailwind の `print:` バリアントと組み合わせて使用する

## 現状コード

### `frontend/src/features/accounting/routes/AccountingDocument.tsx`
```tsx
// ❌ ヘッダー・テーブル行・テキスト色のハードコード（具体行番号は要確認）
<div className="bg-gray-100 text-gray-800 font-medium">
<tr className="bg-gray-50 border-b border-gray-200">
<td className="text-right text-green-700 font-semibold">
<span className="text-red-500">
<p className="text-blue-600">
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ テーブルヘッダー
<div style={{ backgroundColor: C.bgHover, color: C.textMain }} className="font-medium">

// ✅ 合計行
<td style={{ color: C.bgStatusGreenDot }} className="text-right font-semibold">

// ✅ マイナス値
<span style={{ color: C.bgDanger }}>

// ✅ リンク・アクセント
<p style={{ color: C.bgAccent }}>
```

## 影響範囲

| 対象ファイル | 違反箇所 | 状態 |
|---|---|---|
| `features/accounting/routes/AccountingDocument.tsx` | bg-gray-100/50, text-gray-800, text-green-700, text-red-500, text-blue-600, border-gray-200 等 | 未修正 |

## 修正方針

### 印刷ビューのトークン置換方針

印刷ビューでは `@media print` でのスタイル上書きが必要な場合があるが、基本スタイルは他画面と同様トークンを使用する:

```tsx
import { C } from '@/lib/design-tokens';

// テーブル行の背景
<tr style={{ backgroundColor: C.bgHover }}>

// ヘッダーテキスト
<th style={{ color: C.textMain }}>

// 合計額テキスト
<td style={{ color: C.bgStatusGreenDot, fontWeight: 600 }}>

// マイナス/エラー金額
<span style={{ color: C.bgDanger }}>

// ボーダー
<tr style={{ borderBottomColor: C.border }}>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

印刷ビューも例外ではない。

### プロジェクト内参照実装
- `features/accounting/routes/AccountingDetail.tsx` — 同 feature の修正対象（BUG-172）

## 優先度
**Low** — 印刷ビューは限られたユーザーが使用し、機能的な問題はない。ただし BUG-172（AccountingDetail）の修正と同時に対応するとコストが低い。

## 関連チケット
- BUG-172: AccountingDetail.tsx の包括的色違反（セットで修正推奨）
- BUG-162: 複数 feature のハードコードカラー違反

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDocument.tsx`
- `frontend/src/lib/design-tokens.ts` — C トークン定義
