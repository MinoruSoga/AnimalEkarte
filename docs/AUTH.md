# 動物病院管理システム 認証・認可設計書

本ドキュメントは、システムの認証（Authentication）および認可（Authorization）の設計を定義します。
マルチクリニック対応のRBAC（Role-Based Access Control）モデルを採用し、ユーザー種別・職種・権限の3層構造で柔軟なアクセス制御を実現します。

**バージョン**: v8.2 | **最終更新**: 2026-04-30

---

## ユーザーモデル（3層構造）

### 1. ユーザー種別 (UserType)
- **system_admin**: 全クリニックを横断的に管理。システム設定・クリニック作成が可能。
- **clinic_admin**: 所属クリニック内の全機能にアクセス可能。
- **staff**: 所属クリニック内で、付与された権限（PermissionGroup）に基づきアクセス。

### 2. 職種 (StaffType / Occupation)
- **staff_type ENUM**: `doctor` (獣医師), `nurse` (看護師), `resource` (設備・その他) の大分類。
- **occupations マスタ**: 表示用の詳細職種名（例: 院長, 主任看護師, トリマー）。

### 3. 権限グループ (PermissionGroup)
- **リソースベース認可**: 23種類のリソースに対して CRUD（View/Create/Edit/Delete）単位で権限を設定。
- **実効権限**: ユーザーが複数のグループに所属する場合、それらの権限の和集合（OR）が適用される。

---

## 認証・セッション管理

### dual-token 方式
- **Access Token**: JWT (HS256)、有効期限 15分。httpOnly Cookie (`access_token`) に格納。
- **Refresh Token**: Opaque token、有効期限 7日。httpOnly Cookie (`refresh_token`) に格納。
- **DB管理**: `refresh_tokens` テーブルにより、サーバー側でのセッション強制切断、ローテーション、盗難検知を実現。

### パスワード管理
- **ハッシュ化**: bcrypt (cost=10) を使用。
- **再設定フロー**: `POST /api/v1/auth/forgot-password` でトークン付きリンクをメール送信、`POST /api/v1/auth/reset-password` で新パスワードを設定。レート制限 3回/分。アカウント不在でも 200 を返す（メール存在漏洩防止）。
- **パスワード変更 (認証済み)**: `PUT /api/v1/users/me/password` で現在のパスワードを検証してから変更可能（BUG-148）。`ChangePasswordDialog` コンポーネントから呼び出す。

---

## 主要な認可リソース

| カテゴリ | リソース名 (Key) |
|---|---|
| **業務** | `reception`, `owners`, `reservations`, `medical-records`, `hospitalization`, `trimming`, `examinations`, `accounting`, `vaccinations`, `checkups`, `inventory`, `estimates`, `shifts`, `hospital-settings` |
| **マスタ** | `master-animal-species`, `master-medical`, `master-service-type`, `master-hospitalization`, `master-trimming`, `master-permission`, `master-staff`, `master-insurance`, `master-merchandise` |

---

## DB設計（認証関連）

- **accounts**: ログイン認証情報（Email, PasswordHash, システム管理者フラグ）。
- **staffs**: スタッフ名、職種、アカウント紐付け、LINE予約用プロフィール。
- **staff_clinic_assignments**: スタッフの所属医院管理（N:N）。
- **permission_groups**: 権限グループ定義。
- **permission_group_rules**: グループ別のリソース権限。
- **staff_permission_groups**: スタッフへのグループ割当（N:N）。
- **refresh_tokens**: セッション管理用。

---

## 実装状況（2026-04-30時点）

| 項目 | 状態 |
|---|---|
| JWT + httpOnly Cookie 認証 | ✅ 実装済み |
| Refresh Token ローテーション | ✅ 実装済み |
| パスワード再設定 (Forgot/Reset) | ✅ 実装済み |
| 自分のパスワード変更 (`PUT /api/v1/users/me/password`) | ✅ 実装済み |
| RBAC 認可ミドルウェア | ✅ 実装済み |
| フロントエンド権限ガード (`usePermission`) | ✅ 実装済み |
| マルチクリニック所属・切替 | ✅ 実装済み |
