# TASK-005: NotionFilter 未移行ページの移行 + 飼主一覧フィルタ拡充

**作成日**: 2026-03-17
**ステータス**: Open
**依頼元**: ユーザー（TASK-003 の残作業）

---

## 概要

TASK-003 で NotionFilter コンポーネントを作成し8つの一覧ページに適用したが、以下の問題が残っている:
1. 飼主一覧は NotionFilter を使っているが FilterProperty が空で「+ フィルタを追加」が非表示
2. 見積一覧が SearchFilterBar のまま未移行
3. マスタ設定画面（13画面）が SearchFilterBar のまま未移行

## 依頼内容（原文）

> 検索フィルタですが、NotionライクなUIにしてください。
> （確認の結果、移行漏れ・不完全な実装を発見）

## 現状の実装状況

### NotionFilter 移行済み（8ページ）
- owners（飼主）— ただしフィルタプロパティが空
- medical-records（カルテ）
- accounting（会計）
- examinations（検査管理）
- vaccinations（予防接種）
- hospitalization（入院）
- trimming（トリミング）
- inventory（在庫）

### 未移行（14ページ）
- estimates（見積一覧）
- master/Settings（マスタトップ）
- master/CageSettings
- master/TrimmingSettings
- master/MedicineSettings
- master/DiagnosisSettings
- master/StaffSettings
- master/HospitalizationSettings
- master/ServiceTypeSettings
- master/InterviewTemplateSettings
- master/ChiefComplaintSettings
- master/AnimalSpeciesSettings
- master/TreatmentPlanMaster
- hospital-settings/ClinicMasterSettings

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | 飼主一覧に種（犬/猫）フィルタプロパティを追加 | FE | FE-020 | - |
| 2 | 見積一覧を NotionFilter に移行 | FE | FE-021 | - |
| 3 | マスタ設定画面（13画面）を NotionFilter に移行 | FE | FE-022 | - |

## 影響範囲

### DB / Backend
- 変更なし

### Frontend
- `frontend/src/features/owners/routes/OwnersList.tsx` — FilterProperty 追加
- `frontend/src/features/estimates/routes/EstimateList.tsx` — NotionFilter 移行
- `frontend/src/features/master/routes/*.tsx` — 13ファイルを NotionFilter 移行
- `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx` — NotionFilter 移行

## 関連イシュー

- [FE-020: 飼主一覧 フィルタプロパティ追加](../frontend/issues/open/FE-020-owners-list-filter-properties.md)
- [FE-021: 見積一覧 NotionFilter 移行](../frontend/issues/open/FE-021-estimates-notion-filter-migration.md)
- [FE-022: マスタ設定画面 NotionFilter 移行](../frontend/issues/open/FE-022-master-settings-notion-filter-migration.md)
