# FE-061: 診察・ワクチン・カルテ — 未使用フィルタ API 除去

**Status**: Open
**Priority**: Medium
**Affects**: `features/examinations/api/`, `features/vaccinations/api/`, `features/medical-records/api/`
**Date Created**: 2026-03-18
**Related**: TASK-014

## Summary

3 feature から未使用の `getXxxByPetId/ByOwnerId/ByStatus` フィルタ関数と対応 React Query hook を削除する。
UI はメインリスト API + クライアントサイドフィルタで実装されており、これらは一度もインポートされていない。

## 現状のコード

### 1. examinations — 6関数が未使用

```typescript
// frontend/src/features/examinations/api/get-examination.ts
// 以下6関数が一度もインポートされていない:
// :28-35  getExaminationsByPetId()
// :37-43  useGetExaminationsByPetId()
// :46-53  getExaminationsByOwnerId()
// :55-61  useGetExaminationsByOwnerId()
// :64-71  getExaminationsByStatus()
// :73-79  useGetExaminationsByStatus()

// frontend/src/features/examinations/api/index.ts:7-15
// 上記6関数の barrel 再エクスポート（未使用）
```

### 2. vaccinations — 4関数が未使用

```typescript
// frontend/src/features/vaccinations/api/get-vaccination.ts
// 以下4関数が一度もインポートされていない:
// :21-28  getVaccinationsByPetId()
// :30-36  useGetVaccinationsByPetId()
// :39-46  getVaccinationsByOwnerId()
// :48-54  useGetVaccinationsByOwnerId()

// frontend/src/features/vaccinations/api/index.ts:8-12
// 上記4関数の barrel 再エクスポート（未使用）
```

### 3. medical-records — 4関数が未使用

```typescript
// frontend/src/features/medical-records/api/get-medical-record.ts
// 以下4関数が一度もインポートされていない:
// getMedicalRecordsByPetId()
// useGetMedicalRecordsByPetId()
// getMedicalRecordsByOwnerId()
// useGetMedicalRecordsByOwnerId()

// frontend/src/features/medical-records/api/index.ts:8-12
// 上記4関数の barrel 再エクスポート（未使用）
```

## 必要な変更

### 各 feature で同じパターン:

1. `get-xxx.ts` から未使用関数を削除（`getXxx()` と `useGetXxx()` の単体取得は残す）
2. `api/index.ts` から対応する再エクスポート行を削除

### 削除前の確認

各関数に対して `grep -rn "関数名" frontend/src/` を実行し、定義ファイルと index.ts 以外に参照がないことを確認する。

## 完了条件

- [ ] examinations: 6関数が `get-examination.ts` から削除されている
- [ ] examinations: `index.ts` の対応再エクスポートが削除されている
- [ ] vaccinations: 4関数が `get-vaccination.ts` から削除されている
- [ ] vaccinations: `index.ts` の対応再エクスポートが削除されている
- [ ] medical-records: 4関数が `get-medical-record.ts` から削除されている
- [ ] medical-records: `index.ts` の対応再エクスポートが削除されている
- [ ] `npm run build` パス
- [ ] `npm run lint` パス
