# FE-230: 各 feature ルートの必須フィールド・ステータス表示でデザイントークン未使用

## 概要

複数 feature のルートコンポーネントで、必須フィールドのアスタリスクや
在庫ステータス表示に直接 Tailwind カラークラスを使用している。

## 違反ファイル一覧

### `frontend/src/features/inventory/routes/InventoryList.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 270 | `text-amber-600` | 在庫アラートアイコン |
| 273 | `text-red-600` | 在庫切れステータスラベル |
| 278 | `text-amber-600` | 在庫少ステータスラベル |

### `frontend/src/features/inventory/routes/InventoryForm.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 65, 78, 101, 146, 161 | `text-red-500` | 必須フィールドのアスタリスク（5箇所） |

### `frontend/src/features/trimming/routes/TrimmingForm.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 173, 286 | `hover:bg-gray-100` | 画像プレビューの削除ボタン hover |

### `frontend/src/features/trimming/routes/TrimmingList.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 86 | `text-amber-500` | スタッフ名が無効な場合の警告アイコン |

### `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 196, 208 | `text-red-500` | 必須フィールドのアスタリスク（2箇所） |

## 備考

必須フィールドのアスタリスク `text-red-500` は全フォームで統一されていない。
`C.textRequired` または `C.danger` トークンへ統一すること。

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Low** — 機能的障害なし。デザイン一貫性のため修正が必要。

## 関連ファイル
- `frontend/src/features/inventory/routes/InventoryList.tsx`
- `frontend/src/features/inventory/routes/InventoryForm.tsx`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/trimming/routes/TrimmingList.tsx`
- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx`
- `frontend/src/lib/design-tokens.ts`
