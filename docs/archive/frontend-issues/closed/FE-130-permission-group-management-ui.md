# FE-130: 権限グループ管理UI（/settings/permission-groups）

**Status**: Closed
**Priority**: High
**Affects**: features/master/（または新規 feature）, app/router.tsx
**Date Created**: 2026-03-26
**Related**: TASK-032, BE-074（先に完了必要）, FE-128

## Summary

clinic_admin が権限グループを管理するUI。グループ一覧・作成・編集・削除と、
ページ×CRUD のチェックボックステーブルでルール設定ができる画面を実装する。

## 現状のコード

**対象ファイルは存在しない（新規実装）。**

参照実装:
```
frontend/src/features/master/routes/StaffSettings.tsx  — リスト+フォームパターン
frontend/src/features/owners/                           — API hooks + フォームの全パターン
```

## 必要な変更

### 1. API hooks（新規: `frontend/src/features/master/api/permission-groups/`）

```typescript
// get-permission-groups.ts
import { axiosInstance } from "@/lib/axios";
import { useQuery } from "@tanstack/react-query";
import type { PermissionGroup } from "@/types/generated/models";

export async function getPermissionGroups(clinicId: string): Promise<PermissionGroup[]> {
  const { data } = await axiosInstance.get("/permission-groups", {
    params: { clinic_id: clinicId },
  });
  return data;
}

export function useGetPermissionGroups(clinicId: string) {
  return useQuery({
    queryKey: ["permission-groups", clinicId],
    queryFn: () => getPermissionGroups(clinicId),
    enabled: !!clinicId,
  });
}

// create-permission-group.ts
export async function createPermissionGroup(input: {
  clinicId: string;
  name: string;
  description?: string;
  color?: string;
}): Promise<PermissionGroup> {
  const { data } = await axiosInstance.post("/permission-groups", {
    name: input.name,
    description: input.description ?? "",
    color: input.color ?? "#6B7280",
  }, { params: { clinic_id: input.clinicId } });
  return data;
}

// set-permission-group-rules.ts
export interface RuleInput {
  resource: string;
  can_view: boolean;
  can_create: boolean;
  can_edit: boolean;
  can_delete: boolean;
}

export async function setPermissionGroupRules(groupId: string, rules: RuleInput[]): Promise<void> {
  await axiosInstance.put(`/permission-groups/${groupId}/rules`, { rules });
}
```

### 2. リソース定義（定数: `features/master/types/permission-resources.ts`）

```typescript
export interface ResourceDefinition {
  key: string;       // resource識別子（APIに送る値）
  label: string;     // 日本語ラベル
}

export const PERMISSION_RESOURCES: ResourceDefinition[] = [
  { key: "dashboard",        label: "ダッシュボード" },
  { key: "owners",           label: "オーナー管理" },
  { key: "reservations",     label: "予約" },
  { key: "medical-records",  label: "カルテ" },
  { key: "hospitalization",  label: "入院・ホテル" },
  { key: "trimming",         label: "トリミング" },
  { key: "examinations",     label: "検査" },
  { key: "accounting",       label: "会計" },
  { key: "vaccinations",     label: "ワクチン" },
  { key: "checkups",         label: "健診" },
  { key: "inventory",        label: "在庫" },
  { key: "estimates",        label: "見積" },
  { key: "shifts",           label: "シフト" },
  { key: "master",           label: "マスタ管理" },
  { key: "hospital-settings", label: "病院設定" },
] as const;
```

### 3. ルール編集テーブルコンポーネント（`PermissionRuleTable.tsx`）

```typescript
interface PermissionRuleTableProps {
  rules: RuleInput[];
  onChange: (rules: RuleInput[]) => void;
}

// 行: 各リソース（15行）× 列: view/create/edit/delete（4列）のチェックボックステーブル
// PERMISSION_RESOURCES をループして行を生成
// 各セルは <Checkbox> (shadcn/ui) で on/off

export const PermissionRuleTable = memo(function PermissionRuleTable({
  rules,
  onChange,
}: PermissionRuleTableProps) {
  const handleChange = useCallback(
    (resource: string, action: "can_view" | "can_create" | "can_edit" | "can_delete", value: boolean) => {
      onChange(
        rules.map(r =>
          r.resource === resource ? { ...r, [action]: value } : r
        )
      );
    },
    [rules, onChange],
  );

  return (
    <table className="w-full text-sm">
      <thead>
        <tr>
          <th className="text-left py-2 pr-4">ページ</th>
          <th className="text-center w-16">表示</th>
          <th className="text-center w-16">登録</th>
          <th className="text-center w-16">編集</th>
          <th className="text-center w-16">削除</th>
        </tr>
      </thead>
      <tbody>
        {PERMISSION_RESOURCES.map(res => {
          const rule = rules.find(r => r.resource === res.key) ?? {
            resource: res.key,
            can_view: false,
            can_create: false,
            can_edit: false,
            can_delete: false,
          };
          return (
            <tr key={res.key} className="border-t">
              <td className="py-2 pr-4">{res.label}</td>
              {(["can_view", "can_create", "can_edit", "can_delete"] as const).map(action => (
                <td key={action} className="text-center">
                  <Checkbox
                    checked={rule[action]}
                    onCheckedChange={v => handleChange(res.key, action, v === true)}
                  />
                </td>
              ))}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
});
```

### 4. グループ一覧・編集ページ（`PermissionGroupSettings.tsx`）

```typescript
// features/master/routes/PermissionGroupSettings.tsx

// 画面構成:
// - 上部: グループ一覧（カード形式）+ 「新規グループ作成」ボタン
// - 各カード: グループ名・色・説明、編集ボタン・削除ボタン
// - 「編集」クリック → 同ページ内でフォームを展開（または Sheet/Dialog）
// - フォーム内: 名前・説明・カラーピッカー + PermissionRuleTable

export function PermissionGroupSettings() {
  const { currentClinicId } = useAuth();
  const { data: groups = [], isLoading } = useGetPermissionGroups(currentClinicId ?? "");
  const [editingGroupId, setEditingGroupId] = useState<number | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  // 作成・編集・削除の mutation は useTransition で管理
  const [isSavePending, startSaveTransition] = useTransition();

  // ...
}
```

### 5. ルーティング追加（`app/router.tsx`）

```typescript
{
  path: "permission-groups",
  lazy: async () => {
    const { PermissionGroupSettings } = await import(
      "@/features/master/routes/PermissionGroupSettings"
    );
    return { Component: PermissionGroupSettings };
  },
},
```

既存の `/settings/*` グループ内に追加。

## UI 操作フロー

1. `/settings/permission-groups` を開くとグループ一覧が表示される
2. 「新規グループ作成」ボタンをクリック → フォームが展開
3. グループ名・色を入力 → PermissionRuleTable で各ページの view/create/edit/delete を設定
4. 「保存」クリック → POST /permission-groups → PUT /permission-groups/:id/rules
5. 既存グループの「編集」ボタン → フォームに現在の値を反映してルール編集
6. 「削除」ボタン → 確認ダイアログ → DELETE /permission-groups/:id

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] `memo()` で PermissionRuleTable をメモ化
- [ ] `useCallback` でハンドラを安定化

## 依存関係

- BE-074 が完了していること（permission-groups API が存在）
- FE-128 が完了していること（usePermission hook が使用可能）

## 完了条件

- [ ] `/settings/permission-groups` でグループ一覧が表示される
- [ ] グループ作成・編集・削除ができる
- [ ] PermissionRuleTable で15リソース×4アクションのチェックボックスが操作できる
- [ ] 保存後に画面が最新データに更新される（React Query invalidate）
- [ ] `pnpm build` 型エラーなし、`pnpm lint` パス

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**:
  - `frontend/src/features/master/api/permission-groups/permission-groups.ts` — 新規作成（API hooks）
  - `frontend/src/features/master/types/permission-resources.ts` — 新規作成（リソース定義）
  - `frontend/src/features/master/components/PermissionRuleTable.tsx` — 新規作成
  - `frontend/src/features/master/routes/PermissionGroupSettings.tsx` — 新規作成
  - `frontend/src/app/router.tsx` — permission-groups ルート有効化
