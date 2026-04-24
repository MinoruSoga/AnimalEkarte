# FE-220: master feature API 全 useMutation に onError がない（30+ フック）

## 概要

`frontend/src/features/master/api/` 配下のほぼすべての `useMutation` フックに
`onError` コールバックが設定されていない。
マスタデータの作成・更新・削除・並び替えが失敗してもユーザーに通知されない。

## 影響ファイル一覧

### `chief-complaint-categories.ts`
| フック | 行 |
|--------|---|
| `useCreateChiefComplaintCategory` | 98-106 |
| `useUpdateChiefComplaintCategory` | 108-117 |
| `useDeleteChiefComplaintCategory` | 119-127 |

### `medicines.ts`
| フック | 行 |
|--------|---|
| `useCreateMedicine` | 52-60 |
| `useUpdateMedicine` | 67-76 |
| `useDeleteMedicine` | 82-90 |
| `useReorderMedicines` | 96-104 |

### `staffs.ts`
| フック | 行 |
|--------|---|
| `useCreateStaff` | 113-121 |
| `useUpdateStaff` | 123-132 |
| `useDeleteStaff` | 134-142 |
| `useSetStaffPermissionGroups` | 195-215 |
| `useSetStaffClinics` | 264-284 |

### `consultations.ts`
| フック | 行 |
|--------|---|
| `useCreateConsultation` | 28-37 |
| `useUpdateConsultation` | 39-57 |
| `useDeleteConsultation` | 59-65 |
| `useReorderConsultations` | 67-74 |

### `vaccines-master.ts`
| フック | 行 |
|--------|---|
| `useCreateVaccineMaster` | 28-37 |
| `useUpdateVaccineMaster` | 39-54 |
| `useDeleteVaccineMaster` | 56-62 |
| `useReorderVaccinesMaster` | 64-71 |

### `checkup-types.ts`
| フック | 行 |
|--------|---|
| `useCreateCheckupType` | 28-37 |
| `useUpdateCheckupType` | 39-57 |
| `useDeleteCheckupType` | 59-65 |
| `useReorderCheckupTypes` | 67-74 |

### `procedures.ts`
| フック | 行 |
|--------|---|
| `useCreateProcedure` | 28-37 |
| `useUpdateProcedure` | 39-54 |
| `useDeleteProcedure` | 56-62 |
| `useReorderProcedures` | 64-71 |

### `diagnosis.ts`
| フック | 行 |
|--------|---|
| `useCreateDiagnosisCategory` | 169-177 |
| `useUpdateDiagnosisCategory` | 179-188 |
| `useDeleteDiagnosisCategory` | 190-198 |
| `useReorderDiagnosisCategories` | 200-208 |
| `useCreateDiagnosisName` | 223-231 |
| `useUpdateDiagnosisName` | 233-242 |
| `useDeleteDiagnosisName` | 244-252 |
| `useReorderDiagnosisNames` | 254-262 |

### `exam-types-master.ts`
| フック | 行 |
|--------|---|
| `useReplaceExamTypeItems` | 29-38 |

### `create-master-item.ts` / `update-master-item.ts` / `delete-master-item.ts`
| ファイル | フック | 行 |
|---------|--------|---|
| `create-master-item.ts` | `useCreateMasterItem` | 33-41 |
| `update-master-item.ts` | `useUpdateMasterItem` | 33-42 |
| `delete-master-item.ts` | `useDeleteMasterItem` | 16-24 |

## 修正方針

すべての `useMutation` に `onError` を追加する。コンテキスト名はマスタ種別に合わせる。

```ts
// Before
export const useCreateMedicine = () => {
  return useMutation({
    mutationFn: createMedicine,
    onSuccess: () => { ... },
    // onError なし
  });
};

// After
export const useCreateMedicine = () => {
  return useMutation({
    mutationFn: createMedicine,
    onSuccess: () => { ... },
    onError: (error) => handleApiError(error, "薬品マスタの作成"),
  });
};
```

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックおよび `onError` コールバックで `handleApiError(error, "コンテキスト")` を呼び出す。

## 注意

`use-master-save.ts` の catch ブロックで `handleApiError` 未使用の問題は FE-201 で別途管理。
本チケットは API フックレイヤーの `onError` 欠落のみを対象とする。

## 優先度
**High** — マスタデータ（スタッフ・薬品・診断・ワクチン等）の操作失敗がユーザーに通知されない。
特に `useDeleteStaff`・`useSetStaffPermissionGroups` はセキュリティ影響がある。

## 関連ファイル
- `frontend/src/features/master/api/chief-complaint-categories.ts`
- `frontend/src/features/master/api/medicines.ts`
- `frontend/src/features/master/api/staffs.ts`
- `frontend/src/features/master/api/consultations.ts`
- `frontend/src/features/master/api/vaccines-master.ts`
- `frontend/src/features/master/api/checkup-types.ts`
- `frontend/src/features/master/api/procedures.ts`
- `frontend/src/features/master/api/diagnosis.ts`
- `frontend/src/features/master/api/exam-types-master.ts`
- `frontend/src/features/master/api/create-master-item.ts`
- `frontend/src/features/master/api/update-master-item.ts`
- `frontend/src/features/master/api/delete-master-item.ts`
