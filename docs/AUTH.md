# 動物病院管理システム 認証・認可設計書

本ドキュメントは、システムの認証（Authentication）および認可（Authorization）の設計を定義します。
マルチクリニック対応のRBAC（Role-Based Access Control）モデルを採用し、ユーザー種別・職種・権限の3層構造で柔軟なアクセス制御を実現します。

**バージョン**: v4.0 | **最終更新**: 2026-03-26

---

## 目次

1. [設計方針](#設計方針)
2. [ユーザーモデル（3層構造）](#ユーザーモデル3層構造)
   - [2.1 ユーザー種別 (UserType)](#21-ユーザー種別-usertype)
   - [2.2 職種 (JobTitle)](#22-職種-jobtitle)
   - [2.3 権限グループ (PermissionGroup)](#23-権限グループ-permissiongroup)
3. [マルチクリニック設計](#マルチクリニック設計)
4. [認証フロー](#認証フロー)
5. [権限マトリクス](#権限マトリクス)
6. [DB設計](#db設計)
   - [6.1 ENUM型](#61-enum型)
   - [6.2 テーブル](#62-テーブル)
   - [6.3 既存テーブルへの影響](#63-既存テーブルへの影響)
   - [6.4 インデックス](#64-インデックス)
7. [RLSポリシー設計](#rlsポリシー設計)
8. [フロントエンド実装方針](#フロントエンド実装方針)
9. [既存システムとの統合](#既存システムとの統合)

---

## 設計方針

| 項目 | 方針 |
|------|------|
| **3層モデル** | ユーザー種別（システムレベル）→ 職種（表示・テンプレート用）→ 権限グループ（リソース×CRUD制御）の3層で分離 |
| **権限グループの複数保持** | 1ユーザーが複数の権限グループに所属可能（例: 一般 + 管理者） |
| **クリニックスコープ** | 権限はクリニック単位でスコープされる（運営管理者を除く） |
| **職種と権限の分離** | 職種（何者であるか）と権限（何ができるか）を明確に分離。職種は権限テンプレートの初期値として使用 |
| **最小権限の原則** | デフォルトでは最小限の権限のみ付与。必要に応じて権限を追加 |
| **認証方式** | メールアドレス + パスワードによるログイン。将来的にSSO/SAML対応を想定 |
| **セッション管理** | JWT（アクセストークン + リフレッシュトークン）ベース |
| **監査ログ** | ログイン・ログアウト・権限変更を記録（将来実装） |

---

## ユーザーモデル（3層構造）

```
┌─────────────────────────────────────────────────────────────────────┐
│  Layer 1: ユーザー種別 (UserType)                                    │
│  システムレベルのアクセス範囲を決定                                      │
│  ┌──────────────┬──────────────┬──────────────┐                     │
│  │ system_admin  │ clinic_admin │    staff     │                     │
│  │  運営管理者    │  医院管理者   │   スタッフ    │                     │
│  └──────────────┴──────────────┴──────────────┘                     │
├─────────────────────────────────────────────────────────────────────┤
│  Layer 2: 職種 (JobTitle)                                           │
│  業務上の肩書き・デフォルト権限テンプレートの決定                           │
│  ┌──────┬──────┬──────┬──────┬──────┐                               │
│  │ 医師  │ 看護師│トリマー│ 受付  │ 職員  │                               │
│  └──────┴──────┴──────┴──────┴──────┘                               │
├─────────────────────────────────────────────────────────────────────┤
│  Layer 3: 権限グループ (PermissionGroup)                              │
│  リソース×CRUDのグループベースアクセス制御（1ユーザー複数グループ可）          │
│  ┌──────────┬──────────┬──────────┬──────────┬──────────┐           │
│  │  管理者    │    執行    │    一般    │                                    │
│  └──────────┴──────────┴──────────┴──────────┴──────────┘           │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.1 ユーザー種別 (UserType)

システム全体におけるアクセス範囲を決定する最上位の区分。

| ユーザー種別 | 値 | 説明 | クリニックスコープ |
|---|---|---|---|
| **運営管理者** | `system_admin` | 全クリニックを横断的に管理。システム設定・クリニック作成・運営管理者の追加が可能 | 全クリニック |
| **医院管理者** | `clinic_admin` | 所属クリニック内の全機能にアクセス可能。スタッフの権限管理が可能 | 所属クリニックのみ |
| **スタッフ** | `staff` | 所属クリニック内で、付与された権限に基づきアクセス | 所属クリニックのみ |

**暗黙的権限:**
- `system_admin`: 全権限を暗黙的に保持（権限チェックをバイパス）
- `clinic_admin`: 所属クリニック内の全権限を暗黙的に保持
- `staff`: 明示的に付与された権限のみ

### 2.2 職種 (JobTitle)

業務上の肩書き。UI表示やシフト管理でのフィルタリング、権限テンプレートの初期値決定に使用。

| 職種 | 値 | 説明 | デフォルト権限テンプレート |
|---|---|---|---|
| **医師** | `veterinarian` | 獣医師。診療・カルテ・処方等の医療行為全般 | `medical`, `hospitalization` |
| **看護師** | `nurse` | 動物看護師。医師の補助、入院管理、バイタル記録 | `medical_read`, `hospitalization` |
| **トリマー** | `trimmer` | トリミング施術担当 | `trimming` |
| **受付** | `reception` | 受付業務。予約・会計・飼主/ペット情報管理 | `reception`, `billing` |
| **職員** | `general_staff` | 一般職員。在庫管理等のバックオフィス業務 | `inventory` |

> **注**: デフォルト権限テンプレートはアカウント作成時の初期値であり、作成後に個別に権限の追加・削除が可能。

> **既存 `StaffRole` との関係**: 既存の `STAFF_ROLE_VALUES` (`features/master/types`) の `manager` は `UserType.clinic_admin` に対応。`StaffRole` enum は `JobTitle` に移行し、`manager` を `general_staff` に置換する。マスタ設定のスタッフカテゴリ（`StaffSection`）では `JobTitle` を使用して表示。

### 2.3 権限グループ (PermissionGroup)

**グループベース RBAC（実装済み）**: `permission_groups` テーブルで定義されたグループを通じてアクセスを制御。1ユーザーが複数のグループに所属可能。

#### 権限モデル

各権限グループは **15リソース × CRUD（4操作）** の細粒度アクセス制御を持つ。

| リソース | キー | 説明 |
|---|---|---|
| ダッシュボード | `dashboard` | ダッシュボード |
| 飼主管理 | `owners` | 飼主・ペット情報 |
| 予約管理 | `reservations` | 予約登録・変更 |
| 電子カルテ | `medical-records` | カルテ作成・確定 |
| 入院管理 | `hospitalization` | 入院・退院処理 |
| トリミング | `trimming` | トリミング記録 |
| 診察管理 | `examinations` | 検査記録 |
| 会計 | `accounting` | 会計・入金処理 |
| 予防接種 | `vaccinations` | ワクチン記録 |
| 定期健診 | `checkups` | 健診記録 |
| 在庫管理 | `inventory` | 在庫品目管理 |
| 見積 | `estimates` | 見積書管理 |
| シフト管理 | `shifts` | シフト登録 |
| マスタ設定 | `master` | 全マスタ + **権限グループ管理** |
| 病院設定 | `hospital-settings` | クリニック基本情報 |

#### シードデータ：3つの権限グループ

| グループ名 | カラー | 主な用途 |
|---|---|---|
| **管理者** | `#EF4444` | 全15リソースフルアクセス（権限設定管理含む） |
| **執行** | `#6366F1` | 全15リソース閲覧 + ほぼ全創作・編集（権限設定管理含む） |
| **一般** | `#10B981` | 基本業務操作（医療・予約・トリミング等の作成・編集） |

> **`master` リソースへのアクセス = 権限グループ管理画面へのアクセスを含む。**
> 全グループ `master` / `hospital-settings` の閲覧（can_view）は可。作成・編集・削除は 管理者 と 執行 のみ。
> `system_admin` / `clinic_admin` は全リソースへの暗黙的フルアクセス（グループ不要）。

**ユーザーとグループの組み合わせ例:**

| ユーザー | ユーザー種別 | 権限グループ |
|---|---|---|
| 院長（clinic_admin） | `clinic_admin` | （暗黙的に全権限） |
| 院長補佐 | `staff` | 管理者 |
| 部長 | `staff` | 執行 |
| 勤務医 | `staff` | 一般 |
| 動物看護師 | `staff` | 一般 |
| 受付 | `staff` | 一般 |
| トリマー | `staff` | 一般 |

---

## マルチクリニック設計

### クリニック所属モデル

```
┌─────────────┐       ┌──────────────────────┐       ┌──────────────┐
│  UserAccount │ 1   N │ UserClinicMembership  │ N   1 │    Clinic     │
│              │───────│                       │───────│              │
│  id          │       │  user_id (FK)         │       │  id          │
│  email       │       │  clinic_id (FK)       │       │  name        │
│  user_type   │       │  is_main (BOOLEAN)    │       │  branch_name │
│  job_title   │       │  joined_at            │       │  ...         │
│  ...         │       └──────────────────────┘       └──────────────┘
└─────────────┘
         │
         │ N   N  （user_permission_groups 中間テーブル）
         ▼
┌─────────────────────┐       ┌──────────────────────────┐
│  PermissionGroup     │ 1   N │  PermissionGroupRule      │
│                      │───────│                           │
│  id                  │       │  group_id (FK)            │
│  clinic_id (FK)      │       │  resource (TEXT)          │
│  name                │       │  can_view (BOOL)          │
│  description         │       │  can_create (BOOL)        │
│  color               │       │  can_edit (BOOL)          │
└─────────────────────┘       │  can_delete (BOOL)        │
                               └──────────────────────────┘
```

### 所属ルール

| ルール | 説明 |
|---|---|
| **複数クリニック所属** | 1ユーザーは複数のクリニックに所属可能 |
| **メインクリニック** | 各ユーザーは必ず1つのメインクリニック (`is_main = true`) を持つ。ログイン後の初期表示クリニック |
| **メインクリニック一意制約** | 1ユーザーにつき `is_main = true` は1レコードのみ（部分一意インデックスで保証） |
| **権限のクリニックスコープ** | `permission_groups` はクリニック単位で管理。同一ユーザーがA院では管理者グループ、B院では一般グループのみ、という設定が可能 |
| **運営管理者の例外** | `system_admin` はクリニック所属に関係なく全クリニックにアクセス可能 |

### クリニック切替

- サイドナビバーのヘッダー部分に現在のクリニック名を表示
- 所属クリニックが複数ある場合、クリックでドロップダウンメニューを展開
- クリニック切替時:
  1. `currentClinicId` をコンテキストで更新
  2. 切替先クリニックでの権限セットを再ロード
  3. 現在のページを維持（権限がある場合）またはダッシュボードにリダイレクト（権限がない場合）
  4. サイドナビバーのメニュー項目を権限に基づいてフィルタリング

---

## 認証フロー

### ログインフロー

```
ユーザー                    フロントエンド                  バックエンド
  │                           │                             │
  │  メール + パスワード入力    │                             │
  │ ─────────────────────────>│                             │
  │                           │  POST /auth/login           │
  │                           │ ────────────────────────────>│
  │                           │                             │  認証検証
  │                           │                             │  JWT生成
  │                           │  { accessToken,             │
  │                           │    refreshToken,            │
  │                           │    user: {                  │
  │                           │      id, email, userType,   │
  │                           │      jobTitle,              │
  │                           │      mainClinicId,          │
  │                           │      clinics: [...],        │
  │                           │      permissions: {...}     │
  │                           │    }                        │
  │                           │  }                          │
  │                           │ <────────────────────────────│
  │                           │                             │
  │                           │  AuthContext にセット         │
  │                           │  currentClinicId =           │
  │                           │    mainClinicId             │
  │                           │                             │
  │  メインクリニックの         │                             │
  │  ダッシュボード (`/`) へ遷移 │                             │
  │ <─────────────────────────│                             │
```

### ログイン後の初期遷移

| 条件 | 遷移先 |
|---|---|
| 通常ログイン | メインクリニックのダッシュボード (`/`) = 当日の受付ページ |
| セッション復帰（リフレッシュ） | 前回表示していたページ |
| パスワードリセット後 | パスワード変更完了画面 → ダッシュボード |

### セッション管理

| 項目 | 仕様 |
|---|---|
| **アクセストークン** | JWT、有効期限15分 |
| **リフレッシュトークン** | Opaque token、有効期限7日、httpOnly Cookie |
| **トークンリフレッシュ** | アクセストークン期限切れ時に自動リフレッシュ（APIインターセプター） |
| **ログアウト** | リフレッシュトークンをサーバー側で無効化 + クライアントのトークンをクリア |
| **同時セッション** | 1ユーザーにつき最大3デバイスまで（超過時は最古セッションを無効化） |

---

## 権限マトリクス

### 権限グループ別リソースアクセス（実装済みシードデータ）

凡例: ✓=可、−=不可。操作列は View / Create / Edit / Delete の順。

| リソース | 管理者 | 執行 | 一般 |
|---|---|---|---|
| `dashboard` | ✓/−/−/− | ✓/−/−/− | ✓/−/−/− |
| `owners` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/✓/✓/− |
| `reservations` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/✓/✓/− |
| `medical-records` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/✓/✓/− |
| `hospitalization` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/✓/✓/− |
| `trimming` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/✓/✓/− |
| `examinations` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/✓/✓/− |
| `accounting` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `vaccinations` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/✓/✓/− |
| `checkups` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/−/−/− |
| `inventory` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `estimates` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `shifts` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/✓/✓/− |
| `master` | **✓/✓/✓/✓** | **✓/✓/✓/−** | ✓/−/−/− |
| `hospital-settings` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/−/−/− |

> **重要**: `master` リソース = マスタ設定ページ全般（権限グループ管理含む）。
> 全グループが `master` / `hospital-settings` を閲覧可（can_view=true）。
> `master` の作成・編集は管理者・執行が可能。`hospital-settings` の作成・編集・削除は管理者のみ。

### サイドナビバー表示制御

`usePermission(resource).canView` が true のリソースに応じてフィルタリング。

| メニュー項目 | 必要なリソース（`canView` が true） |
|---|---|
| ダッシュボード | `dashboard` |
| 予約管理 | `reservations` |
| 飼主管理 | `owners` |
| カルテ | `medical-records` |
| 入院管理 | `hospitalization` |
| トリミング | `trimming` |
| 検査管理 | `examinations` |
| 会計 | `accounting` |
| 予防接種 | `vaccinations` |
| 在庫管理 | `inventory` |
| シフト管理 | `shifts` |
| マスタ設定 | `master` |

### デモアカウント（シードデータ）

| メールアドレス | パスワード | ユーザー種別 | 権限グループ | 説明 |
|---|---|---|---|---|
| `admin@example.com` | `password` | `clinic_admin` | （暗黙的に全権限） | 医院管理者 |
| `manager@example.com` | `password` | `staff` | 管理者 | 院長補佐（全リソースフルアクセス） |
| `exec@example.com` | `password` | `staff` | 執行 | 部長（業務全般閲覧＋権限管理） |
| `vet@example.com` | `password` | `staff` | 一般 | 勤務医 |
| `nurse@example.com` | `password` | `staff` | 一般 | 動物看護師 |
| `reception@example.com` | `password` | `staff` | 一般 | 受付担当 |
| `trimmer@example.com` | `password` | `staff` | 一般 | トリマー |

---

## DB設計

### 6.1 新規ENUM型

```sql
-- ユーザー種別
CREATE TYPE user_type AS ENUM ('system_admin', 'clinic_admin', 'staff');

-- 職種（既存 staff_role を拡張置換）
-- 既存: CREATE TYPE staff_role AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'manager');
-- 変更: 'manager' → 'general_staff' に置換。'manager' は user_type.clinic_admin に移行
CREATE TYPE job_title AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'general_staff');

-- アカウントステータス
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');
```

> **廃止済み**: `permission_type` ENUM（`account_admin`, `medical` 等の10値）は**実装されていない**。
> 権限制御は `permission_groups` + `permission_group_rules` テーブルで行う（§6.2参照）。

### 6.2 新規テーブル

#### `clinics` — クリニック（`clinic_info` を拡張・複数院対応）

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGINT` | `PK DEFAULT -` | クリニックID |
| `name` | `TEXT` | `NOT NULL` | 医院名 |
| `branch_name` | `TEXT` | `DEFAULT ''` | 支院名 |
| `postal_code` | `TEXT` | `DEFAULT ''` | 郵便番号 |
| `address` | `TEXT` | `DEFAULT ''` | 住所 |
| `phone_number` | `TEXT` | `DEFAULT ''` | 電話番号 |
| `fax_number` | `TEXT` | `DEFAULT ''` | FAX番号 |
| `registration_number` | `TEXT` | `DEFAULT ''` | 開設届出番号 |
| `director_name` | `TEXT` | `DEFAULT ''` | 院長名 |
| `email` | `TEXT` | `DEFAULT ''` | メールアドレス |
| `website` | `TEXT` | `DEFAULT ''` | Webサイト |
| `logo_url` | `TEXT` | | ロゴURL |
| `is_active` | `BOOLEAN` | `DEFAULT true` | 有効フラグ |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 作成日時 |
| `updated_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 更新日時 |

```sql
CREATE TABLE clinics (
  id                  BIGINT PRIMARY KEY DEFAULT -,
  name                TEXT NOT NULL,
  branch_name         TEXT DEFAULT '',
  postal_code         TEXT DEFAULT '',
  address             TEXT DEFAULT '',
  phone_number        TEXT DEFAULT '',
  fax_number          TEXT DEFAULT '',
  registration_number TEXT DEFAULT '',
  director_name       TEXT DEFAULT '',
  email               TEXT DEFAULT '',
  website             TEXT DEFAULT '',
  logo_url            TEXT,
  is_active           BOOLEAN DEFAULT true,
  created_at          TIMESTAMPTZ DEFAULT now(),
  updated_at          TIMESTAMPTZ DEFAULT now()
);
```

> 既存の `clinic_info` テーブル（シングルトン）を `clinics` テーブルに移行。カラム構成は `ClinicInfo` フロントエンド型と同一。`is_active` カラムを追加して閉院済みクリニックをソフト無効化可能。

---

#### `user_accounts` — ユーザーアカウント

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGINT` | `PK DEFAULT -` | ユーザーID |
| `email` | `TEXT` | `NOT NULL UNIQUE` | メールアドレス（ログインID） |
| `display_name` | `TEXT` | `NOT NULL` | 表示名 |
| `display_name_kana` | `TEXT` | | 表示名カナ |
| `user_type` | `user_type` | `NOT NULL DEFAULT 'staff'` | ユーザー種別 |
| `job_title` | `job_title` | | 職種（`system_admin` は NULL 可） |
| `status` | `account_status` | `NOT NULL DEFAULT 'active'` | アカウントステータス |
| `avatar_url` | `TEXT` | | アバター画像URL |
| `staff_master_id` | `BIGINT` | `FK → master_items.id` | 既存スタッフマスタへの紐付け |
| `last_login_at` | `TIMESTAMPTZ` | | 最終ログイン日時 |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 作成日時 |
| `updated_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 更新日時 |

```sql
CREATE TABLE user_accounts (
  id               BIGINT PRIMARY KEY DEFAULT -,
  email            TEXT NOT NULL UNIQUE,
  display_name     TEXT NOT NULL,
  display_name_kana TEXT,
  user_type        user_type NOT NULL DEFAULT 'staff',
  job_title        job_title,
  status           account_status NOT NULL DEFAULT 'active',
  avatar_url       TEXT,
  staff_master_id  BIGINT REFERENCES master_items(id) ON DELETE SET NULL,
  last_login_at    TIMESTAMPTZ,
  created_at       TIMESTAMPTZ DEFAULT now(),
  updated_at       TIMESTAMPTZ DEFAULT now()
);
```

> **`staff_master_id`**: 既存の `master_items` (category='staff') との紐付け。シフト管理・カルテの担当医表示等で既存のスタッフマスタデータを参照。認証実装後、スタッフマスタは `user_accounts` から派生する形に段階的に移行。

---

#### `user_clinic_memberships` — ユーザー・クリニック所属

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGINT` | `PK DEFAULT -` | 所属ID |
| `user_id` | `BIGINT` | `FK → user_accounts.id NOT NULL` | ユーザーID |
| `clinic_id` | `BIGINT` | `FK → clinics.id NOT NULL` | クリニックID |
| `is_main` | `BOOLEAN` | `NOT NULL DEFAULT false` | メインクリニックフラグ |
| `joined_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 所属開始日時 |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 作成日時 |

```sql
CREATE TABLE user_clinic_memberships (
  id         BIGINT PRIMARY KEY DEFAULT -,
  user_id    BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  clinic_id  BIGINT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  is_main    BOOLEAN NOT NULL DEFAULT false,
  joined_at  TIMESTAMPTZ DEFAULT now(),
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, clinic_id)
);

-- 1ユーザーにつき is_main = true は1レコードのみ
CREATE UNIQUE INDEX idx_user_clinic_main
  ON user_clinic_memberships(user_id)
  WHERE is_main = true;
```

---

#### `permission_groups` — 権限グループ（クリニックスコープ）

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGSERIAL` | `PK` | グループID |
| `clinic_id` | `BIGINT` | `FK → clinics.id NOT NULL` | クリニックID |
| `name` | `TEXT` | `NOT NULL` | グループ名 |
| `description` | `TEXT` | `DEFAULT ''` | グループ説明 |
| `color` | `TEXT` | `DEFAULT '#6B7280'` | 表示カラー（HEX） |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 作成日時 |
| `updated_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 更新日時 |

```sql
CREATE TABLE permission_groups (
  id          BIGSERIAL PRIMARY KEY,
  clinic_id   BIGINT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  description TEXT DEFAULT '',
  color       TEXT DEFAULT '#6B7280',
  created_at  TIMESTAMPTZ DEFAULT now(),
  updated_at  TIMESTAMPTZ DEFAULT now()
);
```

---

#### `permission_group_rules` — グループ別リソース権限

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGSERIAL` | `PK` | ルールID |
| `group_id` | `BIGINT` | `FK → permission_groups.id NOT NULL` | グループID |
| `resource` | `TEXT` | `NOT NULL` | リソースキー（例: `medical-records`） |
| `can_view` | `BOOLEAN` | `NOT NULL DEFAULT false` | 閲覧権限 |
| `can_create` | `BOOLEAN` | `NOT NULL DEFAULT false` | 作成権限 |
| `can_edit` | `BOOLEAN` | `NOT NULL DEFAULT false` | 編集権限 |
| `can_delete` | `BOOLEAN` | `NOT NULL DEFAULT false` | 削除権限 |

```sql
CREATE TABLE permission_group_rules (
  id         BIGSERIAL PRIMARY KEY,
  group_id   BIGINT NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
  resource   TEXT NOT NULL,
  can_view   BOOLEAN NOT NULL DEFAULT false,
  can_create BOOLEAN NOT NULL DEFAULT false,
  can_edit   BOOLEAN NOT NULL DEFAULT false,
  can_delete BOOLEAN NOT NULL DEFAULT false,
  UNIQUE(group_id, resource)
);
```

---

#### `user_permission_groups` — ユーザー・グループ中間テーブル

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `user_id` | `BIGINT` | `FK → user_accounts.id NOT NULL` | ユーザーID |
| `group_id` | `BIGINT` | `FK → permission_groups.id NOT NULL` | グループID |

```sql
CREATE TABLE user_permission_groups (
  user_id  BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  group_id BIGINT NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, group_id)
);
```

> **廃止済み**: `user_permissions` テーブル（`permission_type` ENUM ベース）は実装されていない。
> 権限制御は上記3テーブルで行う。

---

### 6.3 既存テーブルへの影響

#### 全テーブルに `clinic_id` を追加

マルチクリニック対応のため、既存の全データテーブルに `clinic_id` カラムを追加。

```sql
-- 例: owners テーブル
ALTER TABLE owners ADD COLUMN clinic_id BIGINT NOT NULL REFERENCES clinics(id);

-- 例: medical_records テーブル
ALTER TABLE medical_records ADD COLUMN clinic_id BIGINT NOT NULL REFERENCES clinics(id);

-- （以下、全27テーブルに同様の ALTER TABLE を実施）
```

> **対象テーブル**: `owners`, `pets`, `medical_records`, `hospitalizations`, `reservation_appointments`, `trimming_records`, `accountings`, `treatment_items`, `vital_entries`, `examination_records`, `examination_record_items`, `vaccination_records`, `checkup_records`, `care_plan_items`, `daily_records`, `vital_records`, `care_log_records`, `staff_note_records`, `treatment_plans`, `accounting_items`, `payment_infos`, `master_items`, `master_item_inspections`, `inventory_items`, `shift_entries`, `trimming_record_options`, `clinic_info`（→ `clinics` に統合）

#### `master_items` (category='staff') との関係

| 現状 | 移行後 |
|---|---|
| スタッフは `master_items` テーブルの `category='staff'` レコードとして管理 | `user_accounts` テーブルが認証の主テーブル。`staff_master_id` で既存マスタと紐付け |
| `staff_role` enum でスタッフ職種を管理 | `job_title` enum に移行（`manager` → `general_staff` 置換） |
| シフト管理は `master_items.id` を `shift_entries.staff_id` として参照 | 段階的に `user_accounts.id` ベースに移行。移行期間中は `staff_master_id` 経由で互換性維持 |

#### `clinic_info` テーブルの移行

| 現状 | 移行後 |
|---|---|
| `clinic_info` テーブル（シングルトン、PK なし） | `clinics` テーブル（複数レコード対応、BIGINT PK） |
| フロントエンド: `features/clinic/api/store.ts` でインメモリ管理 | フロントエンド: `currentClinicId` コンテキストで選択中クリニックを管理 |

### 6.4 インデックス

```sql
-- ===== user_accounts =====
CREATE INDEX idx_user_accounts_email ON user_accounts(email);
CREATE INDEX idx_user_accounts_user_type ON user_accounts(user_type);
CREATE INDEX idx_user_accounts_status ON user_accounts(status);
CREATE INDEX idx_user_accounts_staff_master ON user_accounts(staff_master_id);

-- ===== user_clinic_memberships =====
-- UNIQUE(user_id, clinic_id) が複合インデックスを暗黙作成
CREATE INDEX idx_user_clinic_memberships_clinic ON user_clinic_memberships(clinic_id);

-- ===== permission_groups =====
CREATE INDEX idx_permission_groups_clinic ON permission_groups(clinic_id);

-- ===== permission_group_rules =====
-- UNIQUE(group_id, resource) が複合インデックスを暗黙作成

-- ===== user_permission_groups =====
-- PRIMARY KEY(user_id, group_id) が複合インデックスを暗黙作成
CREATE INDEX idx_user_permission_groups_group ON user_permission_groups(group_id);

-- ===== 既存テーブルの clinic_id インデックス =====
CREATE INDEX idx_owners_clinic ON owners(clinic_id);
CREATE INDEX idx_pets_clinic ON pets(clinic_id);
CREATE INDEX idx_medical_records_clinic ON medical_records(clinic_id);
CREATE INDEX idx_hospitalizations_clinic ON hospitalizations(clinic_id);
CREATE INDEX idx_reservation_appointments_clinic ON reservation_appointments(clinic_id);
CREATE INDEX idx_trimming_records_clinic ON trimming_records(clinic_id);
CREATE INDEX idx_accountings_clinic ON accountings(clinic_id);
CREATE INDEX idx_master_items_clinic ON master_items(clinic_id);
CREATE INDEX idx_inventory_items_clinic ON inventory_items(clinic_id);
CREATE INDEX idx_shift_entries_clinic ON shift_entries(clinic_id);
```

---

## RLSポリシー設計

ERD.md の設計を認証・認可モデルに基づき拡充。

### ヘルパー関数

```sql
-- 現在のユーザーIDを取得
CREATE OR REPLACE FUNCTION auth.current_user_id()
RETURNS BIGINT AS $$
  SELECT id FROM user_accounts WHERE id = auth.uid()
$$ LANGUAGE sql STABLE SECURITY DEFINER;

-- 現在のユーザー種別を取得
CREATE OR REPLACE FUNCTION auth.current_user_type()
RETURNS user_type AS $$
  SELECT user_type FROM user_accounts WHERE id = auth.uid()
$$ LANGUAGE sql STABLE SECURITY DEFINER;

-- 指定クリニック・リソース・操作の権限チェック
CREATE OR REPLACE FUNCTION auth.has_permission(
  p_clinic_id BIGINT,
  p_resource  TEXT,
  p_action    TEXT  -- 'view' | 'create' | 'edit' | 'delete'
)
RETURNS BOOLEAN AS $$
  SELECT EXISTS (
    SELECT 1 FROM user_accounts ua
    WHERE ua.id = auth.uid()
    AND (
      ua.user_type = 'system_admin'
      OR (ua.user_type = 'clinic_admin' AND EXISTS (
        SELECT 1 FROM user_clinic_memberships ucm
        WHERE ucm.user_id = ua.id AND ucm.clinic_id = p_clinic_id
      ))
      OR EXISTS (
        SELECT 1 FROM user_permission_groups upg
        JOIN permission_groups pg ON pg.id = upg.group_id
        JOIN permission_group_rules pgr ON pgr.group_id = pg.id
        WHERE upg.user_id = ua.id
          AND pg.clinic_id = p_clinic_id
          AND pgr.resource = p_resource
          AND CASE p_action
            WHEN 'view'   THEN pgr.can_view
            WHEN 'create' THEN pgr.can_create
            WHEN 'edit'   THEN pgr.can_edit
            WHEN 'delete' THEN pgr.can_delete
            ELSE false
          END
      )
    )
  )
$$ LANGUAGE sql STABLE SECURITY DEFINER;

-- 指定クリニックへの所属チェック
CREATE OR REPLACE FUNCTION auth.is_member_of(p_clinic_id BIGINT)
RETURNS BOOLEAN AS $$
  SELECT EXISTS (
    SELECT 1 FROM user_accounts ua
    WHERE ua.id = auth.uid()
    AND (
      ua.user_type = 'system_admin'
      OR EXISTS (
        SELECT 1 FROM user_clinic_memberships ucm
        WHERE ucm.user_id = ua.id AND ucm.clinic_id = p_clinic_id
      )
    )
  )
$$ LANGUAGE sql STABLE SECURITY DEFINER;
```

### ポリシー定義

```sql
-- === 基本ポリシー: クリニックスコープ READ ===

-- 認証済み + 所属クリニックのデータのみ閲覧可能
CREATE POLICY "clinic_scope_read" ON owners
  FOR SELECT TO authenticated
  USING (auth.is_member_of(clinic_id));

CREATE POLICY "clinic_scope_read" ON pets
  FOR SELECT TO authenticated
  USING (auth.is_member_of(clinic_id));

CREATE POLICY "clinic_scope_read" ON medical_records
  FOR SELECT TO authenticated
  USING (auth.is_member_of(clinic_id));

-- （全 clinic_id 付きテーブルに同様のポリシーを作成）

-- === 機能別 WRITE ポリシー ===

-- カルテ: medical-records リソースの create/edit 権限が必要
CREATE POLICY "medical_write" ON medical_records
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'medical-records', 'create'));

CREATE POLICY "medical_update" ON medical_records
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'medical-records', 'edit'));

-- 会計: accounting リソースの create/edit 権限が必要
CREATE POLICY "billing_write" ON accountings
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'accounting', 'create'));

CREATE POLICY "billing_update" ON accountings
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'accounting', 'edit'));

-- マスタ: master リソースの create/edit 権限が必要
CREATE POLICY "master_write" ON master_items
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'master', 'create'));

CREATE POLICY "master_update" ON master_items
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'master', 'edit'));

-- 入院: hospitalization リソースの create 権限が必要
CREATE POLICY "hospitalization_write" ON hospitalizations
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'hospitalization', 'create'));

-- トリミング: trimming リソースの create 権限が必要
CREATE POLICY "trimming_write" ON trimming_records
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'trimming', 'create'));

-- シフト: shifts リソースの create/edit 権限が必要
CREATE POLICY "shift_write" ON shift_entries
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'shifts', 'create'));

CREATE POLICY "shift_update" ON shift_entries
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'shifts', 'edit'));

-- 在庫: inventory リソースの create/edit 権限が必要
CREATE POLICY "inventory_write" ON inventory_items
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'inventory', 'create'));

CREATE POLICY "inventory_update" ON inventory_items
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'inventory', 'edit'));
```

---

## フロントエンド実装方針

### 実装済みファイル構成

```
features/auth/
├── api/
│   ├── login.ts              # ログインAPI（POST /v1/auth/login）
│   ├── logout.ts             # ログアウトAPI
│   └── refresh-token.ts      # トークンリフレッシュ
├── components/
│   └── LoginForm.tsx          # ログインフォーム（デモボタン付き）
├── hooks/
│   └── use-auth.tsx           # AuthContext + useAuth() フック（実装済み）
└── routes/
    └── Login.tsx              # ログインページ
```

### 型定義（実装済み）

```typescript
// ユーザー種別
export const USER_TYPE_VALUES = ["system_admin", "clinic_admin", "staff"] as const;
export type UserType = (typeof USER_TYPE_VALUES)[number];

// 職種
export const JOB_TITLE_VALUES = ["veterinarian", "nurse", "trimmer", "reception", "general_staff"] as const;
export type JobTitle = (typeof JOB_TITLE_VALUES)[number];

// リソース別CRUD権限（usePermission の戻り値）
export interface ResourcePermissions {
  canView: boolean;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

// 権限グループ（permission_groups テーブル対応）
export interface PermissionGroupRule {
  resource: string;
  can_view: boolean;
  can_create: boolean;
  can_edit: boolean;
  can_delete: boolean;
}

// ログインユーザー情報
export interface AuthUser {
  id: number;
  email: string;
  displayName: string;
  displayNameKana: string;
  userType: UserType;
  jobTitle: JobTitle | null;
  status: string;
  staffMasterId: number | null;
  mainClinicId: number;
  clinics: ClinicMembership[];
  permissionGroups: PermissionGroupWithRules[];
}

export interface PermissionGroupWithRules {
  id: number;
  name: string;
  color: string;
  rules: PermissionGroupRule[];
}

// 認証コンテキスト
export interface AuthContextValue {
  user: AuthUser | null;
  currentClinicId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  switchClinic: (clinicId: string) => void;
  hasPermission: (resource: string, action: "view" | "create" | "edit" | "delete") => boolean;
}
```

### 権限チェックパターン（実装済み）

```typescript
// use-auth.tsx 内の hasPermission 実装
function hasPermission(
  resource: string,
  action: "view" | "create" | "edit" | "delete"
): boolean {
  if (!user) return false;
  // system_admin / clinic_admin は全権限バイパス
  if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
  // staff: permission_groups の rules を走査
  return user.permissionGroups.some((group) =>
    group.rules.some((rule) => {
      if (rule.resource !== resource) return false;
      switch (action) {
        case "view":   return rule.can_view;
        case "create": return rule.can_create;
        case "edit":   return rule.can_edit;
        case "delete": return rule.can_delete;
      }
    })
  );
}
```

### `usePermission(resource)` フック（実装済み）

```typescript
// 使用方法
import { usePermission } from "@/features/auth/hooks/use-auth";

// リソース別 CRUD 権限を取得
const { canView, canCreate, canEdit, canDelete } = usePermission("medical-records");

// 閲覧権限チェック
if (!canView) return <AccessDenied />;

// 権限に応じてボタン表示/非表示
{canCreate ? <Button onClick={handleCreate}>新規登録</Button> : null}
{canDelete ? <Button onClick={handleDelete}>削除</Button> : null}
```

### 権限ガード（ルートレベル）

```typescript
// features/xxx/routes/XxxList.tsx
export function XxxList() {
  const { canView } = usePermission("xxx");

  if (!canView) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">アクセス権限がありません</p>
      </div>
    );
  }
  // ... 正常表示
}
```

### マスタ設定（`master` リソース）ガード

```typescript
// features/master/ 配下の全ページ
const { canView } = usePermission("master");
// canView = true: 管理者・執行・一般・clinic_admin・system_admin（全グループ表示可）
// can_create/can_edit/can_delete = true: 管理者・執行のみ
```

### ルート定義（実装済み）

```typescript
// app/router.tsx
{
  path: "/settings",
  children: [
    {
      path: "permission-groups",
      lazy: async () => {
        const { PermissionGroupSettings } = await import(
          "@/features/master/routes/PermissionGroupSettings"
        );
        return { Component: PermissionGroupSettings };
      },
    },
    // ...
  ],
}
```

### 実装済みルート

| ルート | コンポーネント | 必要権限 |
|---|---|---|
| `/login` | `Login` | 認証不要 |
| `/settings/permission-groups` | `PermissionGroupSettings` | `master.canView` |

---

## 既存システムとの統合

### 段階的移行計画

| フェーズ | 内容 | 影響範囲 |
|---|---|---|
| **Phase 1: 認証基盤** | `AuthProvider` + `LoginForm` + モック認証。既存機能への影響なし | `/features/auth/` 新規作成、`App.tsx` に `AuthProvider` ラップ |
| **Phase 2: クリニック切替** | `ClinicSwitcher` + `currentClinicId` コンテキスト。`Sidebar.tsx` にクリニック切替UI追加 | `Sidebar.tsx` 修正、`clinics` テーブル作成 |
| **Phase 3: 権限ガード** | `PermissionGate` + `ProtectedRoute`。サイドバーメニュー・ルートに権限チェック追加 | 全ルート定義の修正、`Sidebar.tsx` のメニュー項目に権限ガード追加 |
| **Phase 4: RLS 適用** | バックエンド接続後、RLSポリシーを段階的に適用 | DB層のみ。フロントエンドは権限ガードで既に制御済み |

### 既存コンポーネントへの影響

| コンポーネント | 変更内容 |
|---|---|
| `Sidebar.tsx` | クリニック切替UI追加、メニュー項目の `PermissionGate` ラップ |
| `App.tsx` | `AuthProvider` でラップ、`/login` ルート追加 |
| `routes.ts` | `ProtectedRoute` による権限ガード追加 |
| `PageLayout` | ログインユーザー情報表示（アバター・名前）の追加を検討 |
| `features/clinic/` | `clinic_info` → `clinics` テーブル移行、`currentClinicId` 連携 |
| `features/master/` (staff) | `StaffRole` の `manager` → `general_staff` 置換、`StaffSection` の職種表示更新 |
| `features/shifts/` | `ShiftStaffInfo` を `user_accounts` ベースに移行 |

### 実装済み状態

| 項目 | 値 |
|---|---|
| Feature数 | 16（`auth` 含む） |
| 認証関連テーブル | 6（`clinics`, `user_accounts`, `user_clinic_memberships`, `permission_groups`, `permission_group_rules`, `user_permission_groups`） |
| 権限グループ（シード） | 6グループ |
| デモアカウント（シード） | 7アカウント |
| 保護リソース | 15リソース |

---

## 備考

1. **`staff_role` enum の移行済み**: `staff_role` の `manager` は `user_type.clinic_admin` に対応。現在の `job_title` enum は `general_staff` を使用（`manager` は廃止）。

2. **パスワードハッシュ**: `user_accounts` テーブルに `password_hash TEXT` カラムあり（bcrypt）。シードデータは `$2a$10$...` 形式のハッシュ済みパスワードを使用。

3. **権限グループのクリニックスコープ**: `permission_groups.clinic_id` により、グループはクリニック単位で管理される。`user_permission_groups` に `clinic_id` はなく、グループ側で管理。

4. **`master` リソースの特殊性**: `master` リソースへの `canView` が権限グループ管理画面へのアクセス可否を決定する。`canCreate/canEdit` = グループ作成・編集可、`canDelete` = グループ削除可。

5. **`system_admin` / `clinic_admin` のバイパス**: `usePermission` / `hasPermission` は `userType` が `system_admin` または `clinic_admin` の場合、グループルールをチェックせず常に `true` を返す。

6. **WCAG AA 準拠**: ログインフォームは `FormFieldError` + `aria-describedby` パターンに準拠。パスワード入力は `type="password"` + 表示/非表示トグル。

7. **印刷機能との関係**: 帳票（領収書・処方箋等）にクリニック情報を表示する箇所は、`currentClinicId` に基づく `clinics` テーブルのデータを使用。

---

## 関連ドキュメント

| ドキュメント | パス | 関連箇所 |
|---|---|---|
| **仕様定義書** | `docs/SPECIFICATION.md` | Feature一覧、ルーティング構成、ロードマップ |
| **画面仕様書** | `docs/SCREENS.md` | 全ルートの画面仕様（`/login` §15 追加済み） |
| **ER図** | `docs/ERD.md` | エンティティ一覧（4エンティティ追加）、リレーション追加 |
| **ER図** | `docs/ERD.md` | ENUM型追加、テーブル追加、`clinic_id` カラム追加（v5.0 最新版） |
| **デザインシステム** | `docs/DESIGN_SYSTEM.md` | ログインページのUI仕様（§15 Login参照） |
