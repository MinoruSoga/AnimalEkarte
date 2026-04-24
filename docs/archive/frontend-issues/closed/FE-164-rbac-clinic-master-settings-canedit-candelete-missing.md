# FE-164: 病院設定（ClinicMasterSettings）— canEdit・canDelete 未チェック（フォーム常時編集可・削除ボタン常時表示）

## 概要

`/settings/clinic` の `ClinicMasterSettings.tsx` で `usePermission` は呼び出しているが `canCreate` しか取得していない。`canEdit=false` でもフォームが全フィールド編集可能な状態で表示され、`canDelete=false` でも DeleteIconButton（行 369）が常に表示される。

## 現状の挙動（バグ）

```tsx
// ClinicMasterSettings.tsx 行 144 — canCreate のみ取得 ❌
const { canCreate } = usePermission(ResourceHospitalSettings);

// 行 369 — canDelete チェックなしで DeleteIconButton 常時表示 ❌
<DeleteIconButton onClick={() => setPendingDelete(selectedItem)} />

// 行 382-523 — フォームフィールドに disabled なし ❌
<Input name="name" value={formData.name} onChange={...} />
// → canEdit=false でも全フィールドが入力可能
```

`canCreate` でガード済みの操作:
- 「医院を追加」ボタン（行 297-302）✅
- リストへの新規追加（行 343-352）✅

未ガードの操作:
1. フォームの全フィールドが `canEdit=false` でも編集可能 ❌
2. DeleteIconButton が `canDelete=false` でも表示 ❌
3. フォーム送信（保存）が `canEdit=false` でも可能 ❌

## 修正方針

```tsx
// ClinicMasterSettings.tsx 行 144 を修正
const { canCreate, canEdit, canDelete } = usePermission(ResourceHospitalSettings);

// 削除ボタンをガード（行 369 付近）
{selectedItem && canDelete ? (
  <DeleteIconButton onClick={() => setPendingDelete(selectedItem)} />
) : null}

// フォームフィールドに disabled を追加（行 382-523）
<Input
  name="name"
  value={formData.name}
  onChange={...}
  disabled={!canEdit}  // ← 追加
/>

// または fieldset でラップ
<fieldset disabled={!canEdit}>
  {/* 全フォームフィールド */}
</fieldset>

// 保存ボタンを canEdit でガード
{canEdit ? <SubmitButton>保存</SubmitButton> : null}
```

## 優先度

**HIGH** — 病院の基本設定（院名・住所・電話番号等）を `canEdit=false` ユーザーが変更できる UI になっている。また `canDelete=false` でも医院削除ボタンが表示される。

## 関連ファイル

- `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx` (行 144, 369, 382-523)
- 発見日: 2026-04-07（RBAC Phase 2/3 テスト中）
