# TASK-008: マスタ設定ページ 第2段階統一 — handleSave + MasterCRUDPage 抽出

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

TASK-007（FE-026〜030）で `useMasterCRUD` hook と `MasterListPage` コンポーネントを導入済みだが、各ページにはまだ大量の重複コードが残っている。第2段階として、`handleSave` パターンの共通化と `MasterCRUDPage` 高レベルラッパーを作成し、9つの標準CRUDページを30〜50行程度に圧縮する。

## 依頼内容（原文）

> フロントエンドの設定マスタのすべてのページにて、できる限り同じコンポーネントを使用するようにしてください。
> vercel-react-best-practicesのベストプラクティスなコード規約に準拠する実装にして。

## 現状分析

### 第1段階（TASK-007）完了後の状態

| 共有済み | ファイル | 行数 |
|---------|---------|------|
| `useMasterCRUD<T>` hook | `hooks/use-master-crud.ts` | 154 |
| `MasterListPage` component | `components/MasterListPage.tsx` | 117 |

### 残存する重複（全9ページ共通パターン）

1. **handleSave コールバック**（~30行/ページ × 9 = 270行）
   - バリデーション → `startSaveTransition` → create/update 分岐 → toast 通知
   - 差分: リクエスト型変換のみ

2. **MasterListPage props 組み立て**（~20行/ページ × 9 = 180行）
   - `sidePanel={crud.isEditing ? <XxxSidePanel ...> : null}`
   - `deleteOpen={crud.pendingDelete !== null}`
   - `deleteTitle`, `deleteDescription`, `onDeleteConfirm`, `onDeleteCancel`

3. **mutation hook 4行セット**（~4行/ページ × 9 = 36行）
   - `useGetXxx`, `useCreateXxx`, `useUpdateXxx`, `useDeleteXxx`

4. **INPUT_CLASS 定数**（5ページで重複定義）

**合計: ~540行の重複**

### 対象ページ分類

| パターン | ページ数 | ページ | 行数 |
|---------|---------|--------|------|
| A: Simple CRUD | 6 | JobTitle(214), ChiefComplaint(211), Insurance(252), InterviewTemplate(229), Staff(287), Hospitalization(299) | 1,492 |
| B: DnD CRUD | 3 | AnimalSpecies(223), ServiceType(299), CageSettings(420) | 942 |
| C: Tabbed | 3 | Diagnosis(622), Trimming(681), TreatmentPlan(774) | 2,077 |
| D: Complex | 1 | Medicine(915) | 915 |
| E: Form-only | 1 | CompanySettings(388) | 388 |

**本タスクのスコープ: パターンA(6) + パターンB(3) = 9ページ**

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | `useMasterSave` hook 作成 — handleSave パターン抽出 | FE | FE-031 | - |
| 2 | `MasterCRUDPage` 高レベルラッパー作成 | FE | FE-032 | - |
| 3 | パターンA 6ページ移行 | FE | FE-033 | #1, #2 |
| 4 | パターンB 3ページ（DnD）移行 | FE | FE-034 | #1, #2 |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend

**新規作成:**
- `frontend/src/features/master/hooks/use-master-save.ts` — handleSave 共通 hook
- `frontend/src/features/master/components/MasterCRUDPage.tsx` — 高レベルラッパー

**修正:**
- `frontend/src/features/master/routes/JobTitleSettings.tsx` — 214 → ~60行
- `frontend/src/features/master/routes/ChiefComplaintSettings.tsx` — 211 → ~60行
- `frontend/src/features/master/routes/InsuranceSettings.tsx` — 252 → ~80行
- `frontend/src/features/master/routes/InterviewTemplateSettings.tsx` — 229 → ~70行
- `frontend/src/features/master/routes/StaffSettings.tsx` — 287 → ~100行
- `frontend/src/features/master/routes/HospitalizationSettings.tsx` — 299 → ~100行
- `frontend/src/features/master/routes/AnimalSpeciesSettings.tsx` — 223 → ~70行
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx` — 299 → ~100行
- `frontend/src/features/master/routes/CageSettings.tsx` — 420 → ~130行

**期待される削減: ~2,434行 → ~770行（約68%削減）**

## 実装順序

1. FE-031 + FE-032（共通 hook + コンポーネント — 並行可）
2. FE-033（パターンA 6ページ移行）
3. FE-034（パターンB 3ページ移行）

## 関連イシュー

- [FE-031: useMasterSave hook](../../frontend/issues/open/FE-031-use-master-save-hook.md)
- [FE-032: MasterCRUDPage コンポーネント](../../frontend/issues/open/FE-032-master-crud-page-component.md)
- [FE-033: パターンA 6ページ移行](../../frontend/issues/open/FE-033-pattern-a-simple-crud-migration.md)
- [FE-034: パターンB 3ページ（DnD）移行](../../frontend/issues/open/FE-034-pattern-b-dnd-migration.md)

## 前提タスク

- TASK-007（FE-026〜030）: 完了済み
