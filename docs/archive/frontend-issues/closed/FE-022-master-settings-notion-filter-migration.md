# FE-022: マスタ設定画面（13画面） — NotionFilter 移行

**Status**: Open
**Priority**: Low
**Affects**: master feature, hospital-settings feature — マスタ設定画面
**Date Created**: 2026-03-17
**Related**: TASK-005

## Summary

マスタ設定画面13ファイルが SearchFilterBar のまま未移行。NotionFilter に一括置き換えする。マスタ設定はテキスト検索のみのため、FilterProperty は空配列でよい（「+ フィルタを追加」は非表示になるが、マスタ設定では追加フィルタが不要なため問題なし）。

## 対象ファイル（13ファイル）

| ファイル | 行番号 |
|---------|--------|
| `features/master/routes/Settings.tsx` | 214 |
| `features/master/routes/CageSettings.tsx` | 217 |
| `features/master/routes/TrimmingSettings.tsx` | 249, 421 |
| `features/master/routes/MedicineSettings.tsx` | 875 |
| `features/master/routes/DiagnosisSettings.tsx` | 270, 358 |
| `features/master/routes/StaffSettings.tsx` | 294 |
| `features/master/routes/HospitalizationSettings.tsx` | 294 |
| `features/master/routes/ServiceTypeSettings.tsx` | 278 |
| `features/master/routes/InterviewTemplateSettings.tsx` | 228 |
| `features/master/routes/ChiefComplaintSettings.tsx` | 207 |
| `features/master/routes/AnimalSpeciesSettings.tsx` | 208 |
| `features/master/routes/TreatmentPlanMaster.tsx` | 338 |
| `features/hospital-settings/routes/ClinicMasterSettings.tsx` | 279 |

## 必要な変更

各ファイルで以下の置換を行う:

```typescript
// Before:
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
<SearchFilterBar
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  placeholder="..."
  count={...}
/>

// After:
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
<NotionFilter
  properties={[]}
  activeFilters={[]}
  onFilterChange={() => {}}
  searchTerm={searchTerm}
  onSearchChange={setSearchTerm}
  searchPlaceholder="..."
  count={...}
/>
```

注意: TrimmingSettings と DiagnosisSettings は SearchFilterBar が2箇所ある。

## 完了条件

- [ ] 13ファイル全てで NotionFilter を使用
- [ ] SearchFilterBar の import が master/ から全削除
- [ ] テキスト検索が全画面で動作
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
