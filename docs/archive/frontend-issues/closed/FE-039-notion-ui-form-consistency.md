# FE-039: フォーム要素のNotion UI整合性改善 — 色hardcoding→デザイントークン統一

**親タスク**: [TASK-011](../../docs/tasks/open/TASK-011-notion-ui-consistency-audit.md)
**Status**: Open
**Priority**: Medium
**Affects**: VaccinationForm, ExaminationForm, HospitalizationForm, PetSelectionSearchForm, 他
**Date Created**: 2026-03-18
**Related**: なし

## Summary

design-tokens.ts に定義済みのNotionカラートークンが活用されず、各フォームで色がhardcodingされている。統一することで Notion UI の一貫性を確保する。

## 問題箇所と修正

### 1. VaccinationForm — 黒色ボタン（Critical）

```typescript
// frontend/src/features/vaccinations/routes/VaccinationList.tsx
// 検索対象: bg-black text-white

// Before（Notion UIに存在しない黒ボタン）
className="bg-black text-white hover:bg-black/90"

// After（Notion outline スタイル）
className={STYLE.btnOutline}
// または
className="border border-[rgba(55,53,47,0.16)] text-[#37352F]/60 hover:bg-[#F7F6F3]"
```

### 2. VaccinationForm — テーブル hover色（Critical）

```typescript
// 検索対象: hover:bg-gray-50

// Before（Tailwind cool gray）
className="hover:bg-gray-50"

// After（Notion暖色ホバー）
className="hover:bg-[#F7F6F3]"
// または C.hoverBgPage を使用
```

### 3. ExaminationForm — 赤色削除ボタン（Critical）

```typescript
// frontend/src/features/examinations/routes/ExaminationForm.tsx
// 検索対象: text-red-600

// Before（Tailwind red直書き）
className="text-red-600 hover:text-red-700 hover:bg-red-50"

// After（Notion danger スタイル）
className="text-[#EB5757] hover:bg-[#EB5757]/5"
// ※ #EB5757 は Notion の赤色
```

### 4. HospitalizationForm — 赤色削除ボタン（Critical）

```typescript
// frontend/src/features/hospitalization/routes/HospitalizationForm.tsx
// 検索対象: text-red-600

// 同上の修正
```

### 5. 全フォーム — 主要ボタン色の統一（Medium）

各フォームで `bg-[#2383E2] hover:bg-[#1B6EC2]` が直書きされている箇所を `STYLE.confirmPrimary` に統一する。

```typescript
// Before（各フォームで直書き）
className="bg-[#2383E2] hover:bg-[#1B6EC2] text-white"

// After（デザイントークン経由）
className={STYLE.confirmPrimary}
```

対象ファイル（grepで特定）:
- `grep -rn "bg-\[#2383E2\]" frontend/src/features/`
- `grep -rn "bg-\[#1B6EC2\]" frontend/src/features/`

### 6. 全フォーム — border直書きの統一（Medium）

```typescript
// Before（毎回直書き）
className="border-[rgba(55,53,47,0.16)]"

// After（デザイントークン経由）
className={C.borderMedium}
```

### 7. 全フォーム — text60直書きの統一（Medium）

```typescript
// Before（毎回直書き）
className="text-[#37352F]/60"

// After（デザイントークン経由）
className={C.text60}
```

## design-tokens.ts への追加が必要なトークン

以下のトークンが不足している場合は追加する:

```typescript
// 既に存在するか確認が必要
btnDanger: `text-[#EB5757] hover:bg-[#EB5757]/5 ...`,  // Notion赤ボタン
```

## 修正対象ファイル特定方法

```bash
# 黒ボタン
docker compose exec frontend grep -rn "bg-black" src/features/

# Tailwind gray hover
docker compose exec frontend grep -rn "hover:bg-gray" src/features/

# Tailwind red
docker compose exec frontend grep -rn "text-red-" src/features/

# 主要ボタン直書き
docker compose exec frontend grep -rn "bg-\[#2383E2\]" src/features/

# border直書き（件数確認）
docker compose exec frontend grep -rn "rgba(55,53,47,0.16)" src/features/ | wc -l
```

## Notion UI カラーリファレンス

| 用途 | Notion色 | Tailwind色（使用禁止） |
|------|---------|---------------------|
| テキスト | `#37352F` | `#000`, `#0f172a` |
| テキスト薄 | `#37352F/60` | `text-muted-foreground`, `text-gray-500` |
| 背景 | `#F7F6F3` | `bg-gray-50`, `bg-gray-100` |
| ボーダー | `rgba(55,53,47,0.16)` | `border-gray-200`, `border-gray-300` |
| ホバー | `#F7F6F3` or `rgba(55,53,47,0.04)` | `hover:bg-gray-50` |
| アクセント | `#2383E2` | `bg-blue-500`, `bg-blue-600` |
| 危険 | `#EB5757` | `text-red-600`, `text-red-700` |
| 成功 | `#4DAB9A` | `text-green-600` |
| 警告 | `#D9730D` | `text-orange-500` |

## プロジェクトルール遵守チェック

- [x] `any` 型なし
- [x] `FC` / `forwardRef` なし
- [x] barrel index 経由 import なし
- [x] 条件レンダー `? ... : null`（`&&` 禁止）
- [x] CSSクラスの変更のみ（ロジック変更なし）

## 完了条件

- [ ] `bg-black` がフォーム要素に使われていない
- [ ] `text-red-600`, `text-red-700`, `hover:bg-red-50` がフォーム要素に使われていない
- [ ] `hover:bg-gray-50` がフォーム要素に使われていない
- [ ] 主要ボタンが `STYLE.confirmPrimary` または同等のトークンを使用
- [ ] 全リスト・フォーム画面で目視確認（色の一貫性）
- [ ] `docker compose exec frontend pnpm lint` エラーなし
- [ ] `docker compose exec frontend pnpm build` 成功
