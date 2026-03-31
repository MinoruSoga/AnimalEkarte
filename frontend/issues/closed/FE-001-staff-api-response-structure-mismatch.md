# FE-001: スタッフAPI - レスポンス構造の不一致

**Status**: Open
**Priority**: High
**Affects**: Medical Record feature (doctor selection modal)
**Date Created**: 2026-03-16
**Related**: BE-007

## 問題

BE-007 で `/v1/masters/staffs` がページネーション対応から直接配列返却に変更されました。
しかし、FE側に **2つの異なるレスポンス構造を期待する実装** が存在し、これが矛盾しています。

## 現在の状態

### 医療記録内の実装（問題あり）
**ファイル**: `frontend/src/features/medical-records/api/get-staffs.ts`

```typescript
interface StaffListResponse {
  data: BackendStaff[];
  total: number;
  page: number;
  limit: number;
}

export function useGetStaffs() {
  return useQuery({
    queryFn: async () => {
      const { data } = await axios.get<StaffListResponse>("/v1/masters/staffs");
      return data.data;  // ← Pagination ラップを期待
    },
  });
}
```

### マスタ管理側の実装（正しい）
**ファイル**: `frontend/src/features/master/api/staffs.ts:84`

```typescript
export async function listStaffs(): Promise<Staff[]> {
  const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
  return data.map(transformStaff);  // ← 直接配列を期待
}
```

## 修正対応

### 方法: 医療記録内の get-staffs.ts を削除し、master/staffs.ts を再利用

医療記録内で独立した `useGetStaffs` を定義するのではなく、
マスタ管理の `useGetStaffs` を再利用するように修正。

1. `frontend/src/features/medical-records/api/get-staffs.ts` を削除
2. 医療記録内で使用している場所を以下に置き換え：
   ```typescript
   import { useGetStaffs } from "@/features/master/api/staffs";
   ```

## ブロッカー

BE-007 のAPI応答が直接配列（Pagination ラップなし）に変更されたため、
医療記録機能のドクター選択モーダルが機能していません。

## テスト環境

- 記録 ID: 医療記録作成フロー
- テスト日時: 2026-03-16 12:50 JST以降
