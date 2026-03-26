# TASK-032: 権限グループ（RBAC）システムの実装

**作成日**: 2026-03-26
**ステータス**: Closed
**依頼元**: ユーザーによる口頭依頼

---

## 概要

動物病院スタッフ向けに権限グループ制度（RBAC）を実装する。
clinic_admin が自由にグループを作成し、ページ単位で view/create/edit/delete を設定できる。
ユーザーは複数グループに所属可能で、実効権限はグループのUNIONで解決する。

## 依頼内容（原文）

> 権限の機能を追加したいです。権限グループ的なのを作成し、権限毎にどのページへのアクセス権、編集権、削除権、登録権があるかなどを設定できればいいかなと思っています。例えば。受付権限のみ会計精算が可能など。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 権限グループはカスタム作成可能か？ | 自由に作成可能 |
| 2 | ユーザーは複数グループに所属できるか？ | はい |
| 3 | 既存のuser_permissionsとの関係 | 廃止して新システムに完全移行（ベストプラクティスを提案・採用） |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | DBマイグレーション（3テーブル追加・user_permissions廃止）| DB+BE | BE-073 | - | [x] |
| 2 | permission-groups CRUD API | BE | BE-074 | #1 | [x] |
| 3 | GET /me の実効権限レスポンス拡張 | BE | BE-075 | #1 | [x] |
| 4 | PUT /users/:id/permission-groups ユーザーグループ割当API | BE | BE-076 | #1, #2 | [x] |
| 5 | AuthContext型更新 + usePermission hook | FE | FE-128 | #3 | [x] |
| 6 | RequirePermission コンポーネント + ルートガード | FE | FE-129 | #5 | [x] |
| 7 | 権限グループ管理UI（/settings/permission-groups）| FE | FE-130 | #2, #4 | [x] |
| 8 | ユーザー管理へのグループ割当UI統合 | FE | FE-131 | #5, #7 | [x] |

## 受入条件（Acceptance Criteria）

- [x] AC-1: clinic_admin が「受付スタッフ」グループを作成し、accounting にのみ view/create を付与できる
- [x] AC-2: 「受付スタッフ」グループを複数ユーザーに割り当てられる
- [x] AC-3: 受付グループのみのユーザーで `/accounting` にアクセスできる。`/medical-records` はアクセス拒否になる
- [x] AC-4: accounting ページで create ボタンが表示され、edit/delete ボタンは非表示になる
- [x] AC-5: ユーザーが複数グループに所属する場合、いずれかのグループにある権限は有効になる（UNION）
- [x] AC-6: system_admin / clinic_admin はすべてのページにアクセス可能（グループ不要）
- [x] AC-7: グループを削除するとユーザーへの割当も消える（CASCADE）

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 既存user_permissionsの扱い | 廃止・完全移行 | 2システム並存で「どちらが正」か曖昧になる | 並存させる |
| 実効権限の取得方法 | /me レスポンスに含める | 追加APIコールが不要。既存の refreshToken() パターンを活用できる | GET /me/permissions を別エンドポイントにする |
| 権限の解決 | server-side UNION後にフラットマップで返す | フロントで GROUP JOIN ロジックを持たせない | フロントでグループを受け取り計算 |
| UIのグループ管理 | /settings/permission-groups に新規ページ | master 系ページと同じパターン | ユーザー管理に統合 |

## 影響範囲

### DB
- `user_permissions` テーブル: **廃止**
- `permission_type` ENUM: **廃止**
- `permission_groups` テーブル: **新規追加**
- `permission_group_rules` テーブル: **新規追加**
- `user_permission_groups` テーブル: **新規追加**

### Backend
- `backend/internal/model/clinic.go` — UserPermission/PermissionType 削除、3モデル追加 → make codegen
- `backend/internal/handler/auth_handler.go` — MeResponse.Permissions 型変更
- `backend/internal/handler/user_account_handler.go` — SetPermissions 削除、SetPermissionGroups 追加
- `backend/internal/handler/permission_group_handler.go` — **新規**
- `backend/internal/service/permission_group_service.go` — **新規**
- `backend/internal/repository/permission_group_repository.go` — **新規**
- `backend/cmd/api/main.go` — ルーティング追加

### Frontend
- `frontend/src/features/auth/types/index.ts` — Permission 型削除、EffectivePermissions 型追加
- `frontend/src/features/auth/hooks/use-auth.tsx` — hasPermission シグネチャ変更
- `frontend/src/features/auth/hooks/use-permission.ts` — **新規** (resource×action hook)
- `frontend/src/components/shared/RequirePermission.tsx` — **新規**
- `frontend/src/app/router.tsx` — ルートガード追加
- `frontend/src/features/master/routes/` — 権限グループ管理ページ追加
- `frontend/src/types/generated/models.ts` — make codegen で自動更新

## 参照実装

- `features/owners/` — API hooks・フォーム・リスト全パターン
- `backend/internal/handler/owner_handler.go` — handler実装パターン
- `backend/internal/service/user_account_service.go` — SetPermissions トランザクションパターン（→ SetPermissionGroups に流用）
- `features/master/routes/StaffSettings.tsx` — master系リスト画面パターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| hasPermission() の呼び出し箇所の全変更 | 高 | grep で全箇所洗い出し後に FE-128 で一括対応 |
| user_permissions 廃止により既存権限データが消える | 中 | DBリセット運用のため問題なし（新規シード不要） |
| /me レスポンス型変更でフロント既存コードが壊れる | 高 | BE-075 と FE-128 をセットで対応する |

## 実装順序

1. BE-073（DB + Model + make codegen）
2. BE-074（permission-groups CRUD API）
3. BE-075（/me レスポンス実効権限拡張）
4. BE-076（ユーザーグループ割当API）
5. FE-128（AuthContext型更新 + usePermission）
6. FE-129（RequirePermission + ルートガード）
7. FE-130（権限グループ管理UI）
8. FE-131（ユーザー管理グループ割当UI）

## 関連イシュー

- [BE-073](../../backend/issues/open/BE-073-permission-group-db-migration.md)
- [BE-074](../../backend/issues/open/BE-074-permission-group-crud-api.md)
- [BE-075](../../backend/issues/open/BE-075-me-endpoint-effective-permissions.md)
- [BE-076](../../backend/issues/open/BE-076-user-permission-group-assignment.md)
- [FE-128](../../frontend/issues/open/FE-128-auth-context-use-permission-hook.md)
- [FE-129](../../frontend/issues/open/FE-129-require-permission-route-guard.md)
- [FE-130](../../frontend/issues/open/FE-130-permission-group-management-ui.md)
- [FE-131](../../frontend/issues/open/FE-131-user-permission-group-assignment-ui.md)
