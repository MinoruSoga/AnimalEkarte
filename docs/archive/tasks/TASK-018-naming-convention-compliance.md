# TASK-018: ファイル命名規則を kebab-case に統一（Vercel React Best Practices / naming-analyzer 準拠）

**作成日**: 2026-03-18
**ステータス**: Closed
**依頼元**: ユーザー

---

## 概要

フロントエンドの hook ファイル名が camelCase（`useOwnerForm.ts`）と kebab-case（`use-owner.ts`）で混在している。プロジェクト規約（ESLint check-file 強制）および Vercel React Best Practices / naming-analyzer の推奨に準拠し、全 hook ファイル名を kebab-case に統一する。

## 依頼内容（原文）

> 変数、メソッド名などの命名規則を下記の2つのSkillsに準拠するタスクを作成して。
> ・vercel-react-best-practices
> ・naming-analyzer

## 仕様確認ログ

確認事項なし。プロジェクト規約（CLAUDE.md / CODING_RULES.md）で「ファイル名は kebab-case」と明確に定義されており、曖昧な点がない。

## 調査結果サマリー

### 違反状況

| カテゴリ | 違反数 | 重要度 |
|---------|--------|--------|
| FE: Hook ファイル名 camelCase | 26 ファイル | **高**（規約違反） |
| FE: API mutation hook 命名 | 0 | - （全て正しい） |
| FE: イベントハンドラ命名 | 0 | - （全て正しい） |
| FE: Boolean 変数命名 | 0 | - （全て正しい） |
| FE: 型/Interface 命名 | 0 | - （全て PascalCase） |
| BE: パッケージ/ファイル/インターフェース命名 | 0 | - （全て正しい） |
| BE: エクスポート/非エクスポート関数命名 | 0 | - （全て正しい） |
| BE: Request/Response struct 非エクスポート | 適正 | - （handler 内部型として意図的設計） |

### 結論

**唯一の違反は hook ファイル名の camelCase 混在のみ。** 関数名・変数名・型名・メソッド名は全て規約に準拠しており、修正不要。

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | 共有 hooks のファイル名 kebab-case 化 + import 更新 | FE | FE-071 | - | [x] |
| 2 | Feature hooks のファイル名 kebab-case 化 + import 更新 | FE | FE-072 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: `frontend/src/hooks/` 内の全 `.ts` ファイルが kebab-case である
- [ ] AC-2: `frontend/src/features/*/hooks/` 内の全 `.ts`/`.tsx` ファイルが kebab-case である
- [ ] AC-3: 全 import パスが新しい kebab-case ファイル名を参照している
- [ ] AC-4: `npm run build` がエラーなく通る
- [ ] AC-5: `npm run lint` がエラーなく通る
- [ ] AC-6: feature hooks の barrel index（`index.ts`）も更新済み

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| BE struct 命名 | 変更しない | handler 内部 DTO の非エクスポートは Go の idiom として適正。パッケージ外に公開しない型を小文字にするのは正しい設計 | PascalCase に変更 |
| BE 変数 `s` | 変更しない | if-block 内のスコープ限定1文字変数は Effective Go 準拠 | `statusStr` 等に変更 |

## 影響範囲

### Backend
- 変更なし

### Frontend — 共有 hooks（5 ファイル rename）

| 現在のファイル名 | 変更後 |
|----------------|--------|
| `hooks/usePagination.ts` | `hooks/use-pagination.ts` |
| `hooks/useStaffValidation.ts` | `hooks/use-staff-validation.ts` |
| `hooks/useReducedMotion.ts` | `hooks/use-reduced-motion.ts` |
| `hooks/useUnsavedChanges.ts` | `hooks/use-unsaved-changes.ts` |
| `hooks/useSortableList.ts` | `hooks/use-sortable-list.ts` |

### Frontend — Feature hooks（21 ファイル rename）

| 現在のファイル名 | 変更後 |
|----------------|--------|
| `auth/hooks/useAuth.tsx` | `auth/hooks/use-auth.tsx` |
| `owners/hooks/useOwnerForm.ts` | `owners/hooks/use-owner-form.ts` |
| `dashboard/hooks/useDashboardKanban.ts` | `dashboard/hooks/use-dashboard-kanban.ts` |
| `estimates/hooks/useEstimateForm.ts` | `estimates/hooks/use-estimate-form.ts` |
| `examinations/hooks/useExaminationForm.ts` | `examinations/hooks/use-examination-form.ts` |
| `examinations/hooks/useExaminationRecords.ts` | `examinations/hooks/use-examination-records.ts` |
| `hospitalization/hooks/useHospitalizationForm.ts` | `hospitalization/hooks/use-hospitalization-form.ts` |
| `hospitalization/hooks/useHospitalizationList.ts` | `hospitalization/hooks/use-hospitalization-list.ts` |
| `hospitalization/hooks/useHospitalizationDetail.ts` | `hospitalization/hooks/use-hospitalization-detail.ts` |
| `hospitalization/hooks/useHospitalizations.ts` | `hospitalization/hooks/use-hospitalizations.ts` |
| `hospitalization/hooks/useDailyRecordLogic.ts` | `hospitalization/hooks/use-daily-record-logic.ts` |
| `hospital-settings/hooks/useClinicSettingsForm.ts` | `hospital-settings/hooks/use-clinic-settings-form.ts` |
| `inventory/hooks/useInventory.ts` | `inventory/hooks/use-inventory.ts` |
| `master/hooks/useServiceTypeColorMap.ts` | `master/hooks/use-service-type-color-map.ts` |
| `medical-records/hooks/useMedicalRecords.ts` | `medical-records/hooks/use-medical-records.ts` |
| `medical-records/hooks/useMedicalRecordForm.ts` | `medical-records/hooks/use-medical-record-form.ts` |
| `reservations/hooks/useReservationManagement.ts` | `reservations/hooks/use-reservation-management.ts` |
| `trimming/hooks/useTrimmingForm.ts` | `trimming/hooks/use-trimming-form.ts` |
| `trimming/hooks/useTrimmingRecords.ts` | `trimming/hooks/use-trimming-records.ts` |
| `vaccinations/hooks/useVaccinationForm.ts` | `vaccinations/hooks/use-vaccination-form.ts` |
| `vaccinations/hooks/useVaccinations.ts` | `vaccinations/hooks/use-vaccinations.ts` |

### Frontend — import 更新が必要なファイル（~40 箇所）

共有 hooks の import 更新（21 箇所）:
- `@/hooks/usePagination` → `@/hooks/use-pagination` (3 ファイル)
- `@/hooks/useStaffValidation` → `@/hooks/use-staff-validation` (2 ファイル)
- `@/hooks/useReducedMotion` → `@/hooks/use-reduced-motion` (2 ファイル)
- `@/hooks/useUnsavedChanges` → `@/hooks/use-unsaved-changes` (8 ファイル)
- `@/hooks/useSortableList` → `@/hooks/use-sortable-list` (6 ファイル)

Feature hooks の import 更新（~20 箇所）:
- 各 feature 内の相対 import パス更新
- `features/auth/index.ts` の re-export パス更新
- cross-feature import（`useAuth`, `useServiceTypeColorMap`）のパス更新
- `hospitalization/hooks/index.ts` の re-export パス更新

## 参照実装

- `features/master/hooks/use-master-items.ts` — 既に kebab-case で命名されている正しい例
- `hooks/use-pet-selection.ts` — 共有 hooks の正しい kebab-case 例

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| import パスの更新漏れでビルドエラー | 高 | `npm run build` で全エラーを検出可能。grep で旧パスが残っていないか確認 |
| git の大文字/小文字区別（macOS デフォルト case-insensitive） | 中 | `git mv` でファイル rename することで正しく追跡される |
| barrel index の更新漏れ | 中 | `hospitalization/hooks/index.ts`, `auth/index.ts` を忘れずに更新 |

## 未解決事項

- なし

## 実装順序

1. FE-071: 共有 hooks rename + import 更新（影響ファイル少ない、先にやると安全）
2. FE-072: Feature hooks rename + import 更新

## 関連イシュー

- [FE-071: 共有 hooks ファイル名 kebab-case 化](../../frontend/issues/open/FE-071-shared-hooks-kebab-case.md)
- [FE-072: Feature hooks ファイル名 kebab-case 化](../../frontend/issues/open/FE-072-feature-hooks-kebab-case.md)
