# FE-001: NavigationBlocker 未適用フォーム（6件）

## 問題
以下のフォームに `NavigationBlocker` / `useUnsavedChanges` が未適用。
未保存のまま画面遷移するとデータが失われる。

## 対象フォーム
1. `features/medical-records/routes/MedicalRecordForm.tsx`
2. `features/estimates/routes/EstimateForm.tsx`
3. `features/vaccinations/routes/VaccinationForm.tsx`
4. `features/examinations/routes/ExaminationForm.tsx`
5. `features/inventory/routes/InventoryForm.tsx`
6. `features/hospitalization/routes/HospitalizationForm.tsx`

## 参照実装
`features/owners/routes/OwnerForm.tsx` — `useUnsavedChanges` + `<NavigationBlocker>` パターン
`features/trimming/routes/TrimmingForm.tsx` — 同上

## 修正方針
```tsx
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";

const { isDirty, markDirty, markClean } = useUnsavedChanges();
// ...
<NavigationBlocker when={isDirty} />
```

各フォームの "dirty" 条件はフォーム入力開始時点で `markDirty()`、
保存成功後に `markClean()` を呼ぶ。

## 優先度
HIGH（ユーザーデータ損失リスク）
