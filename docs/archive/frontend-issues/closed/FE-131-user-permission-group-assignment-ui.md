# FE-131: ユーザー管理への権限グループ割当UI統合

**Status**: Closed
**Priority**: Medium
**Affects**: features/auth/ または features/master/ のユーザー管理UI（既存画面の拡張）
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-076（先に完了必要）, FE-128, FE-130

## Summary

ユーザー詳細・編集画面に「権限グループ」割当UIを追加する。
既存の `Permission[]` チェックボックスUIを削除し、
グループをマルチセレクトで割り当てる UIに置き換える。

## 現状のコード

ユーザー管理UIのパスを特定する必要がある。調査が必要。

```bash
# 実装前に以下で実際のファイルを確認すること
find frontend/src -name "*.tsx" | xargs grep -l "permission" 2>/dev/null
find frontend/src -name "*.tsx" | xargs grep -l "user_account\|UserAccount\|/users" 2>/dev/null
```

ユーザー管理関連の既存UIファイル（実装前に実在を確認）:
- `frontend/src/features/auth/routes/` または `features/master/routes/` 内のユーザー管理ページ

## 必要な変更

### 1. API hooks（新規: `features/auth/api/set-user-permission-groups.ts`）

```typescript
export async function setUserPermissionGroups(
  userId: string,
  groupIds: number[],
): Promise<void> {
  await axiosInstance.put(`/users/${userId}/permission-groups`, { group_ids: groupIds });
}
```

### 2. ユーザー詳細画面への権限グループ割当UI追加

**変更対象**: ユーザー編集フォーム内（実際のファイルは実装前に確認）

```typescript
// PermissionGroupSelector コンポーネント（新規）
interface PermissionGroupSelectorProps {
  availableGroups: PermissionGroup[];      // 医院全グループ
  selectedGroupIds: number[];              // 現在の割当
  onChange: (groupIds: number[]) => void;
}

export function PermissionGroupSelector({
  availableGroups,
  selectedGroupIds,
  onChange,
}: PermissionGroupSelectorProps) {
  return (
    <div className="flex flex-wrap gap-2">
      {availableGroups.map(group => {
        const isSelected = selectedGroupIds.includes(group.id);
        return (
          <button
            key={group.id}
            type="button"
            onClick={() =>
              onChange(
                isSelected
                  ? selectedGroupIds.filter(id => id !== group.id)
                  : [...selectedGroupIds, group.id],
              )
            }
            className={`px-3 py-1.5 rounded-full text-sm font-medium border transition-colors ${
              isSelected
                ? "bg-primary text-primary-foreground border-primary"
                : "bg-background text-foreground border-border hover:bg-muted"
            }`}
            style={{ borderColor: isSelected ? group.color : undefined, backgroundColor: isSelected ? group.color : undefined }}
          >
            {group.name}
          </button>
        );
      })}
      {availableGroups.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          権限グループが作成されていません。
        </p>
      ) : null}
    </div>
  );
}
```

### 3. ユーザー編集フォームの変更点

```typescript
// 削除: 旧 Permission[] チェックボックスUI
// (account_admin, medical, billing... のチェックボックス一覧)

// 追加: PermissionGroupSelector

// フォームの保存処理
const handleSave = () => {
  startSaveTransition(async () => {
    // 既存のユーザー更新 API 呼び出し
    await updateUser(userId, formData);
    // 権限グループ割当更新
    await setUserPermissionGroups(userId, selectedGroupIds);
    // React Query invalidate
    queryClient.invalidateQueries({ queryKey: ["users"] });
  });
};
```

### 4. user_type が system_admin / clinic_admin の場合の UI

`user_type` が `system_admin` または `clinic_admin` の場合は、
PermissionGroupSelector の代わりに「全権限（管理者）」と表示する。
グループ割当は不要（バックエンド側でバイパスされる）。

```typescript
{user.userType === "staff" ? (
  <PermissionGroupSelector
    availableGroups={groups}
    selectedGroupIds={selectedGroupIds}
    onChange={setSelectedGroupIds}
  />
) : (
  <p className="text-sm text-muted-foreground">
    管理者アカウントはすべての権限を持ちます。
  </p>
)}
```

## UI 操作フロー

1. ユーザー管理画面でユーザーを選択・編集モードに入る
2. 「権限グループ」セクションに利用可能なグループが表示される（色付きバッジ）
3. バッジをクリックしてグループを選択/解除（マルチセレクト）
4. 「保存」クリック → `PUT /users/:id/permission-groups` → 成功トースト

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] barrel index 経由 import なし

## 依存関係

- BE-076 が完了していること（PUT /users/:id/permission-groups API が存在）
- FE-130 が完了していること（権限グループ一覧APIが使用可能）
- FE-128 が完了していること（型定義が更新済み）

## ⚠️ 実装前確認事項

このイシューは「ユーザー管理UI」の既存ファイルを変更するが、
そのファイルパスが現時点で未確認。実装前に以下を必ず確認すること:

```bash
# 実際のユーザー管理UIファイルを特定
grep -rn "permission\|UserAccount\|/users" frontend/src --include="*.tsx" | grep -v "node_modules"
```

## 完了条件

- [ ] ユーザー編集画面に「権限グループ」セクションが表示される
- [ ] 医院に存在するグループが選択可能なバッジとして表示される
- [ ] グループの選択/解除ができる
- [ ] 保存で `PUT /users/:id/permission-groups` が呼ばれる
- [ ] 旧 `Permission[]` チェックボックスUIが削除されている
- [ ] `pnpm build` 型エラーなし、`pnpm lint` パス

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `frontend/src/features/master/api/user-accounts.ts` — 新規作成（API hooks: getUsers, getUser, setUserPermissionGroups）
  - `frontend/src/features/master/components/PermissionGroupSelector.tsx` — 新規作成
  - `frontend/src/features/master/routes/UserAccountSettings.tsx` — 新規作成
  - `frontend/src/app/router.tsx` — /settings/user-accounts ルート追加
- **備考**: ユーザー管理UIが未実装だったため、新規ページとして実装。/settings/user-accounts に配置。
