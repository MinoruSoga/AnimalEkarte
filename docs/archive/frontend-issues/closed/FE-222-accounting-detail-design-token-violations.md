# FE-222: AccountingDetail のデザイントークン違反（18箇所）

## 概要

`frontend/src/features/accounting/routes/AccountingDetail.tsx` で
多数の直接 Tailwind カラークラスが使用されている。プロジェクト規約違反。

## 違反箇所一覧

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 147 | `text-blue-500 bg-blue-50`（カルテ連携バッジ） | `C.textAccent`, `C.bgAccentLight` |
| 187 | `text-green-600`（保険適用の bullet） | `C.textSuccess` 相当 |
| 189 | `text-gray-300`（非保険のダッシュ） | `C.text30` 相当 |
| 229 | `border-blue-600 text-blue-600`（アクティブタブ） | `C.bgAccent`, `C.textAccent` |
| 240 | `border-blue-600 text-blue-600`（同上、別条件） | 同上 |
| 312 | `text-red-500`（必須フィールドのアスタリスク） | `C.textRequired` または `C.danger` |
| 323 | `text-red-500`（同上） | 同上 |
| 386 | `bg-gray-50`（フッター背景） | `C.bgLight` |
| 416 | `bg-gray-50`（InsuranceCard ヘッダー） | `C.bgLight` |
| 440 | `text-green-700 bg-green-50`（保険金額表示） | デザイントークンへ |
| 487 | `text-gray-500`（請求金額ラベル） | `C.text50` |
| 579 | `bg-gray-100`（お釣り表示コンテナ） | `C.bgMedium` 相当 |
| 580 | `text-gray-600`（お釣りラベル） | `C.text60` |
| 582 | `text-red-500` / `text-gray-900`（お釣り金額、マイナス時） | `C.danger` / `C.text` |
| 644 | `bg-gray-50`（返金セクションヘッダー） | `C.bgLight` |
| 647 | `text-orange-500`（返金アイコン） | `C.textWarning` 相当 |
| 653 | `text-orange-600 bg-orange-50`（返金ステータスバッジ） | `C.textWarning`, `C.bgWarningLight` |
| 731 | `text-orange-600`（返金金額） | `C.textWarning` 相当 |
| 1157 | `bg-gray-100`（プレビュー背景） | `C.bgMedium` 相当 |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 補足

対応する `C.*` トークンが `design-tokens.ts` に存在しない場合は、
対応するトークンを追加してから参照すること（Hex カラー直書きへのフォールバックは禁止）。

## 優先度
**Medium** — 機能的障害はないが、一貫性のためテーマ変更に追従できない状態。

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/lib/design-tokens.ts`
