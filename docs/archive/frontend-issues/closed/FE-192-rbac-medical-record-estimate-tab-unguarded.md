# FE-192: カルテ見積書タブ — MedicalRecordEstimate に RBAC なし（FE-156 修正漏れ）

## 概要

`/medical-records/:id` の **見積書タブ**（`MedicalRecordEstimate` / `EstimateForm`）に
`canEdit` / `canCreate` ガードが一切なく、`canEdit=false` / `canCreate=false` ユーザーが
見積書の行追加・編集・削除を実行できる。

FE-156 のチケットには `MedicalRecordEstimate` が対象として記載されていたが、
実装時に修正されなかった（修正漏れ）。

## 再現手順

1. `canCreate=false` / `canEdit=false` ユーザーでログイン
2. `/medical-records/22` 等カルテ詳細へ移動
3. 「見積書」タブを開く

## 確認された問題（スクリーンショット確認済み）

| 問題 | 状態 |
|------|------|
| 「+ 行を追加」ボタン表示・操作可能 | ❌ `disabled: false` |
| 見積書件名テキストフィールド操作可能 | ❌ `disabled: false` |
| 値引額入力フィールド操作可能 | ❌ `disabled: false` |
| コメント・備考テキストエリア操作可能 | ❌ `disabled: false` |

## 根本原因

```tsx
// MedicalRecordEstimate.tsx — usePermission も canEdit も一切なし
// (grep: canEdit/canCreate/usePermission → No matches found)

// EstimateForm.tsx — 同様に RBAC ゼロ
// (grep: canEdit/canCreate/usePermission/disabled → No matches found)
```

## 修正方針

`MedicalRecordDiagnosisPlan` / `MedicalRecordBillCheck` と同パターンで実装。

```tsx
// MedicalRecordEstimate.tsx
const { canEdit, canCreate, canDelete } = usePermission("medical-records");

// EstimateForm に canEdit を渡す
<EstimateForm
  ...
  canEdit={canEdit}
  canCreate={canCreate}
/>

// EstimateForm.tsx
// 「+ 行を追加」ボタン
{canCreate ? <Button onClick={handleAddRow}>+ 行を追加</Button> : null}

// 各入力フィールド
<Input disabled={!canEdit} ... />  // 見積書件名
<Input disabled={!canEdit} ... />  // 値引額
<Textarea disabled={!canEdit} ... />  // コメント・備考
```

## 優先度

**HIGH** — 閲覧のみユーザーが見積書を作成・編集できる。
API 呼び出しは 403 で弾かれるが UI 上は操作可能な状態。

## 関連ファイル

- `frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx`
- `frontend/src/features/medical-records/components/EstimateForm.tsx`
- 発見日: 2026-04-08
- 関連: FE-156（本チケットの修正漏れ）
