# BUG-127: select-pet ルートに create 権限ガードがない

## 概要
`/examinations/select-pet`、`/accounting/select-pet` 等の「ペット選択」ルートに
`RequirePermission action="create"` ガードが適用されていない。
`create` 権限がないユーザーが select-pet ステップまで進め、その後 `/new` でブロックされる。

UX として「ペット選択 → 新規作成ページでアクセス拒否」は不親切。
select-pet の時点でブロックすべき。

## 脆弱性分類
- **CWE-284**: Improper Access Control (UI層)
- **影響**: セキュリティ実害は軽微（`/new` ルートでブロック、API でも 403）。UX 問題。

## 再現手順
1. `vet@example.com`（RBAC検証用グループ: examinations create=F, accounting create=F）でログイン
2. ブラウザで `/examinations/select-pet` にアクセス
3. **結果**: ペット選択画面が表示される（本来は create 権限がないのでブロックすべき）

## ブラウザテスト結果

| ルート | create 権限 | AccessDenied? | 判定 |
|--------|-----------|-------------|------|
| `/examinations/select-pet` | create=**F** | ❌ 表示される | **❌ FAIL** |
| `/accounting/select-pet` | create=**F** | ❌ 表示される | **❌ FAIL** |
| `/hospitalization/select-pet` | view=**F** | ✅ ブロック | ✅ PASS（親ガード view=F で） |
| `/vaccinations/select-pet` | create=T | ❌ 表示される | ✅ PASS（正しく表示） |
| `/trimming/select-pet` | create=T | ❌ 表示される | ✅ PASS（正しく表示） |
| `/medical-records/select-pet` | create=T | ❌ 表示される | ✅ PASS（正しく表示） |

## 影響範囲

| ルート | 親ガード | select-pet ガード | `/new` ガード |
|--------|---------|------------------|--------------|
| `/examinations/select-pet` | `ResourceExaminations` view | ❌ なし | ✅ `action="create"` |
| `/accounting/select-pet` | `ResourceAccounting` view | ❌ なし | ✅ `action="create"` |
| `/vaccinations/select-pet` | `ResourceVaccinations` view | ❌ なし | ✅ `action="create"` |
| `/trimming/select-pet` | `ResourceTrimming` view | ❌ なし | ✅ `action="create"` |
| `/medical-records/select-pet` | `ResourceMedicalRecords` view | ❌ なし | ✅ `action="create"` |
| `/hospitalization/select-pet` | `ResourceHospitalization` view | ❌ なし | ✅ `action="create"` |

## 現状コード

### `frontend/src/app/router.tsx` — examinations の例

```typescript
{
  path: "/examinations",
  element: (
    <RequirePermission resource={ResourceExaminations}>
      <Outlet />
    </RequirePermission>
  ),
  children: [
    { index: true, lazy: async () => { /* ExaminationsList */ } },
    {
      path: "select-pet",
      lazy: async () => { /* SelectPetPage */ },  // ❌ create ガードなし
    },
    {
      path: "new",
      element: (
        <RequirePermission resource={ResourceExaminations} action="create">
          <Outlet />
        </RequirePermission>
      ),
      children: [{ index: true, lazy: async () => { /* ExaminationForm */ } }],
    },
  ],
},
```

## 修正方針

select-pet ルートに `action="create"` ガードを追加:

```typescript
{
  path: "select-pet",
  element: (
    <RequirePermission resource={ResourceExaminations} action="create">
      <Outlet />
    </RequirePermission>
  ),
  children: [{ index: true, lazy: async () => { /* SelectPetPage */ } }],
},
```

全6リソースの select-pet ルートに同じガードを追加する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/security.md`
> "Validate on both client and server"

select-pet はフォーム作成フローの一部。create 権限がない場合はフロー開始時点でブロックすべき。

### `.claude/CLAUDE.md` — Conditional Render
> 必ず `? (...) : null`

AccessDenied フォールバックは既存の `RequirePermission` コンポーネントで対応済み。

## 優先度
**Low** — セキュリティ実害なし（`/new` ルートと API で二重にブロック）。UX 改善。

## 関連チケット
- BUG-121（修正済み）: `/settings` ルートガード
- BUG-126: 詳細ルート `:id` のガード

## 関連ファイル
- `frontend/src/app/router.tsx` — 全 select-pet ルート（6箇所）
