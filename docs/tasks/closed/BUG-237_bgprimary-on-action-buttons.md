# BUG-237: アクションボタンで `C.bgPrimary`（Notion gray）を誤用 — `C.bgAccent`（brand blue）が正しい

## 概要

`C.bgPrimary = "bg-[#37352F]"`（Notion document primary gray）は `STYLE.iconButton`（h-8 w-8 の小型アイコンボタン）専用のトークン。しかし `ReservationDetailModal.tsx`・`ReservationFormModal.tsx`・`VaccinationHistory.tsx` の 3 箇所で、h-9/h-10 の通常サイズアクションボタンに `C.bgPrimary` が使われており、ユーザーアクションを表す primary button が brand blue ではなく暗いグレーで表示されている。BUG-232（EstimateForm）と同種の問題。

## 再現手順

1. 予約詳細モーダル（ReservationDetailModal）を開く
2. 「カルテ作成」系ボタンを確認する
3. **結果**: 暗いグレー背景（`#37352F`）のボタンが表示される
4. 見積作成フォームや検査フォームの保存ボタン（brand blue `#2383E2`）と比較する
5. **期待**: ページ内のアクションボタンはすべて brand blue を使用すること

## 期待する動作

- h-9/h-10 以上の通常アクションボタンは `C.bgAccent`（`bg-[#2383E2]`、brand blue）を使用

## 現状コード

### `features/reservations/components/ReservationDetailModal.tsx:226`
```tsx
// ❌ C.bgPrimary (dark gray #37352F) on h-9 button
className={`${C.bgPrimary} ${C.textWhite} ${C.hoverBgPrimaryDark} h-9 text-sm gap-1.5 shadow-sm`}
```

### `components/shared/ReservationFormModal/ReservationFormModal.tsx:319`
```tsx
// ❌ C.bgPrimary (dark gray #37352F) on h-10 button
className={`${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} h-10 text-sm min-w-[100px]`}
```

### `features/medical-records/components/VaccinationHistory.tsx:138`
```tsx
// ❌ C.bgPrimary (dark gray #37352F) on h-10 button
className={`h-10 w-[50px] text-sm ${C.bgPrimary} ${C.textWhite} ${C.hoverBgPrimaryDark} hover:text-white border-transparent px-0`}
```

### デザイントークン確認

```typescript
// src/lib/design-tokens.ts
C.bgPrimary   = "bg-[#37352F]"   // ← Notion document primary gray
C.bgAccent    = "bg-[#2383E2]"   // ← brand blue（アクションボタン正しい色）

// C.bgPrimary の正しい使用場所
STYLE.iconButton = `h-8 w-8 ${C.bgPrimary} text-white ...`  // ← 小型アイコンボタン専用
```

### 比較: 正しい実装（同プロジェクト参照実装）

```tsx
// ✅ features/examinations/routes/ExaminationForm.tsx:182
className={`${C.bgAccent} ${C.bgAccentHover} text-white h-10 text-sm`}

// ✅ features/hospitalization/routes/HospitalizationForm.tsx:174
className={`${C.bgAccent} ${C.bgAccentHover} text-white rounded-[6px] h-10 text-sm px-4`}
```

## 影響範囲

| 対象 | 行 | ボタンサイズ | 状態 |
|------|-----|------------|------|
| `features/reservations/components/ReservationDetailModal.tsx` | 226 | h-9 | 未修正 |
| `components/shared/ReservationFormModal/ReservationFormModal.tsx` | 319 | h-10 | 未修正 |
| `features/medical-records/components/VaccinationHistory.tsx` | 138 | h-10 | 未修正 |

## 修正方針

各ファイルで `C.bgPrimary` → `C.bgAccent`、`C.hoverBgPrimaryDark` → `C.bgAccentHover` に置換する。

```tsx
// Before
className={`${C.bgPrimary} ${C.textWhite} ${C.hoverBgPrimaryDark} h-9 text-sm`}

// After
className={`${C.bgAccent} ${C.textWhite} ${C.bgAccentHover} h-9 text-sm`}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用

### プロジェクト内参照実装
- `features/examinations/routes/ExaminationForm.tsx:182` — `C.bgAccent` の正しいアクションボタン
- `features/hospitalization/routes/HospitalizationForm.tsx:174` — 同様
- `src/lib/design-tokens.ts:790` — `STYLE.iconButton` が `C.bgPrimary` の唯一の正しい使用場所

## 優先度
**Medium** — ページによってアクションボタンの色が異なる視覚的不一致。予約詳細モーダルと予約フォームモーダルで灰色ボタン、他ページで brand blue ボタンが表示される。

## 関連チケット
- BUG-232: EstimateForm の同種違反（`C.bgPrimary` on SubmitButton）
- BUG-231: ghost-danger ボタンの色不一致

## 関連ファイル
- `frontend/src/features/reservations/components/ReservationDetailModal.tsx:226`
- `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx:319`
- `frontend/src/features/medical-records/components/VaccinationHistory.tsx:138`
- `frontend/src/lib/design-tokens.ts:281,430,444,790`
