# FE-121: 医院設定 - 税率マスタ設定 UI

**Status**: Closed
**Priority**: High
**Affects**: features/hospital-settings
**Date Created**: 2026-03-25
**Related**: TASK-029, BE-059（先行必須）

## Summary

医院設定画面に「税率設定」セクションを追加する。
「通常課税（10%）」「軽減税率（8%）」の税率名称と値を表示・編集できる UI を実装する。

## 現状のコード

```typescript
// frontend/src/features/hospital-settings/routes/ClinicSettings.tsx
// 現在: 医院基本情報（名前・住所・電話等）の設定フォームが実装されている
// 税率設定セクションは未実装

// frontend/src/features/hospital-settings/api/clinics.ts
// 現在: UpdateClinicRequest に standard_tax_rate/reduced_tax_rate なし
// BE-059 完了後 models.ts に Clinic.standard_tax_rate/reduced_tax_rate が追加される
```

## 必要な変更

### 1. 型定義（types.ts）

```typescript
// frontend/src/features/hospital-settings/api/types.ts
// BE-059 + make codegen 後に models.ts から導出
import type { Clinic } from "@/types/generated/models";

// 既存の UpdateClinicRequest に追加
export interface UpdateClinicRequest {
  // ... 既存フィールド
  standard_tax_rate?: number; // 0.10 or 0.08
  reduced_tax_rate?: number;  // 0.10 or 0.08
}
```

### 2. 設定フォームへの税率セクション追加

```typescript
// frontend/src/features/hospital-settings/routes/ClinicSettings.tsx
// 既存の設定フォームの末尾に「税率設定」セクションを追加

// 表示内容（Figmaデザインなし - 既存の設定セクションのスタイルに合わせる）:
// ┌─────────────────────────────────────┐
// │ 税率設定                              │
// │                                     │
// │ 通常課税     [____10____] %          │
// │ 軽減税率     [_____8____] %          │
// │                      [保存]          │
// └─────────────────────────────────────┘
```

```typescript
// 実装パターン（hospital-settings の既存パターンに合わせる）
// useClinicSettingsForm.ts の既存フォームフックを拡張

// features/hospital-settings/hooks/use-clinic-settings-form.ts に追加
// standard_tax_rate, reduced_tax_rate のフィールドを追加
```

### 3. フォームフック拡張

```typescript
// frontend/src/features/hospital-settings/hooks/use-clinic-settings-form.ts

// 既存の useClinicSettingsForm に standard_tax_rate/reduced_tax_rate を追加
// useTransition でサブミット管理（useOwnerForm のパターンを参照）
const [isSavePending, startSaveTransition] = useTransition();

const handleSave = () => {
  startSaveTransition(async () => {
    await updateClinic({
      standard_tax_rate: formData.standard_tax_rate,
      reduced_tax_rate: formData.reduced_tax_rate,
    });
    // toast success
  });
};
```

## UI 操作フロー

1. ユーザーが医院設定画面を開く
2. 「税率設定」セクションに通常課税（10%）・軽減税率（8%）が表示される
3. 値を変更して「保存」をクリック
4. PATCH /v1/clinics/:id が呼ばれ、成功トースト表示

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし（直接ファイル import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] 型は `models.ts` から導出（Clinic.standard_tax_rate）

## 依存関係

- BE-059 が先に完了している必要がある（PATCH /v1/clinics/:id に standard_tax_rate/reduced_tax_rate が追加済みであること）
- `make codegen` で `models.ts` に `Clinic.standard_tax_rate`, `Clinic.reduced_tax_rate` が含まれていること

## 完了条件

- [ ] 医院設定画面に「税率設定」セクションが表示される
- [ ] 通常課税・軽減税率の値を変更・保存できる
- [ ] 保存後に画面を再読み込みしても値が保持されている
- [ ] `pnpm build` 型エラーなし
- [ ] `pnpm lint` エラーなし
