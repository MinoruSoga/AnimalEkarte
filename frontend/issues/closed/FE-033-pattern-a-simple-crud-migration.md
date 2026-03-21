# FE-033: パターンA 6ページを MasterCRUDPage + useMasterSave に移行

**Status**: Open
**Priority**: High
**Affects**: JobTitle, ChiefComplaint, Insurance, InterviewTemplate, Staff, Hospitalization
**Date Created**: 2026-03-18
**Related**: TASK-008, FE-031, FE-032

## Summary

単純 CRUD パターンの6ページを `useMasterSave` + `MasterCRUDPage` に移行する。各ページは SidePanel コンポーネント + 設定定数 + 薄い組み立てコードのみに圧縮される。

## 対象ページと期待行数

| ページ | 現在 | 移行後（目安） | 削減率 |
|--------|------|--------------|--------|
| JobTitleSettings.tsx | 214行 | ~60行 | 72% |
| ChiefComplaintSettings.tsx | 211行 | ~60行 | 72% |
| InsuranceSettings.tsx | 252行 | ~80行 | 68% |
| InterviewTemplateSettings.tsx | 229行 | ~70行 | 69% |
| StaffSettings.tsx | 287行 | ~100行 | 65% |
| HospitalizationSettings.tsx | 299行 | ~100行 | 67% |
| **合計** | **1,492行** | **~470行** | **~68%** |

## 現状のコード

各ページの構造は同一:

```
1. import群（~20行）
2. COLUMNS 定数（~5行）
3. INPUT_CLASS 定数（1行）← 共有化で削除
4. FormData interface（~5行）
5. SidePanel コンポーネント（memo）（~40行）← 残す（各ページ固有）
6. handleSave コールバック（~30行）← useMasterSave で削除
7. MasterListPage + DataTable JSX（~50行）← MasterCRUDPage で圧縮
```

## 必要な変更

### 各ページの移行パターン

**Before（JobTitleSettings 214行）:**
```typescript
// 6つのセクション: imports, COLUMNS, FormData, SidePanel, handleSave, JSX
```

**After（~60行）:**
```typescript
// frontend/src/features/master/routes/JobTitleSettings.tsx

import { memo, useState } from "react";
import { Briefcase } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, LAYOUT } from "@/lib/design-tokens";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import { useGetAllJobTitles, useCreateJobTitle, useUpdateJobTitle, useDeleteJobTitle } from "@/features/master/api/job-titles";
import type { JobTitle, CreateJobTitleRequest, UpdateJobTitleRequest } from "@/features/master/api/job-titles";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─── FormData ───
interface JobTitleFormData { name: string; description: string; isActive: boolean; }

// ─── SidePanel（ページ固有のフォーム内容のみ） ───
const JobTitleSidePanel = memo(function JobTitleSidePanel({ item, onClose, onSave, onDeleteRequest }: {
  item: JobTitle | null; onClose: () => void; onSave: (d: JobTitleFormData) => void; onDeleteRequest: (i: JobTitle) => void;
}) {
  const [f, setF] = useState<JobTitleFormData>(() => ({
    name: item?.name ?? "", description: item?.description ?? "", isActive: item?.isActive ?? true,
  }));
  return (
    <MasterSidePanel isNew={!item} title={f.name} onTitleChange={(v) => setF((p) => ({ ...p, name: v }))}
      onClose={onClose} onSave={() => onSave(f)} onDelete={item ? () => onDeleteRequest(item) : undefined}
      icon={<Briefcase className={LAYOUT.pageIcon.innerIcon} />}>
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
      <PropertyRow label="説明">
        <PropInput value={f.description} onChange={(v) => setF((p) => ({ ...p, description: v }))} placeholder="説明を入力" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─── Page ───
export function JobTitleSettings() {
  const { data } = useGetAllJobTitles();
  const createMutation = useCreateJobTitle();
  const updateMutation = useUpdateJobTitle();
  const deleteMutation = useDeleteJobTitle();
  const crud = useMasterCRUD<JobTitle>({ data, deleteMutation, entityLabel: "役職" });
  const { handleSave } = useMasterSave({
    crud, createMutation, updateMutation,
    validate: (d: JobTitleFormData) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d): CreateJobTitleRequest => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
    toUpdateRequest: (d): UpdateJobTitleRequest => ({ name: d.name, description: d.description || undefined, is_active: d.isActive }),
  });

  return (
    <MasterCRUDPage title="役職マスタ" icon={<Briefcase className="size-5 text-[#37352F]" />} entityLabel="役職"
      searchPlaceholder="役職名で検索..." emptyMessage="役職が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-sm ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-sm ${C.text}`}>{item.description || "-"}</TableCell>
          <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="p-0 text-right"><RowActionButton onClick={() => onEdit(item)} /></TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <JobTitleSidePanel key={props.item?.id ?? "new"} {...props} />}
    />
  );
}
```

### ページ固有の注意点

#### StaffSettings
- `searchFilter` カスタム: name + staffRole ラベル検索
- バリデーション追加: 新規時は email（必須）+ password（8文字以上）
- SidePanel: 新規時のみ email/password フィールド表示（`item === null ? (...) : null`）

#### InsuranceSettings
- SidePanel: coverageRate（number input）、contactPhone（tel input）、description
- `toCreateRequest`/`toUpdateRequest` で `coverageRate` の Number 変換

#### HospitalizationSettings
- SidePanel: bodySize（Select）、billingUnit（Select）、price（MoneyInput）、description
- hoisted JSX: `BODY_SIZE_SELECT_ITEMS`, `BILLING_UNIT_SELECT_ITEMS`

#### InterviewTemplateSettings
- `searchFilter` カスタム: title + category 検索
- バリデーション追加: category も必須
- SidePanel: title ではなく `formData.title` を MasterSidePanel の title に渡す
- `deleteNameField`: `"title"` を指定（`name` ではない）

#### ChiefComplaintSettings
- SidePanel: description（textarea）

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useMasterSave` 経由）
- [ ] `useCallback` で安定化（`useMasterSave` 内部）
- [ ] `memo()` で SidePanel メモ化
- [ ] `useState(() => ...)` lazy init
- [ ] `INPUT_CLASS` → `MASTER_INPUT_CLASS` 共有定数使用
- [ ] Vercel Best Practices 全パターン準拠

## 依存関係

- FE-031（`useMasterSave` hook）が完了している必要がある
- FE-032（`MasterCRUDPage` コンポーネント）が完了している必要がある

## 完了条件

- [ ] 6ページすべて移行完了
- [ ] 各ページの `handleSave` ボイラープレート削除
- [ ] 各ページの `MasterListPage` props ボイラープレート削除
- [ ] `INPUT_CLASS` の各ページ定義を `MASTER_INPUT_CLASS` に置換
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
- [ ] UI が既存と同一の動作（見た目・操作フロー変更なし）
