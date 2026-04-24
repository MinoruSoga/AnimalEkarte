# FE-187: 編集ルートへの URL 直接アクセスで RequirePermission(action="update") が欠落

## 概要

`canEdit=false` の閲覧のみユーザーが、編集フォームの URL（`/owners/1`、`/inventory/6` 等）をアドレスバーに直接入力することで、全フォームフィールドが有効な編集画面にアクセスできる。新規作成ルートには `RequirePermission` が設定されているが、編集（update）ルートには設定されていない。

## 影響範囲

| ルート | 確認結果 | 深刻度 |
|--------|---------|--------|
| `/owners/:id` | 飼主・ペット編集フォーム全フィールド有効（17 フィールド中 16 が入力可能） | HIGH |
| `/inventory/:id` | 在庫編集フォーム全フィールド有効（品名・カテゴリ・在庫数・保管場所等） | HIGH |
| `/trimming/:id` | トリミング編集フォームが開く（ID が存在すれば全フィールド有効） | HIGH |
| `/medical-records/:id` | カルテ編集画面が開き、問診・診察・検査等全サブタブで編集 UI が操作可能 | HIGH |
| `/examinations/:id` | 検査編集フォームが開く（推定） | HIGH |
| `/hospitalization/:id` | 入院詳細・編集フォームが開く（推定） | HIGH |
| `/vaccinations/:id` | ワクチン編集フォームが開く（推定） | HIGH |
| `/accounting/:id` | 会計精算画面が開き、支払方法ボタン・お預かり金額が操作可能 | HIGH |

## 実際に確認した問題

### `/owners/1` へ直接アクセス
- 飼主・ペット編集フォームが完全に表示される
- 17 フィールド中 16 フィールドが enabled（飼主No のみ readOnly）
- 保存ボタンは非表示（canEdit ガードが機能）だが全データが編集可能状態
- RequirePermission が `action="create"` にしか設定されていない

### `/inventory/6` へ直接アクセス
- 在庫編集フォームが完全に表示される
- 品名・カテゴリ・在庫数・最低在庫数・保管場所・有効期限・仕入先・最終入荷日が全て enabled
- 保存ボタンは非表示だが全フィールド操作可能

## 根本原因

```tsx
// router.tsx — 新規作成は RequirePermission で保護されている ✅
{
  path: "new",
  element: (
    <RequirePermission resource={ResourceOwner} action="create">
      <OwnerFormPage />
    </RequirePermission>
  ),
},

// 編集ルートには RequirePermission がない ❌
{
  path: ":id",
  element: <OwnerFormPage />,
  // ↑ RequirePermission(action="update") が欠落
},
```

FE-159（行クリックで編集 UI が開く）はルート内 UI の問題だが、本イシューはルートガード自体が存在しない問題であり、URL 直接アクセスによる迂回が可能。

## 期待する挙動

`canEdit=false` の場合、`/owners/:id` に直接アクセスしたとき：
1. 「アクセス権限がありません」画面を表示する
2. または canView=true の場合は読み取り専用ビューを表示し全フィールドを disabled にする

## 修正方針

### 方針 A: RequirePermission(action="update") をルートに追加

```tsx
// router.tsx
{
  path: ":id",
  element: (
    <RequirePermission resource={ResourceOwner} action="update">
      <OwnerFormPage />
    </RequirePermission>
  ),
},
```

ただし canView=true ユーザーが参照できなくなる問題がある。

### 方針 B: 詳細参照を許可しフォームフィールドを readOnly に（readOnly パターン）

```tsx
// OwnerForm.tsx — canEdit を確認してフィールドを disabled にする
const { canEdit } = usePermission(ResourceOwner);

<Input
  value={formData.name}
  disabled={!canEdit}   // ← 追加
/>
```

canView=true, canEdit=false のユーザーは参照のみできる（FE-166 と同様の問題）。

**推奨**: 方針 B（readOnly ビュー）が UX として望ましい。方針 A と B を組み合わせ、canView=false のみブロックし、canEdit=false はフォームを disabled にする。

## 優先度

**HIGH** — 閲覧のみユーザーが URL を直接入力するだけで全ての編集フォームにアクセスでき、データ表示・フォーム操作が可能。保存ボタンは隠れているがデータが可視化される問題がある。FE-166（フォームフィールド disabled 欠落）と組み合わせると、ユーザーが自由にフィールドを変更できる状態が可視化される。

## 関連ファイル

- `frontend/src/app/router.tsx` — 各 feature の `:id` ルート定義
- `frontend/src/features/owners/routes/OwnerFormPage.tsx`
- `frontend/src/features/inventory/routes/InventoryForm.tsx`
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
- `frontend/src/components/shared/RequirePermission/RequirePermission.tsx`
- 発見日: 2026-04-08（RBAC Phase 3 テスト中）
- 関連: FE-159（行クリック編集 UI 開放）、FE-166（フォームフィールド disabled 欠落）、FE-185（canCreate/canEdit 混在）
