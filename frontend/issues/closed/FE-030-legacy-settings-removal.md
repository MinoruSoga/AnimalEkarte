# FE-030: レガシー Settings.tsx 廃止 + Insurance/JobTitle 専用ページ化

**Status**: Open
**Priority**: Low
**Affects**: master feature — Settings.tsx, InsuranceSettings, JobTitleSettings, InquiryTemplateSettings
**Date Created**: 2026-03-17
**Related**: TASK-007, FE-026, FE-027

## Summary

旧 `master_items` STI API を使用するレガシーテンプレート `Settings.tsx`（416行）を廃止し、Insurance と JobTitle を専用 API ページとして新規作成する。InquiryTemplateSettings は既に `InterviewTemplateSettings.tsx` が存在するため、ラッパーを削除して統合する。

## レガシーファイル一覧

| ファイル | 行数 | 状態 |
|---------|------|------|
| `routes/Settings.tsx` | 416 | 廃止対象 |
| `routes/InsuranceSettings.tsx` | 5 | `<Settings category="insurance" />` ラッパー |
| `routes/JobTitleSettings.tsx` | 5 | `<Settings category="job_title" />` ラッパー |
| `routes/InquiryTemplateSettings.tsx` | 5 | `<Settings category="inquiry_template" />` ラッパー（重複） |
| `hooks/use-master-items.ts` | ? | 旧 API hook |
| `api/get-master-items.ts` | ? | 旧 API |
| `api/create-master-item.ts` | ? | 旧 API |
| `api/update-master-item.ts` | ? | 旧 API |
| `api/delete-master-item.ts` | ? | 旧 API |

## 必要な変更

### 1. InsuranceSettings — 専用ページ新規作成

```typescript
// routes/InsuranceSettings.tsx（新規・専用API版）
// パターンA: useMasterCRUD + MasterListPage
// API: useGetInsurances, useCreateInsurance, useUpdateInsurance, useDeleteInsurance
// フィールド: name, is_active
```

Backend API が `GET /v1/masters/insurances` 等で存在するか確認が必要。

### 2. JobTitleSettings — 専用ページ新規作成

```typescript
// routes/JobTitleSettings.tsx（新規・専用API版）
// パターンA: useMasterCRUD + MasterListPage
// API: useGetJobTitles, useCreateJobTitle, useUpdateJobTitle, useDeleteJobTitle
// フィールド: name, is_active
```

### 3. InquiryTemplateSettings — InterviewTemplateSettings に統合

`InquiryTemplateSettings.tsx` は `<Settings category="inquiry_template" />` のラッパー。
`InterviewTemplateSettings.tsx` が同じ機能の専用API版として既に存在するため:
- `InquiryTemplateSettings.tsx` を削除
- router.tsx のルーティングを `InterviewTemplateSettings` に統合

### 4. レガシーファイル削除

- `routes/Settings.tsx`
- `hooks/use-master-items.ts`
- `api/get-master-items.ts`
- `api/create-master-item.ts`
- `api/update-master-item.ts`
- `api/delete-master-item.ts`
- `constants/category-config.ts`（Settings.tsx 専用なら削除、他で使われていれば残す）

## 依存関係

- **FE-026** と **FE-027** が必要（新規ページで useMasterCRUD + MasterListPage を使用）
- Backend の Insurance / JobTitle 専用 API が存在するか確認必要（なければ BE イシュー追加）

## 完了条件

- [ ] InsuranceSettings.tsx が専用 API + useMasterCRUD で動作
- [ ] JobTitleSettings.tsx が専用 API + useMasterCRUD で動作
- [ ] InquiryTemplateSettings.tsx 削除、InterviewTemplateSettings に統合
- [ ] Settings.tsx 削除
- [ ] use-master-items.ts + 旧 API ファイル 4件 削除
- [ ] router.tsx のルーティング更新
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
