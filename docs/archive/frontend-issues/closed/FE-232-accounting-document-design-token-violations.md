# FE-232: AccountingDocument のデザイントークン違反

## 概要

`frontend/src/features/accounting/components/AccountingDocument.tsx`（領収書・請求書コンポーネント）で
直接 Tailwind カラークラスが使用されている。

## 違反箇所

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 56 | `text-gray-500` | 日付テキスト |
| 63 | `text-gray-500` | ペット名ラベル |
| 66 | `bg-gray-50` | 請求金額ボックス背景 |
| 77 | `text-green-700` | 保険適用金額テキスト |
| 96, 112, 115 | `text-gray-600` | 各種ラベル（3箇所） |
| 146, 163 | `text-gray-500` | カテゴリ・フッターテキスト（2箇所） |
| 177 | `text-green-700` | 保険控除額表示 |

## 修正方針

| 違反クラス | 置換先 |
|-----------|--------|
| `text-gray-500` | `C.text50` |
| `text-gray-600` | `C.text60` |
| `bg-gray-50` | `C.bgLight` |
| `text-green-700` | `C.textSuccess` 相当（design-tokens.ts に追加が必要な場合あり） |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Low** — 印刷・PDF 表示用コンポーネントのため機能的影響は軽微。テーマ変更時の一貫性のため修正が必要。

## 関連ファイル
- `frontend/src/features/accounting/components/AccountingDocument.tsx`
- `frontend/src/lib/design-tokens.ts`
