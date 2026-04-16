# BUG-185: medical-records 印刷ビュー・VaccinationForm 内部の gray 系 Tailwind 大量違反

## 概要

`features/medical-records/components/MedicalRecordPrintView.tsx` と `features/medical-records/components/VaccinationForm.tsx`（カルテ内ワクチンフォーム）で `border-gray-400`・`text-gray-500`・`text-gray-600`・`border-gray-300` 等の Tailwind gray プリセットが大量使用されている。BUG-174（AccountingDocument 印刷ビュー）と同様のパターン。

## 再現手順

1. 電子カルテの印刷ボタンをクリックして印刷プレビューを表示する
2. ソースコードで `MedicalRecordPrintView.tsx` を確認
3. **結果**: `border-gray-400`, `text-gray-600`, `border-gray-300`, `text-gray-500`, `border-gray-200` が多数ハードコード

## 期待する動作

- gray 系テキスト: `C.textSecondary` / `C.text60` / `C.text50` 等を使用
- gray 系ボーダー: `C.borderMedium` / `C.border` 等を使用
- gray 系背景: `C.bgHover` / `C.bgPage` 等を使用

## 現状コード

### `frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx:69-143（多数）`
```tsx
// ❌ gray 系ハードコード（代表例）
<div className="border-gray-400 ...">
<p className="text-gray-600">
<table className="border-gray-300 ...">
<td className="text-gray-500 border-gray-200 ...">
<th className="border-gray-400 bg-gray-100 ...">
```

### `frontend/src/features/medical-records/components/VaccinationForm.tsx:106-153（多数）`
```tsx
// ❌ border-gray-400 多数 (L106, L112, L116, L120, L126, L135, L141, L147, L153)
<input className="... border-gray-400 ...">
<select className="... border-gray-400 ...">
```

### `frontend/src/features/medical-records/components/TreatmentTable.tsx:127,132`
```tsx
// ❌ hover・テキスト gray ハードコード
<tr className="hover:bg-gray-50">
<span className="text-gray-300">—</span>
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ ボーダー
style={{ borderColor: C.borderMedium }}

// ✅ テキスト
style={{ color: C.textSecondary }}

// ✅ 背景
style={{ backgroundColor: C.bgHover }}
```

## 影響範囲

| 対象ファイル | 違反箇所数 | 種別 | 状態 |
|---|---|---|---|
| `features/medical-records/components/MedicalRecordPrintView.tsx` | 20+箇所 (L69〜L143) | border-gray-*, text-gray-*, bg-gray-* | 未修正 |
| `features/medical-records/components/VaccinationForm.tsx` | 9箇所 (L106, L112, L116, L120, L126, L135, L141, L147, L153) | border-gray-400 | 未修正 |
| `features/medical-records/components/TreatmentTable.tsx` | 2箇所 (L127, L132) | hover:bg-gray-50, text-gray-300 | 未修正 |

## 修正方針

### デザイントークンマッピング

| Tailwind クラス | 代替トークン |
|---|---|
| `border-gray-400` | `style={{ borderColor: C.borderMedium }}` |
| `border-gray-300` | `style={{ borderColor: C.border }}` |
| `border-gray-200` | `style={{ borderColor: C.borderLight }}` |
| `text-gray-600` | `style={{ color: C.textSecondary }}` |
| `text-gray-500` | `style={{ color: C.text50 }}` (または `C.textSecondary`) |
| `text-gray-300` | `style={{ color: C.text30 }}` (または適切なトークン) |
| `bg-gray-100` | `style={{ backgroundColor: C.bgHover }}` |
| `bg-gray-50` | `style={{ backgroundColor: C.bgPage }}` |
| `hover:bg-gray-50` | `className="hover:bg-[${C.bgPage}]"` |

### 印刷ビューでの注意点

`MedicalRecordPrintView.tsx` は印刷・PDF 出力目的のコンポーネント。Tailwind の `print:` バリアントと組み合わせて使用する場合は、inline style で CSS 変数参照が必要か確認すること:
```tsx
// 印刷でも正しく色が出るよう inline style を使用
<td style={{ borderColor: C.borderMedium, color: C.textSecondary }}>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

### プロジェクト内参照実装
- BUG-174 の修正後 `AccountingDocument.tsx` — 印刷ビューでのトークン使用参照実装

## 優先度
**Low** — 印刷ビューはUI表示に直接影響しないが、BUG-174（AccountingDocument）と合わせて対応することで効率が高い。`VaccinationForm.tsx` の border-gray-400 は UI に表示されるため優先度を Medium に引き上げてもよい。

## 関連チケット
- BUG-174: AccountingDocument.tsx 印刷ビューの color 違反（同パターン）
- BUG-162: medical-records の他の color 違反

## 関連ファイル
- `frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx`
- `frontend/src/features/medical-records/components/VaccinationForm.tsx`（カルテ内部フォーム）
- `frontend/src/features/medical-records/components/TreatmentTable.tsx`
- `frontend/src/lib/design-tokens.ts`
