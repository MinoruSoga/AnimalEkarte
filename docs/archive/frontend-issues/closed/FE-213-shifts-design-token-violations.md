# FE-213: シフト機能のデザイントークン違反（直 Tailwind カラークラス多数）

## 概要

`frontend/src/features/shifts/` 配下の複数ファイルで、デザイントークン（`C.*`）を使わずに
直接 Tailwind カラークラスを使用している。プロジェクト規約「Hex カラー・直接 Tailwind カラークラス禁止」に違反。

## 影響ファイルと違反箇所

### `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx`

| 行 | 違反 | 修正方針 |
|----|------|---------|
| 54 | `text-red-500`（API エラー文字色） | `C.danger` または `C.textRequired` |

### `frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx`

| 行 | 違反 | 修正方針 |
|----|------|---------|
| 20 | `text-gray-600`, `border-gray-200`（ヘッダー列） | `C.text60`, `C.borderMedium` |
| 185 | `text-red-500`（日曜日日付色） | `C.danger` |
| 187 | `text-blue-500`（土曜日日付色） | `C.bgAccent`相当のテキストトークン |
| 188 | `text-gray-700`（平日テキスト） | `C.text80` または `C.text` |
| 205 | `border-gray-100`, `hover:bg-gray-50/50`（スタッフ行） | `C.borderLight`, hover トークン |
| 208 | `text-gray-800`（スタッフ名テキスト） | `C.text` または `C.text90` |
| 238 | `text-gray-400`（空状態メッセージ） | `C.text40` または `C.text50` |

### `frontend/src/features/shifts/components/ShiftCell/ShiftCell.tsx`

| 行 | 違反 | 修正方針 |
|----|------|---------|
| 39 | `text-gray-300 hover:text-gray-500 hover:bg-gray-50`（追加ボタン） | `C.*` トークンへ置換 |

### `frontend/src/features/shifts/types/index.ts` ⚠️ 最重要

```ts
// 現状: SHIFT_TYPE_COLORS がすべて直接 Tailwind クラス
export const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  [ShiftTypeFull]:      "bg-blue-100 text-blue-800 border-blue-200",
  [ShiftTypeMorning]:   "bg-green-100 text-green-800 border-green-200",
  [ShiftTypeAfternoon]: "bg-teal-100 text-teal-800 border-teal-200",
  [ShiftTypeOff]:       "bg-gray-100 text-gray-600 border-gray-200",
  [ShiftTypePaidLeave]: "bg-purple-100 text-purple-800 border-purple-200",
};
```

シフトタイプバッジ（5種類）のカラーマップがすべてハードコード。
`design-tokens.ts` に専用トークンを追加し、定数を参照するよう変更すること。

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: Hex カラー・直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Medium** — 機能的障害はないが、将来のテーマ変更時に修正漏れとなる。

## 関連ファイル
- `frontend/src/features/shifts/routes/ShiftCalendarPage.tsx`
- `frontend/src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx`
- `frontend/src/features/shifts/components/ShiftCell/ShiftCell.tsx`
- `frontend/src/features/shifts/types/index.ts`
- `frontend/src/lib/design-tokens.ts` — トークン追加先
