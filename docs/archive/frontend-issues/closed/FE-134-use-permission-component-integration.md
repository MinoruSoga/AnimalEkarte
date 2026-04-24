# FE-134: usePermission hook を各 feature コンポーネントに統合

**Status**: Open
**Priority**: Medium
**Affects**: 全 feature の routes/ および components/（usePermission 呼び出し箇所全体）
**Date Created**: 2026-03-29
**Related**: BUG-056, FE-133, TASK-048

---

## Summary

`usePermission` hook は定義されているが、各 feature コンポーネント内でほぼ使用されていない。
結果として **ルートガード（RequirePermission）は通過できても、ページ内のボタン・フォームが
全ユーザーに操作可能な状態**になっている。

例: `staff` ユーザーが `canCreate = false` であっても「新規登録」ボタンが表示される。

---

## 現状の問題

### `usePermission` は定義済みだが未使用

```typescript
// features/auth/hooks/use-permission.ts（定義済み）
export function usePermission(resource: string): {
  canView: boolean;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}
```

各 feature の routes/ や components/ で `usePermission` を使っている箇所がほぼない。
`RequirePermission` でページ単位のガードはあるが、**ボタン・フォームレベルの制御がない**。

### 具体的な漏れパターン

```tsx
// ❌ 現状: 権限チェックなしでボタンを表示
export function OwnersList() {
  return (
    <div>
      <Button onClick={handleCreate}>新規登録</Button>   {/* ← canCreate チェックなし */}
      <Button onClick={handleDelete}>削除</Button>       {/* ← canDelete チェックなし */}
    </div>
  );
}

// ❌ 現状: 権限チェックなしでフォームを表示
export function OwnerForm() {
  return (
    <form>
      <Button type="submit">保存</Button>               {/* ← canCreate/canEdit チェックなし */}
    </form>
  );
}
```

---

## 実装方針

### 基本パターン

```tsx
// ✅ 修正後: usePermission でアクション別に制御
export function OwnersList() {
  const { canCreate, canDelete } = usePermission("owners");

  return (
    <div>
      {canCreate ? <Button onClick={handleCreate}>新規登録</Button> : null}
      {/* 削除ボタンは行レベルで制御 → OwnerTableRow に委譲 */}
    </div>
  );
}

// 行レベルの制御
const OwnerTableRow = memo(function OwnerTableRow({ owner, onDelete }: Props) {
  const { canEdit, canDelete } = usePermission("owners");

  return (
    <tr>
      <td>{owner.name}</td>
      <td>
        {canEdit ? <Button onClick={() => handleEdit(owner.id)}>編集</Button> : null}
        {canDelete ? <Button onClick={() => onDelete(owner.id)}>削除</Button> : null}
      </td>
    </tr>
  );
});
```

### フォームの `canEdit` / `canCreate` 制御

新規作成フォームと編集フォームで使うアクションが異なる点に注意する。

```tsx
// ✅ 新規作成フォーム: canCreate を確認
export function OwnerCreateForm() {
  const { canCreate } = usePermission("owners");

  return (
    <form>
      {/* フィールド */}
      <Button type="submit" disabled={!canCreate}>保存</Button>
    </form>
  );
}

// ✅ 編集フォーム: canEdit を確認
export function OwnerEditForm({ owner }: { owner: Owner }) {
  const { canEdit } = usePermission("owners");

  return (
    <form>
      {/* フィールド（canEdit でない場合は読み取り専用にする） */}
      <input value={owner.name} readOnly={!canEdit} />
      {canEdit ? <Button type="submit">保存</Button> : null}
    </form>
  );
}
```

### 削除確認ダイアログの制御

```tsx
// ✅ 削除確認前に権限チェック
export function useOwnerActions(ownerId: string) {
  const { canDelete } = usePermission("owners");

  const handleDeleteRequest = useCallback(() => {
    if (!canDelete) return;  // 念のためガード
    setDeleteTarget(ownerId);
  }, [ownerId, canDelete]);

  return { handleDeleteRequest, canDelete };
}
```

---

## 対象 feature 一覧

以下 15 リソースの対応が必要。優先度は機密データを扱う順。

| feature | resource | 優先度 | 理由 |
|---------|---------|--------|------|
| `medical-records` | `medical-records` | 高 | 患者の診療情報（機密性最高） |
| `owners` | `owners` | 高 | 個人情報（連絡先・住所） |
| `accounting` | `accounting` | 高 | 金銭情報 |
| `hospitalization` | `hospitalization` | 高 | 入院患者情報 |
| `reservations` | `reservations` | 中 | 予約情報 |
| `examinations` | `examinations` | 中 | 診察情報 |
| `vaccinations` | `vaccinations` | 中 | ワクチン記録 |
| `master` | `master` | 中 | マスタデータ変更は全体に影響 |
| `hospital-settings` | `hospital-settings` | 中 | クリニック設定 |
| `trimming` | `trimming` | 低 | トリミング予約 |
| `inventory` | `inventory` | 低 | 在庫管理 |
| `estimates` | `estimates` | 低 | 見積 |
| `shifts` | `shifts` | 低 | シフト管理 |
| `checkups` | `checkups` | 低 | 健診 |
| `dashboard` | `dashboard` | 低 | ダッシュボード（read-only） |

---

## 実装時の注意点

### `canView` はルートガードで担保されている

`RequirePermission` コンポーネントがページレベルで `canView = false` の場合にリダイレクトする。
よって各コンポーネント内では `canCreate / canEdit / canDelete` の制御に集中すればよい。

### FE-132 完了後の型強化

FE-132 が完了すると `usePermission` の引数型が `string` → `Resource` に強化される。
本チケットの実装は FE-132 の前後どちらでも着手可能だが、FE-132 完了後に実施すると型エラーで
渡し間違いをコンパイル時に検出できる。

### `usePermission` は各コンポーネントで個別に呼ぶ

同一リソースに対して同じ hook を複数コンポーネントで呼んでも問題ない。
React Query（または Context）でキャッシュされるためパフォーマンス上の問題はない。

---

## 変更ファイル一覧（代表例）

| ファイル | 変更内容 |
|---------|---------|
| `features/owners/routes/OwnersList.tsx` | 新規登録ボタンに `canCreate` チェック追加 |
| `features/owners/components/OwnerTableRow.tsx`（または相当） | 編集・削除ボタンに `canEdit`/`canDelete` チェック追加 |
| `features/medical-records/routes/MedicalRecordsList.tsx` | 同上（優先度高） |
| 全 feature の一覧・詳細・フォームコンポーネント | `usePermission` 統合 |

---

## 受入条件

- [ ] `canCreate = false` の `staff` ユーザーに「新規登録」ボタンが表示されない（全 feature）
- [ ] `canEdit = false` の `staff` ユーザーに「編集」ボタン・フォームのサブミットが表示されない（全 feature）
- [ ] `canDelete = false` の `staff` ユーザーに「削除」ボタンが表示されない（全 feature）
- [ ] `clinic_admin` / `system_admin` は全操作ボタンが表示される
- [ ] `docker compose exec frontend pnpm build` 成功
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
- [ ] `docker compose exec frontend pnpm test:run` 成功
