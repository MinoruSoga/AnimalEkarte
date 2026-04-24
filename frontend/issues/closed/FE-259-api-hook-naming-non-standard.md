# FE-259: API hook 命名規則の逸脱（useGet/useCreate/useUpdate/useDelete 以外の動詞）

**Status**: Open  
**Priority**: Low  
**Type**: Refactor  
**Date Created**: 2026-04-19  

## 背景

プロジェクト規約では API hook 名を `useGet*` / `useCreate*` / `useUpdate*` / `useDelete*` に統一している。
以下の hook がこの規約から逸脱している。

## 対象一覧

| ファイル | 現在の hook 名 | 規約準拠案 | 備考 |
|---------|--------------|-----------|------|
| `features/medical-records/api/medical-record-images.ts` | `useUploadImages` | `useCreateMedicalRecordImages` | CRUD 動詞に置換可能 |
| `features/medical-records/api/billing-confirmation.ts` | `useConfirmBillingConfirmation` | `useUpdateBillingConfirmation` | 状態遷移（確定）。ドメイン動詞の方が意図が明確な場合は維持も可 |
| `features/medical-records/api/billing-confirmation.ts` | `useReturnBillingConfirmation` | `useUpdateBillingConfirmation` | 状態遷移（差戻）。同上 |
| `features/master/api/staffs.ts` | `useSetStaffPermissionGroups` | `useUpdateStaffPermissionGroups` | Set → Update |
| `features/master/api/staffs.ts` | `useSetStaffClinics` | `useUpdateStaffClinics` | Set → Update |
| `features/master/api/staffs.ts` | `useSetStaffExcludedReservationTypes` | `useUpdateStaffExcludedReservationTypes` | Set → Update |

## 判断基準

- `useSet*` → `useUpdate*` に統一すべき（単純な語彙違い）
- `useUploadImages` → `useCreateMedicalRecordImages` に統一すべき
- `useConfirm*` / `useReturn*` は**状態機械のトランジション**であり、`useUpdate` では意図が薄れる。
  `useUpdateBillingConfirmationStatus` 等の明示名も検討余地あり。
  → チームで判断後、いずれかに統一すること

## 完了条件

- [ ] `useSet*` 系 3 hook を `useUpdate*` にリネーム、呼び出し箇所を更新
- [ ] `useUploadImages` → `useCreateMedicalRecordImages` にリネーム、呼び出し箇所を更新
- [ ] `useConfirm*` / `useReturn*` の命名方針をチームで決定し、必要に応じてリネーム
- [ ] lint / 型チェック / ビルドが通る
