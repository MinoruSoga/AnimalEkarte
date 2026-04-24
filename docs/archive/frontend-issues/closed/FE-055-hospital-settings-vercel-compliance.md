# FE-055: 病院設定 — react-hook-form 除去 + Vercel Best Practices 準拠

**Status**: Open
**Priority**: High
**Affects**: `features/hospital-settings/`
**Date Created**: 2026-03-18
**Related**: TASK-013

## Summary

`hospital-settings` feature から react-hook-form を除去し、プロジェクト標準の useTransition + useState パターンに置換する。
加えて、useCallback deps のオブジェクト参照問題と型定義の重複を解消する。

## 現状のコード

### 1. react-hook-form 使用（ClinicSettings.tsx）

```typescript
// frontend/src/features/hospital-settings/routes/ClinicSettings.tsx:3
import { useForm } from "react-hook-form";

// :39-51 — useForm 初期化
const { register, handleSubmit, reset, formState: { isDirty } } = useForm<ClinicFormData>({
  defaultValues: { ... }
});

// :53-67 — useEffect + reset（アンチパターン）
useEffect(() => {
  if (clinic) {
    reset({ ... });
  }
}, [clinic, reset]);

// :153 — mutation.isPending（useTransition ではない）
<Button type="submit" disabled={!isDirty || updateMutation.isPending}>
```

### 2. useCallback deps にオブジェクト（ClinicMasterSettings.tsx）

```typescript
// frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx:245
const handleSave = useCallback(() => {
  // ...
}, [formData, selectedItem, updateMutation, createMutation]);
// ❌ formData, selectedItem はオブジェクト → primitive 抽出が必要
```

### 3. 型定義の重複

```typescript
// api/clinics.ts:8-96 — 独自の BackendClinic, Clinic 型を定義
export interface BackendClinic { id: number; company_id: number; ... }
export interface Clinic { id: number; name: string; ... }

// api/types.ts:5-21 — generated/models.ts からインポート
import type { Clinic, Staff } from "@/types/generated/models";
export type BackendClinic = Clinic;

// ClinicSettings.tsx:21-31 — camelCase の ClinicFormData
interface ClinicFormData { name: string; postalCode: string; ... }

// ClinicMasterSettings.tsx:101-112 — snake_case の ClinicFormData（別定義）
interface ClinicFormData { name: string; postal_code: string; ... }
```

## 必要な変更

### 1. ClinicSettings.tsx: react-hook-form → useTransition + useState

**Before**:
```typescript
import { useForm } from "react-hook-form";

const { register, handleSubmit, reset, formState: { isDirty } } = useForm<ClinicFormData>({...});

useEffect(() => { if (clinic) reset({...}); }, [clinic, reset]);
```

**After**（owners パターン準拠）:
```typescript
// hooks/useClinicSettingsForm.ts を新規作成
import { useState, useCallback, useTransition } from "react";

export function useClinicSettingsForm(clinic: Clinic | undefined) {
  const [formData, setFormData] = useState<ClinicFormData>(() =>
    clinic ? transformToFormData(clinic) : defaultFormData
  );
  const [isSavePending, startSaveTransition] = useTransition();

  const handleInputChange = useCallback((field: keyof ClinicFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  }, []);

  const handleSave = useCallback((mutation: UseMutateFunction) => {
    startSaveTransition(async () => {
      await mutation(transformToRequest(formData));
    });
  }, [formData]);

  return { formData, handleInputChange, handleSave, isSavePending };
}
```

**ClinicSettings.tsx を簡素化**:
```typescript
export function ClinicSettings() {
  const { formData, handleInputChange, handleSave, isSavePending } = useClinicSettingsForm(clinic);

  return (
    <form onSubmit={(e) => { e.preventDefault(); handleSave(updateMutation.mutateAsync); }}>
      <Input value={formData.name} onChange={(e) => handleInputChange("name", e.target.value)} />
      <Button type="submit" disabled={isSavePending}>保存</Button>
    </form>
  );
}
```

### 2. ClinicMasterSettings.tsx: useCallback deps からオブジェクト除去

**Before**:
```typescript
const handleSave = useCallback(() => {
  // ...
}, [formData, selectedItem, updateMutation, createMutation]);
```

**After**:
```typescript
const selectedItemId = selectedItem?.id ?? null;
const formDataName = formData.name;
// ... 必要な primitive を抽出

const handleSave = useCallback(() => {
  // ...
}, [selectedItemId, formDataName, ...]);
```

### 3. 型定義統一

**api/clinics.ts から独自型定義を削除**:
- `BackendClinic`, `Clinic` インターフェースを削除
- `api/types.ts` から `generated/models.ts` 経由の型を使用
- `transformClinic()` は `api/transforms.ts` に移動（既存パターン準拠）

**ClinicFormData を統一**:
- `features/hospital-settings/types/index.ts` に1つの `ClinicFormData` を定義
- camelCase に統一（フロントエンド標準）

### 4. react-hook-form パッケージ依存確認

```bash
# 他の feature で使われていないか確認
docker compose exec frontend grep -rn "react-hook-form" src/
# hospital-settings のみなら package.json から削除可能
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）
- [ ] react-hook-form の import が0件

## 依存関係

- 依存なし（独立して着手可能）

## 完了条件

- [ ] react-hook-form が ClinicSettings.tsx から完全除去されている
- [ ] `hooks/useClinicSettingsForm.ts` が作成され、useTransition で pending 管理している
- [ ] ClinicMasterSettings.tsx の useCallback deps がプリミティブのみ
- [ ] 型定義が `types/index.ts` + `api/types.ts` に統一されている（重複なし）
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] 設定画面の保存・表示が正常動作
