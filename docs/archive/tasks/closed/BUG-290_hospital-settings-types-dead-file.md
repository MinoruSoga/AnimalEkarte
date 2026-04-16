# BUG-290: hospital-settings/api/types.ts — 未使用デッドファイル

## 概要

`frontend/src/features/hospital-settings/api/types.ts` は全41行、プロジェクト内のどのファイルからもインポートされていない完全なデッドファイルである。

## 問題

```typescript
// frontend/src/features/hospital-settings/api/types.ts
export type BackendClinic = Clinic;
export type BackendStaff = Staff;
export interface UpdateClinicRequest { ... }
export interface CreateStaffRequest { ... }
export interface UpdateStaffRequest { ... }
```

- `UpdateClinicRequest` は `hospital-settings/api/clinics.ts` で再定義済み
- `CreateStaffRequest` / `UpdateStaffRequest` は `master/api/staffs.ts` で定義済み
- `BackendClinic` / `BackendStaff` は `@/types/generated/models.ts` から直接インポート可能
- Grep結果: このファイルへのインポートは0件

## 影響

- デッドコードによる認知負荷増加
- 将来の開発者が誤って使用するリスク（実際に使われているAPIリクエスト型と乖離する可能性）

## 修正

```bash
# ファイルを削除する
rm frontend/src/features/hospital-settings/api/types.ts
```

## ステータス

- [x] ドキュメント作成
- [x] 実装完了（ファイル削除済み）
