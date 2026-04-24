# FE-215: 予約機能のデザイントークン違反（MonthView・WeekView）

## 概要

`frontend/src/features/reservations/components/` 配下の `MonthView.tsx` と `WeekView.tsx` で、
直接 Tailwind カラークラスを使用している。プロジェクト規約に違反。

## 影響ファイルと違反箇所

### `frontend/src/features/reservations/components/MonthView.tsx`

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 39 | `text-red-500`（日曜日ヘッダー） | デザイントークン（`C.danger` 相当）に置換 |
| 39 | `text-blue-500`（土曜日ヘッダー） | デザイントークン（`C.textAccent` 相当）に置換 |
| 73 | `bg-blue-50/30`（今日の日付ハイライト） | `C.bgAccent` 相当 + 透過クラスに置換 |
| 79 | `bg-blue-600 text-white hover:bg-blue-100 hover:text-blue-700`（日付ボタン） | `C.bgAccent`, `C.hoverBgAccent` 等に置換 |
| 105 | `bg-red-100 text-red-600`（"初"バッジ = 初診） | `C.bgRedLight`, `C.textNotionRed` 相当に置換 |
| 106 | `bg-blue-100 text-blue-600`（"再"バッジ = 再診） | デザイントークンに置換 |

### `frontend/src/features/reservations/components/WeekView.tsx`

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 215 | `decoration-red-500/50`（キャンセル予約の打ち消し線色） | デザイントークンに置換 |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hex カラー・直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Medium** — 機能的障害はないが、将来のテーマ変更時に修正漏れとなる。

## 関連ファイル
- `frontend/src/features/reservations/components/MonthView.tsx`
- `frontend/src/features/reservations/components/WeekView.tsx`
- `frontend/src/lib/design-tokens.ts`
