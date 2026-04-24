# FE-180: カルテ内 MedicalRecordImage — 画像アップロード・削除の権限ガード欠落

## 概要

カルテ（MedicalRecordForm）内の `MedicalRecordImage` コンポーネントに `usePermission` が実装されていない。画像のアップロード（POST）と削除（DELETE）が権限チェックなしで実行できる。

## 影響範囲

| ファイル | 問題操作 | API 呼び出し | 深刻度 |
|---------|---------|------------|--------|
| `MedicalRecordImage.tsx` | 画像アップロード（行 47: `uploadMutation.mutate`）・画像削除（行 56: `deleteMutation.mutate`） | POST `/v1/medical-records/*/images`, DELETE `/v1/medical-records/*/images/:id` | HIGH |

## 根本原因

```tsx
// MedicalRecordImage.tsx — usePermission なし ❌
import { useUploadImages, useDeleteImage } from "...";

const uploadMutation = useUploadImages(resolvedId ?? "");
const deleteMutation = useDeleteImage(resolvedId ?? "");

// 行 47: canCreate チェックなし → POST /images ❌
uploadMutation.mutate(files);

// 行 56: canDelete チェックなし → DELETE /images/:id ❌
deleteMutation.mutate(imageId, {
  onSuccess: () => { ... }
});
```

アップロード・削除の両 callback が権限チェックなしで呼び出し可能であり、`canCreate=false` / `canDelete=false` ユーザーでも操作できる状態になっている。

## 修正方針

```tsx
import { usePermission } from "@/features/auth";

const { canCreate, canDelete } = usePermission("medical-records");

const handleUpload = useCallback((files: File[]) => {
  if (!canCreate) return;
  uploadMutation.mutate(files);
}, [canCreate, uploadMutation]);

const handleDelete = useCallback((imageId: string) => {
  if (!canDelete) return;
  deleteMutation.mutate(imageId);
}, [canDelete, deleteMutation]);

// アップロードトリガーを canCreate でガード
{canCreate ? (
  <ImageUploadButton onUpload={handleUpload} />
) : null}

// 削除ボタンを canDelete でガード（各画像オーバーレイ）
{canDelete ? (
  <DeleteOverlayButton onClick={() => handleDelete(image.id)} />
) : null}
```

## 優先度

**HIGH** — カルテ画像（X線・検査結果写真等）は医療記録として重要。`canDelete=false` ユーザーが削除できてしまうと医療記録の完全性が損なわれる。

## 関連ファイル

- `frontend/src/features/medical-records/components/MedicalRecordImage.tsx` (行 31-32: mutations, 行 47: uploadMutation.mutate, 行 56: deleteMutation.mutate)
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` (行 156: usePermission, 子コンポーネントに未渡し)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-156（カルテ画像タブのその他 RBAC 問題）、FE-177/178（カルテ内 usePermission 欠落パターン）
