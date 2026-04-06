# BUG-126: 詳細・編集ルート (/:id) にフロントエンド権限ガードがない

## 概要
`/owners/:id`、`/medical-records/:id` 等の詳細・編集ルートに `RequirePermission` ガードが
適用されていない。一覧ページ（`/owners`）には `RequirePermission resource={ResourceOwners}` が
あるが、その子ルートである `/:id` にはガードがない。

URL 直打ちで詳細ページにアクセスした場合、親ルートのガードを通過せずに表示される可能性がある。

## 脆弱性分類
- **CWE-862**: Missing Authorization (フロントエンド層)
- **影響**: 詳細ページの URL を知っていれば、権限がないユーザーでもアクセス可能な可能性

## 影響範囲

| パス | 親ルートのガード | `:id` ルートのガード | 問題 |
|------|-----------------|---------------------|------|
| `/owners/:id` | ✅ `ResourceOwners` view | ❌ なし | `:id` にガードなし |
| `/medical-records/:id` | ✅ `ResourceMedicalRecords` view | ❌ なし | `:id` にガードなし |
| `/hospitalization/:id` | ✅ `ResourceHospitalization` view | ❌ なし | `:id` にガードなし |
| `/trimming/:id` | ✅ `ResourceTrimming` view | ❌ なし | `:id` にガードなし |
| `/examinations/:id` | ✅ `ResourceExaminations` view | ❌ なし | `:id` にガードなし |
| `/accounting/:id` | ✅ `ResourceAccounting` view | ❌ なし | `:id` にガードなし |
| `/vaccinations/:id` | ✅ `ResourceVaccinations` view | ❌ なし | `:id` にガードなし |
| `/inventory/:id` | ✅ `ResourceInventory` view | ❌ なし | `:id` にガードなし |
| `/estimates/:id` | ✅ `ResourceEstimates` view | ❌ なし | `:id` にガードなし |

**注意**: React Router の nested route 構造では、`/owners/:id` は `/owners` の子ルートとして
定義されている場合、親の `RequirePermission` が自動的に適用される。
ただし、`/:id` が親の `children` 内ではなく独立したルートとして定義されている場合はガードが効かない。
**実際のルート構造を確認して、親ガードが `:id` にも適用されているか検証が必要。**

## 現状コード

### `frontend/src/app/router.tsx` — owners の例

```typescript
{
  path: "/owners",
  element: (
    <RequirePermission resource={ResourceOwners}>
      <Outlet />
    </RequirePermission>
  ),
  children: [
    { index: true, lazy: async () => { /* OwnersList */ } },
    {
      path: "new",
      element: (
        <RequirePermission resource={ResourceOwners} action="create">
          <Outlet />
        </RequirePermission>
      ),
      children: [{ index: true, lazy: async () => { /* OwnerFormPage */ } }],
    },
    {
      path: ":id",
      lazy: async () => { /* OwnerDetailPage */ },  // ❌ RequirePermission なし
    },
  ],
},
```

## 修正方針

`:id` ルートが親の `children` 内にある場合、親の `RequirePermission` が適用されるため
追加のガードは不要。ただし、編集ページには `action="edit"` ガードを追加すべき:

```typescript
{
  path: ":id",
  element: (
    <RequirePermission resource={ResourceOwners} action="edit">
      <Outlet />
    </RequirePermission>
  ),
  children: [{ index: true, lazy: async () => { /* OwnerDetailPage */ } }],
},
```

**ただし、詳細ページが「閲覧のみ」の場合は `action="view"` で十分（親ガードと同じ）。**
編集フォームを含む詳細ページの場合は `action="edit"` を追加すべき。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md`
> "Validate on both client and server"

フロントエンドでもルートレベルで一貫した認可チェックを適用すべき。

### プロジェクト内参照実装
- `/hospitalization/:id/edit` — `RequirePermission resource={ResourceHospitalization} action="edit"` が正しく適用されている
- `/estimates/:id/edit` — 同様に `action="edit"` ガードあり

## 優先度
**Medium** — 親ルートのガードが children に継承される場合は実害なし。
ただし、ルート構造の変更時に漏れるリスクがあるため、明示的なガード追加を推奨。

## 関連チケット
- BUG-121（修正済み）: `/settings` ルートガード
- BUG-125: バックエンド CRUD 粒度

## 関連ファイル
- `frontend/src/app/router.tsx` — 全ルート定義
- `frontend/src/components/shared/RequirePermission.tsx` — ガードコンポーネント
