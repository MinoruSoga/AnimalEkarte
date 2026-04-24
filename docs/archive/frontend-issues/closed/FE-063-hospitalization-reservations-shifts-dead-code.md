# FE-063: 入院・予約・シフト — 未使用フィルタ API 除去

**Status**: Open
**Priority**: Medium
**Affects**: `features/hospitalization/api/`, `features/reservations/api/`
**Date Created**: 2026-03-18
**Related**: TASK-014

## Summary

hospitalization から未使用フィルタ API 関数6件、reservations から4件を削除する。
shifts はデッドコードなし。

## 現状のコード

### 1. hospitalization — 6関数が未使用

```typescript
// frontend/src/features/hospitalization/api/get-hospitalization.ts
// 以下6関数が一度もインポートされていない:
// :49-65  getHospitalizationsByPetId()
// :49-65  useGetHospitalizationsByPetId()
// :68-84  getHospitalizationsByOwnerId()
// :68-84  useGetHospitalizationsByOwnerId()
// :87-102 getHospitalizationsByStatus()
// :87-102 useGetHospitalizationsByStatus()

// frontend/src/features/hospitalization/api/index.ts:10-15
// 上記6関数の barrel 再エクスポート（未使用）
```

### 2. reservations — 4関数が未使用

```typescript
// frontend/src/features/reservations/api/get-reservation.ts
// 以下4関数が一度もインポートされていない:
// :32-38  getReservationsByPetId()
// :41-46  useGetReservationsByPetId()
// :50-56  getReservationsByOwnerId()
// :59-65  useGetReservationsByOwnerId()

// frontend/src/features/reservations/api/index.ts:8-11
// 上記4関数の barrel 再エクスポート（未使用）
```

### 3. shifts — デッドコードなし ✅

全 API 関数が使用されていることを確認済み。

## 必要な変更

### 1. hospitalization

- `get-hospitalization.ts` から未使用6関数を削除
- `index.ts` から対応する再エクスポート行を削除

### 2. reservations

- `get-reservation.ts` から未使用4関数を削除
- `index.ts` から対応する再エクスポート行を削除

## 完了条件

- [ ] hospitalization: 6関数が `get-hospitalization.ts` から削除されている
- [ ] hospitalization: `index.ts` の対応再エクスポートが削除されている
- [ ] reservations: 4関数が `get-reservation.ts` から削除されている
- [ ] reservations: `index.ts` の対応再エクスポートが削除されている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
