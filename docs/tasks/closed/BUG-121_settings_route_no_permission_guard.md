# BUG-121: /settings 配下のマスタページにフロントエンド権限ガードがない

## 概要
`/settings` 配下の全マスタ設定ページ（15+ルート）に `RequirePermission` ガードが適用されていない。
サイドバーでは `SidebarItemWithPermission` により権限チェックで非表示にしているが、
URL 直打ちで全マスタページにアクセス可能。

## 脆弱性分類
- **CWE-862**: Missing Authorization (フロントエンド層)
- **影響**: 認証済みユーザーが権限のないマスタ設定画面を閲覧可能

## 再現手順
1. `vet@example.com` / `password`（一般権限、`master-permission` = アクセス不可）でログイン
2. サイドバーでは「権限グループ」リンクが正しく非表示 (OK)
3. ブラウザで `http://localhost:3003/settings/permission-groups` に直接アクセス
4. **結果**: 権限グループ一覧が表示され、「新規登録」ボタン・「操作」ボタンまで見える

## 期待する動作
- 各マスタページに対応するリソースの `view` 権限がない場合、`AccessDenied` フォールバックを表示する
- `/settings` インデックスページでも、権限がないカードは非表示にする（→ BUG-123 で別途対応）

## 現状コード

### `frontend/src/app/router.tsx:626-751`

```typescript
// ── Settings（master） ─────────────────────────────────
{
  path: "/settings",
  element: <Outlet />,  // ❌ RequirePermission なし
  children: [
    { index: true, lazy: async () => { /* MasterSettingsIndex */ } },
    { path: "staff", lazy: async () => { /* StaffSettings */ } },
    { path: "treatment-items", lazy: async () => { /* TreatmentPlanMaster */ } },
    { path: "diagnosis", lazy: async () => { /* DiagnosisSettings */ } },
    { path: "animal-species", lazy: async () => { /* AnimalSpeciesSettings */ } },
    { path: "trimming", lazy: async () => { /* TrimmingSettings */ } },
    { path: "medicine", lazy: async () => { /* MedicineSettings */ } },
    { path: "service-type", lazy: async () => { /* ServiceTypeSettings */ } },
    { path: "hospitalization", lazy: async () => { /* HospitalizationSettings */ } },
    { path: "cage", lazy: async () => { /* CageSettings */ } },
    { path: "merchandise-items", lazy: async () => { /* MerchandiseItemSettings */ } },
    { path: "insurance", lazy: async () => { /* InsuranceSettings */ } },
    { path: "occupations", lazy: async () => { /* OccupationSettings */ } },
    { path: "permission-groups", lazy: async () => { /* PermissionGroupSettings */ } },
    { path: "inquiry-templates", lazy: async () => { /* InterviewTemplateSettings */ } },
    { path: "interview/chief-complaint", lazy: async () => { /* ChiefComplaintSettings */ } },
    { path: "interview/templates", lazy: async () => { /* InterviewTemplateSettings */ } },
  ],
},
```

### 比較: 他のルートの正しい実装（`router.tsx:55-71`）

```typescript
{
  path: "/",
  element: (
    <RequirePermission resource={ResourceDashboard}>
      <Outlet />
    </RequirePermission>
  ),
  children: [
    { index: true, lazy: async () => { /* Dashboard */ } },
  ],
},
```

## 影響範囲（全ルートとリソースマッピング）

| パス | リソース定数 | 一般権限の canView |
|------|-------------|-------------------|
| `/settings` (index) | — | — |
| `/settings/staff` | `ResourceMasterStaff` | true |
| `/settings/occupations` | `ResourceMasterStaff` | true |
| `/settings/treatment-items` | `ResourceMasterMedical` | true |
| `/settings/diagnosis` | `ResourceMasterMedical` | true |
| `/settings/medicine` | `ResourceMasterMedical` | true |
| `/settings/inquiry-templates` | `ResourceMasterMedical` | true |
| `/settings/service-type` | `ResourceMasterServiceType` | true |
| `/settings/hospitalization` | `ResourceMasterHosp` | true |
| `/settings/cage` | `ResourceMasterHosp` | true |
| `/settings/trimming` | `ResourceMasterTrim` | true |
| `/settings/merchandise-items` | `ResourceMasterMerchandise` | true |
| `/settings/insurance` | `ResourceMasterInsurance` | true |
| `/settings/permission-groups` | `ResourceMasterPermission` | **false** |
| `/settings/clinic` | `ResourceHospitalSettings` | true (view only) |
| `/settings/interview/chief-complaint` | `ResourceMasterMedical` | true |
| `/settings/interview/templates` | `ResourceMasterMedical` | true |

## 修正方針

### 1. Resource 定数の import 追加（`router.tsx:6-8`）

現在の import に以下を追加:
```typescript
import {
  // ... 既存の import
  ResourceMasterStaff,
  ResourceMasterMedical,
  ResourceMasterServiceType,
  ResourceMasterHosp,
  ResourceMasterTrim,
  ResourceMasterPermission,
  ResourceMasterInsurance,
  ResourceMasterMerchandise,
} from "@/types/generated/models";
```

### 2. 各ルートに RequirePermission を追加（`router.tsx:626-751`）

```typescript
{
  path: "/settings",
  element: <Outlet />,
  children: [
    { index: true, lazy: async () => { /* MasterSettingsIndex — ガード不要、BUG-123でカードフィルタ */ } },
    {
      path: "staff",
      element: <RequirePermission resource={ResourceMasterStaff}><Outlet /></RequirePermission>,
      children: [{ index: true, lazy: async () => { /* StaffSettings */ } }],
    },
    {
      path: "permission-groups",
      element: <RequirePermission resource={ResourceMasterPermission}><Outlet /></RequirePermission>,
      children: [{ index: true, lazy: async () => { /* PermissionGroupSettings */ } }],
    },
    // ... 他の全ルートも同様にラップ
  ],
},
```

### 3. RequirePermission コンポーネント（既存・変更不要）

`frontend/src/components/shared/RequirePermission.tsx` で実装済み:
- `hasPermission(resource, "view")` で判定
- 権限なしの場合 `AccessDenied` コンポーネントを表示（「アクセス権限がありません」メッセージ）
- `isSystemAdmin=true` は `useAuth().hasPermission()` 内部で常に `true` を返す

## 優先度
**High** — URL直打ちで権限バイパス可能。フロントエンドのみの問題だが、
バックエンド API の GET は認証のみで権限チェックなしのため（BUG-122）、実質的にデータ閲覧が可能。

## 関連チケット
- BUG-056（closed）: サイドバー RBAC リーク — 修正済み（`SidebarItemWithPermission`）
- BUG-122: バックエンド API 書き込み操作の権限チェック未適用
- BUG-123: `/settings` インデックスのマスタカード権限フィルタリング

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — フロントエンドベストプラクティス
- **Feature Indexing**: `RequirePermission` や Resource 定数は feature の `index.ts` 経由で import すること（Deep Import 禁止）
- **Conditional Render**: 権限チェック結果の条件レンダリングは `? (...) : null` を使用（`&&` 禁止）
- **Dependency Inversion**: `router.tsx`（app 層）で権限ガードを合成するのは正しいパターン。feature 内部に認可ロジックを持たせない

### `.claude/rules/security.md` — Input Validation
> "Validate on both client and server"

現状はサーバー側（バックエンド API）でも認可チェックが不完全（BUG-122）だが、
クライアント側（フロントエンド）でも防御層を持つべきという原則に違反している。
**Defense in Depth（多層防御）** の原則に基づき、フロントエンドでもルートガードを適用する。

### `.claude/rules/typescript-react.md` — React 19 Patterns
- `RequirePermission` は既存の共有コンポーネント（`components/shared/`）として実装済み
- ルートガードは `<RequirePermission resource={...}><Outlet /></RequirePermission>` パターンで統一
- Dashboard, Owners, Accounting 等の既存ルート（`router.tsx:55-625`）が参照実装

### 既存の正しい実装（プロジェクト内参照実装）
| ルート | 実装箇所 | パターン |
|--------|---------|---------|
| `/` (Dashboard) | `router.tsx:55-71` | `RequirePermission resource={ResourceDashboard}` |
| `/owners` | `router.tsx:73-130` | `RequirePermission resource={ResourceOwners}` + action="create" |
| `/accounting` | `router.tsx:337-390` | `RequirePermission resource={ResourceAccounting}` + action="create" |
| `/settings/clinic` | `router.tsx:754-770` | `RequirePermission resource={ResourceHospitalSettings}` |

**`/settings` 配下のみが、このプロジェクト標準パターンから逸脱している。**

## 関連ファイル
- `frontend/src/app/router.tsx:626-751` — 修正対象
- `frontend/src/components/shared/RequirePermission.tsx` — 権限ガードコンポーネント（既存）
- `frontend/src/features/auth/hooks/use-auth.tsx:104-114` — `hasPermission` 実装
- `frontend/src/types/generated/models.ts` — Resource 定数定義
