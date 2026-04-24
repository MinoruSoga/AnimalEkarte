# FE-158: OwnerForm — canEdit=false でもフォームフィールドが disabled/readOnly でなく Enter キーで送信可能

## 概要

`/owners/:id/edit` の `OwnerForm.tsx` で、`canEdit=false` の閲覧のみユーザーに対して `SubmitButton`（保存ボタン）は非表示になっているが、**16 個のテキスト/セレクト入力フィールドが `disabled` / `readOnly` になっていない**。フォーム全体が `<form action={formAction}>` で囲まれているため、フィールドにフォーカスした状態で Enter キーを押すと `formAction` が発火し `PATCH /owners/:id` が呼ばれる。

## 影響範囲

- `frontend/src/features/owners/routes/OwnerForm.tsx`
- 権限: `can_edit = false` のユーザー（閲覧のみ）

## 現状の挙動（バグ）

`canEdit=false` のユーザーが `/owners/:id/edit` を開くと：
1. 「保存」ボタンは表示されない ✅
2. しかしオーナー名・フリガナ・メール・電話番号・郵便番号・住所など 16 フィールドすべてが入力可能 ❌
3. フィールドに入力 → Enter キー → `PATCH /api/v1/owners/:id` → 403 Forbidden ❌

## 根本原因

```tsx
// OwnerForm.tsx — SubmitButton は canEdit でガード済み ✅
{canEdit ? <SubmitButton>保存</SubmitButton> : null}

// しかし各 Input には disabled がない ❌
<Input
  name="last_name"
  value={formData.last_name}
  onChange={(e) => handleInputChange("last_name", e.target.value)}
  // disabled={!canEdit}  ← なし
/>
```

`<form action={formAction}>` 内のボタン以外の要素で Enter を押した場合、ブラウザは最初の `submit` ボタンを探す。`SubmitButton` が非表示でも DOM に存在する場合や、hidden submit ボタンがある場合は送信が発火する。また入力可能状態のフィールドは閲覧のみユーザーに誤った操作可能感を与える。

## 期待する挙動

`canEdit=false` の場合：
1. 全フィールドが `disabled` または `readOnly` になっている
2. Enter キーを押してもフォームが送信されない
3. フィールドの値を変更できない

## 修正方針

### 方針 A: 各 Input/Select に disabled={!canEdit} を追加（推奨）

```tsx
<Input
  name="last_name"
  value={formData.last_name}
  onChange={(e) => handleInputChange("last_name", e.target.value)}
  disabled={!canEdit}  // ← 追加
/>
```

全 16 フィールドに `disabled={!canEdit}` を追加する。

### 方針 B: フォーム全体を fieldset でラップ

```tsx
<fieldset disabled={!canEdit}>
  {/* 全フィールド */}
</fieldset>
```

HTML の `<fieldset disabled>` は内包するすべてのフォームコントロールを一括 disable にする。変更箇所が最小限で済む。

### 根本対策: useActionState フォームに onKeyDown ガード追加

```tsx
// Enter キーによる意図しない送信を防止
<form
  action={formAction}
  onKeyDown={(e) => {
    if (e.key === "Enter" && !canEdit) {
      e.preventDefault();
    }
  }}
>
```

方針 A または B と組み合わせることを推奨。

## 優先度

**HIGH** — `canEdit=false` のユーザーがすべてのオーナー情報フィールドを編集でき、Enter キーで PATCH が発火して 403 が返される。実際のデータ変更は 403 で防がれるが、UI が誤った操作可能感を与え、コンソールに 403 エラーが出力される。オーナー情報（個人情報）の編集 UI が開くため、GDPR/個人情報保護の観点からも問題がある。

## 関連ファイル

- `frontend/src/features/owners/routes/OwnerForm.tsx`
- 発見日: 2026-04-07（RBAC Phase 2 テスト中）
