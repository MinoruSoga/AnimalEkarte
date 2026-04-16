# BUG-312: useMutation オブジェクトを useCallback deps に入れている — rerender-dependencies 違反

## 概要

複数のコンポーネントで `useMutation()` が返すオブジェクト全体を `useCallback` の deps 配列に入れており、
TanStack Query v5 では mutation オブジェクト参照がレンダーごとに再生成されるため、
useCallback が毎レンダーで再生成される。正しくは `.mutate` / `.mutateAsync` 関数を分割代入して deps に入れるべき。

## 違反ルール

`rerender-dependencies` — useCallback/useMemo の deps にオブジェクト参照を入れてはならない。primitive を抽出するか、stable な関数参照のみを使用する。

## 違反箇所

### 1. `ClinicalPlanSection.tsx`

**File:** `frontend/src/features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx:55`

```typescript
}, [canEdit, physicalExam, diagnosisCategoryId, diagnosisNameId, diagnosisDetails, treatmentPolicy, updateMutation]);
//                                                                                                    ↑ VIOLATION
```

**修正案:**
```typescript
const { mutate: updateRecord } = useUpdateMedicalRecord();
// ...
}, [canEdit, physicalExam, diagnosisCategoryId, diagnosisNameId, diagnosisDetails, treatmentPolicy, updateRecord]);
```

---

### 2. `MedicalRecordBillCheck.tsx`

**File:** `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx`

複数箇所:
- Line 58: `[canEdit, confirmMutation, userId]`
- Line 69: `[canEdit, returnMutation]`
- Line 100: `[canEdit, updateTreatmentMutation]`
- Line 107: `[canDelete, deleteMutation]`

---

### 3. `CheckupsTab.tsx`

**File:** `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx:259`

```typescript
[canDelete, deleteMutation]
```

---

### 4. `StaffSettings.tsx`

**File:** `frontend/src/features/master/routes/StaffSettings.tsx:647-666`

```typescript
const handleSaveGroups = useCallback(..., [setGroupsMutation]);
const handleSaveClinics = useCallback(..., [setClinicsMutation]);
const handleSaveExcludedServiceTypes = useCallback(..., [setExcludedMutation]);
```

---

## 修正方針

各箇所で `useMutation()` の結果を分割代入し、`mutate` または `mutateAsync` のみを deps に入れる。

```typescript
// Before
const updateMutation = useUpdateSomething();
const handleSave = useCallback(() => {
  updateMutation.mutate(data);
}, [updateMutation]);  // ← オブジェクト全体

// After
const { mutate: updateSomething } = useUpdateSomething();
const handleSave = useCallback(() => {
  updateSomething(data);
}, [updateSomething]);  // ← stable な関数参照のみ
```

## 優先度

**Low** — TanStack Query の mutation オブジェクトが stable かどうかは実装依存であり、
実際の動作に大きな影響はないが、規約準拠のために修正すべき。

## 影響範囲

| ファイル | 箇所数 |
|---------|--------|
| `medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx` | 1 |
| `medical-records/components/MedicalRecordBillCheck.tsx` | 4 |
| `medical-records/components/CheckupsTab/CheckupsTab.tsx` | 1 |
| `master/routes/StaffSettings.tsx` | 3 |
