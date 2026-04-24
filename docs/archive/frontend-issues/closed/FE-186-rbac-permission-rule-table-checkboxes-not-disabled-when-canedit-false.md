# FE-186: 権限ルールテーブル — canEdit=false でもチェックボックスが操作可能（PermissionRuleTable）

## 概要

`PermissionGroupSettings.tsx` で権限ルールを表示・編集する `PermissionRuleTable` コンポーネントのチェックボックス（canView/canCreate/canEdit/canDelete 各権限）が、`canEdit=false` のユーザーに対しても操作可能な状態で表示される。権限管理は最高権限の設定操作であり、誤って権限を変更されると全ユーザーのアクセス制御が崩壊する。

## 影響範囲

| ファイル | 問題 UI | 深刻度 |
|---------|---------|--------|
| `PermissionRuleTable.tsx` | 各リソースの canView/canCreate/canEdit/canDelete チェックボックス（行 113-142）が `canEdit=false` でも操作可能 | **CRITICAL** |

## 根本原因

```tsx
// PermissionGroupSettings.tsx 行 218 — canEdit は取得済み ✅
const { canEdit } = usePermission(ResourceMasterPermission);

// 行 204 — PermissionRuleTable を canEdit チェックなしでレンダリング ❌
<PermissionRuleTable
  rules={formRules}
  onRuleChange={handleRuleChange}  // canEdit チェックなしで渡す ❌
/>
```

```tsx
// PermissionRuleTable.tsx 行 113-142 — チェックボックスに disabled なし ❌
// canView チェックボックス
<Checkbox
  checked={rule.canView}
  onCheckedChange={() => onRuleChange(rule.resource, "canView", !rule.canView)}
  // disabled={!canEdit}  ← なし ❌
/>

// canCreate チェックボックス
<Checkbox
  checked={rule.canCreate}
  onCheckedChange={() => onRuleChange(rule.resource, "canCreate", !rule.canCreate)}
  // disabled={!canEdit}  ← なし ❌
/>
// ... canEdit, canDelete も同様 ❌
```

親の `handleRuleChange` がチェックボックス変更で状態を更新し、最終的に権限ルールの保存 API が呼ばれる。`canEdit=false` ユーザーは保存ボタン（行 328: `canEdit` でガード済み）を押せないため、実際に保存はされないが、フォーム状態が変更された状態でページを離脱しようとすると混乱を招く。

さらに `<form action={formAction}>` パターンが使われている場合、Enter キーで保存が試みられるリスクがある。

## 修正方針

### 方針 A: PermissionRuleTable に canEdit prop を追加

```tsx
// PermissionGroupSettings.tsx
<PermissionRuleTable
  rules={formRules}
  onRuleChange={handleRuleChange}
  canEdit={canEdit}  // ← 追加
/>
```

```tsx
// PermissionRuleTable.tsx
interface PermissionRuleTableProps {
  canEdit?: boolean;
  onRuleChange: (resource: string, field: string, value: boolean) => void;
}

<Checkbox
  checked={rule.canView}
  onCheckedChange={() => onRuleChange(rule.resource, "canView", !rule.canView)}
  disabled={!canEdit}  // ← 追加
/>
```

### 方針 B: fieldset disabled で一括 disable

```tsx
<fieldset disabled={!canEdit}>
  <PermissionRuleTable rules={formRules} onRuleChange={handleRuleChange} />
</fieldset>
```

## 優先度

**CRITICAL** — 権限管理（PermissionGroup）は全ユーザーのアクセス制御の基盤。`canEdit=false` のユーザーが誤ってチェックボックスを変更できる状態は、セキュリティ上の重大なリスクである（保存ボタンを押せなくても操作感が生まれる UX 問題）。

## 関連ファイル

- `frontend/src/features/master/components/PermissionRuleTable.tsx` (行 113-142: 各権限チェックボックス)
- `frontend/src/features/master/routes/PermissionGroupSettings.tsx` (行 218: usePermission, 行 204: PermissionRuleTable レンダリング, 行 328: 保存ボタン canEdit ガード)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-172（PermissionGroupSettings の canDelete 未取得）、FE-166（フォームフィールド disabled 欠落の系統的問題）
