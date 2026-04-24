# FE-224: shared コンポーネント群のデザイントークン違反（7ファイル）

## 概要

`frontend/src/components/shared/` 配下の複数コンポーネントで
直接 Tailwind カラークラスが使用されている。共有コンポーネントの違反はアプリ全体に波及するため優先度が高い。

## 違反ファイル一覧

### `ReservationFormModal/ReservationFormModal.tsx` — 7箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 38 | `text-blue-600`（アクティブステップ） | `C.textAccent` 相当 |
| 41 | `bg-blue-600 text-white`（アクティブステップ丸） | `C.bgAccent text-white` |
| 64 | `hover:bg-red-50`（削除ボタン hover） | `C.hoverBgDanger5` 相当 |
| 66 | `text-red-600 hover:text-red-700`（削除アイコン） | `C.danger` + hover トークン |
| 183 | `text-blue-600`（Calendar アイコン） | `C.textAccent` 相当 |
| 225 | `border-red-600 bg-red-50`（初診ラジオ選択時） | デザイントークンに |
| 239 | `border-blue-600 bg-blue-50`（再診ラジオ選択時） | デザイントークンに |

### `PetSelection/PetSelection.tsx` — 8箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 56 | `bg-blue-50/50`（選択済みペット背景） | `C.bgAccentLight` 相当 |
| 74 | `text-blue-600`（チェックアイコン） | `C.textAccent` 相当 |
| 82 | `bg-blue-50/50`（選択済みコンテナ） | 同上 |
| 83 | `text-blue-700`（見出し） | `C.textAccent` 相当 |
| 88 | `border-blue-200 text-blue-800`（選択済みペットチップ） | デザイントークンに |
| 91 | `text-blue-400`（区切り線） | デザイントークンに |
| 94 | `hover:bg-blue-100`（削除ボタン hover） | デザイントークンに |

### `PatientInfoCard/PatientInfoCard.tsx` — 1箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 62 | `bg-gray-50/80`（死亡ペットの背景） | `C.bgLight` または `C.bgMuted` |

### `RowActionDropdown/RowActionDropdown.tsx` — 1箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 47 | `text-red-600 focus:text-red-600 focus:bg-red-50`（削除アクション） | `C.danger` + `C.bgDangerLight` 相当 |

### `DeleteIconButton/DeleteIconButton.tsx` — 1箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 31 | `hover:text-red-600 hover:bg-red-50`（削除アイコンボタン） | `C.danger` + hover トークン |

### `CharCountTextarea/CharCountTextarea.tsx` — 1箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 52 | `text-red-500`（文字数超過時） | `isOver ? C.danger : C.text40` |

### `Feedback/ImageWithFallback.tsx` — 1箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 18 | `bg-gray-100`（フォールバック背景） | `C.bgMedium` または `C.bgLight` |

### `PermissionBadges/PermissionBadges.tsx` — 1箇所

| 行 | 違反コード | 修正方針 |
|----|-----------|---------|
| 35, 60 | `bg-neutral-100 border-neutral-200`（バッジ背景・ボーダー） | `C.bgLight`, `C.borderLight` |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**High** — 共有コンポーネントの違反はアプリ全体に波及する。特に `PetSelection`・`ReservationFormModal`・`RowActionDropdown` は多数ページで使用される。

## 関連ファイル
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx`
- `frontend/src/components/shared/PetSelection/PetSelection.tsx`
- `frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx`
- `frontend/src/components/shared/RowActionDropdown/RowActionDropdown.tsx`
- `frontend/src/components/shared/DeleteIconButton/DeleteIconButton.tsx`
- `frontend/src/components/shared/CharCountTextarea/CharCountTextarea.tsx`
- `frontend/src/components/shared/Feedback/ImageWithFallback.tsx`
- `frontend/src/components/shared/PermissionBadges/PermissionBadges.tsx`
- `frontend/src/lib/design-tokens.ts`
