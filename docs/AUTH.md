# 動物病院管理システム 認証・認可設計書

本ドキュメントは、システムの認証（Authentication）および認可（Authorization）の設計を定義します。
マルチクリニック対応のRBAC（Role-Based Access Control）モデルを採用し、ユーザー種別・職種・権限の3層構造で柔軟なアクセス制御を実現します。

**バージョン**: v3.0 | **最終更新**: 2026-03-12

---

## 目次

1. [設計方針](#設計方針)
2. [ユーザーモデル（3層構造）](#ユーザーモデル3層構造)
   - [2.1 ユーザー種別 (UserType)](#21-ユーザー種別-usertype)
   - [2.2 職種 (JobTitle)](#22-職種-jobtitle)
   - [2.3 権限 (Permission)](#23-権限-permission)
3. [マルチクリニック設計](#マルチクリニック設計)
4. [認証フロー](#認証フロー)
5. [権限マトリクス](#権限マトリクス)
6. [DB設計](#db設計)
   - [6.1 新規ENUM型](#61-新規enum型)
   - [6.2 新規テーブル](#62-新規テーブル)
   - [6.3 既存テーブルへの影響](#63-既存テーブルへの影響)
   - [6.4 インデックス](#64-インデックス)
7. [RLSポリシー設計](#rlsポリシー設計)
8. [フロントエンド実装方針](#フロントエンド実装方針)
9. [既存システムとの統合](#既存システムとの統合)

---

## 設計方針

| 項目 | 方針 |
|------|------|
| **3層モデル** | ユーザー種別（システムレベル）→ 職種（表示・テンプレート用）→ 権限（機能アクセス制御）の3層で分離 |
| **権限の複数保持** | 1ユーザーが複数の権限を同時に保持可能（例: 医師権限 + アカウント管理者権限） |
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
│  Layer 3: 権限 (Permission)                                         │
│  機能単位のアクセス制御（複数保持可）                                     │
│  ┌────────────┬────────┬────────┬────────┬──────┬──────┬──────┐     │
│  │account_admin│medical │trimming│billing │master│shift │ ...  │     │
│  └────────────┴────────┴────────┴────────┴──────┴──────┴──────┘     │
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

### 2.3 権限 (Permission)

機能単位のアクセス制御。1ユーザーが複数の権限を同時に保持可能。

| 権限 | 値 | 説明 | 対応する機能 |
|---|---|---|---|
| **アカウント管理者** | `account_admin` | スタッフアカウントの追加・編集・無効化、権限の付与・剥奪 | マスタ（スタッフ）、権限管理画面 |
| **医師** | `medical` | カルテの作成・編集・確定、処方箋発行、診断書発行、検査オーダー | カルテ、検査、予防接種、定期健診 |
| **カルテ閲覧** | `medical_read` | カルテの閲覧（編集不可）。看護師等の補助スタッフ向け | カルテ（読み取り専用） |
| **トリマー** | `trimming` | トリミング記録の作成・編集 | トリミング |
| **会計** | `billing` | 会計の作成・編集、入金処理、領収書・明細書発行 | 会計 |
| **受付** | `reception` | 予約の作成・編集・キャンセル、飼主/ペット情報の登録・編集、ダッシュボードのステータス遷移 | 予約、飼主、ペット、ダッシュボード |
| **入院管理** | `hospitalization` | 入院の登録・編集・退院処理、ケアプラン管理、デイリーログ記録 | 入院管理 |
| **マスタ管理** | `master_admin` | マスタデータ（診療項目・薬剤・ケージ等）の追加・編集・無効化 | マスタ設定（全15カテゴリ） |
| **シフト管理** | `shift_admin` | シフトの作成・編集。全スタッフのシフトを管理可能 | シフト管理 |
| **在庫管理** | `inventory` | 在庫品目の登録・編集・数量更新 | 在庫管理 |

**権限の組み合わせ例:**

| ユーザー | ユーザー種別 | 職種 | 保持権限 |
|---|---|---|---|
| 院長（A院） | `clinic_admin` | `veterinarian` | （暗黙的に全権限） |
| 勤務医（A院・B院） | `staff` | `veterinarian` | `medical`, `hospitalization` |
| 主任看護師（A院） | `staff` | `nurse` | `medical_read`, `hospitalization`, `shift_admin`, `inventory` |
| トリマー兼受付（A院） | `staff` | `trimmer` | `trimming`, `reception`, `billing` |
| 受付スタッフ（A院） | `staff` | `reception` | `reception`, `billing` |
| 事務職員（A院） | `staff` | `general_staff` | `inventory`, `billing` |
| 本部管理者 | `system_admin` | — | （暗黙的に全権限） |

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
         │ 1   N
         ▼
┌─────────────────────┐
│  UserPermission      │
│                      │
│  user_id (FK)        │
│  clinic_id (FK)      │
│  permission          │
└─────────────────────┘
```

### 所属ルール

| ルール | 説明 |
|---|---|
| **複数クリニック所属** | 1ユーザーは複数のクリニックに所属可能 |
| **メインクリニック** | 各ユーザーは必ず1つのメインクリニック (`is_main = true`) を持つ。ログイン後の初期表示クリニック |
| **メインクリニック一意制約** | 1ユーザーにつき `is_main = true` は1レコードのみ（部分一意インデックスで保証） |
| **権限のクリニックスコープ** | `UserPermission` はクリニック単位。同一ユーザーがA院では `medical` 権限、B院では `trimming` 権限のみ、という設定が可能 |
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

### ページアクセス権限

| ページ | ルート | `reception` | `medical` | `medical_read` | `trimming` | `billing` | `hospitalization` | `master_admin` | `shift_admin` | `inventory` | `account_admin` |
|---|---|---|---|---|---|---|---|---|---|---|---|
| ダッシュボード | `/` | R/W | R | R | R | R | R | — | — | — | — |
| 飼主一覧 | `/owners` | R/W | R | R | R | R | R | — | — | — | — |
| 飼主登録/編集 | `/owners/new`, `/owners/:id` | R/W | R/W | R | — | R | — | — | — | — | — |
| 予約管理 | `/reservations` | R/W | R | R | R | — | — | — | — | — | — |
| カルテ一覧 | `/medical-records` | R | R/W | R | — | R | R | — | — | — | — |
| カルテ作成/編集 | `/medical-records/new`, `/:id` | — | R/W | R | — | — | — | — | — | — | — |
| 入院一覧 | `/hospitalization` | R | R | R | — | — | R/W | — | — | — | — |
| 入院登録/詳細 | `/hospitalization/new`, `/:id` | — | R/W | R | — | — | R/W | — | — | — | — |
| トリミング一覧 | `/trimming` | R | — | — | R/W | R | — | — | — | — | — |
| トリミング登録/編集 | `/trimming/new`, `/:id` | — | — | — | R/W | — | — | — | — | — | — |
| 検査管理 | `/examinations` | — | R/W | R | — | — | — | — | — | — | — |
| 会計一覧 | `/accounting` | R | R | R | — | R/W | — | — | — | — | — |
| 会計詳細 | `/accounting/:id` | — | R | R | — | R/W | — | — | — | — | — |
| 予防接種一覧 | `/vaccinations` | R | R/W | R | — | — | — | — | — | — | — |
| 定期健診一覧 | `/checkups` | R | R/W | R | — | — | — | — | — | — | — |
| 在庫管理 | `/inventory` | — | — | — | — | — | — | — | — | R/W | — |
| シフト管理 | `/shifts` | R | R | R | R | R | R | — | R/W | — | — |
| マスタ設定 | `/settings/*` | — | — | — | — | — | — | R/W | — | — | — |
| 病院情報 | `/settings/clinic` | — | — | — | — | — | — | R/W | — | — | — |

> **凡例**: R = 閲覧のみ、R/W = 閲覧+編集、— = アクセス不可
> **注**: `system_admin` と `clinic_admin` は全ページにR/Wアクセス可能（マトリクスから省略）。
> **注**: ダッシュボード (`/`) は全権限で閲覧可能。カードのD&D（ステータス遷移）操作は `reception` 権限が必要。

### 機能別操作権限

ページ内の特定機能に対する操作権限の詳細定義。

#### カルテ管理（`/medical-records/:id`）

| タブ / 機能 | 閲覧 | 編集・操作 | 説明 |
|---|---|---|---|
| **問診タブ** | `medical`, `medical_read` | `medical` | 主訴・現病歴・身体検査の入力 |
| **治療タブ** | `medical`, `medical_read`, `billing` | `medical` | 治療項目の追加・削除、数量・単価編集 |
| **処方タブ** | `medical`, `medical_read` | `medical` | 処方薬の追加・削除、処方内容編集 |
| **予防接種タブ** | `medical`, `medical_read`, `reception` | `medical` | ワクチン接種記録の登録・編集 |
| **定期健診タブ** | `medical`, `medical_read`, `reception` | `medical` | 健診記録の登録・編集 |
| **検査タブ** | `medical`, `medical_read` | `medical` | 検査オーダー・結果記録 |
| **画像タブ** | `medical`, `medical_read` | `medical` | 画像アップロード・削除 |
| **見積書タブ** | `medical`, `medical_read`, `billing` | `medical` | 見積PDF出力 |
| **会計確認タブ** | `medical`, `medical_read`, `billing` | `medical` | 算定チェック・確認 |
| **会計確認タブ - チェック完了** | — | `medical` | 「チェック完了」「未チェックに戻す」トグル操作（医師による算定最終確認） |
| **会計確認タブ - 会計へ進む/会計を確認** | `medical`, `billing` | `medical`, `billing` | 会計画面への遷移・既存会計の確認 |
| **バイタル入力** | — | `medical`, `medical_read`, `hospitalization` | 体温・心拍数・呼吸数・体重の記録 |
| **カルテ確定** | — | `medical` | カルテステータスを「確定」に変更（編集ロック） |

#### 入院管理（`/hospitalization/:id`）

| 機能 | 閲覧 | 編集・操作 | 説明 |
|---|---|---|---|
| **入院基本情報** | `medical`, `medical_read`, `hospitalization` | `medical`, `hospitalization` | 入院日・退院予定日・ケージ・主治医等 |
| **ケアプラン** | `medical`, `medical_read`, `hospitalization` | `medical`, `hospitalization` | ケア項目・タイミング・担当スタッフの設定 |
| **デイリーログ** | `medical`, `medical_read`, `hospitalization` | `hospitalization` | 日次のケア実施記録・バイタル・スタッフメモ |
| **退院処理** | — | `medical`, `hospitalization` | 退院ステータスへの変更 |

#### 会計管理（`/accounting/:id`）

| 機能 | 閲覧 | 編集・操作 | 説明 |
|---|---|---|---|
| **会計明細編集** | `medical`, `medical_read`, `billing` | `billing` | 明細項目の追加・削除・数量変更 |
| **入金処理** | `medical`, `medical_read`, `billing` | `billing` | 入金額・支払方法の記録 |
| **領収書・明細書発行** | `medical`, `medical_read`, `billing` | `billing` | PDF出力 |
| **会計確定** | — | `billing` | 会計ステータスを「確定」に変更 |

#### トリミング管理（`/trimming/:id`）

| 機能 | 閲覧 | 編集・操作 | 説明 |
|---|---|---|---|
| **施術内容編集** | `trimming`, `billing` | `trimming` | コース・オプション・担当トリマーの設定 |
| **施術完了** | — | `trimming` | ステータスを「完了」に変更 |

> **凡例**:
> - 閲覧権限を持つユーザーはタブ/機能を表示・参照可能
> - 編集・操作権限を持つユーザーのみがボタン操作・データ更新可能
> - 複数の権限が記載されている場合、いずれか1つを保持していれば操作可能

### サイドナビバー表示制御

権限に基づきサイドナビバーのメニュー項目を動的にフィルタリング。

| メニュー項目 | 必要な権限（いずれか1つ） |
|---|---|
| ダッシュボード | （常時表示） |
| 予約管理 | `reception`, `medical`, `medical_read`, `trimming` |
| 飼主管理 | `reception`, `medical`, `medical_read`, `billing` |
| カルテ | `medical`, `medical_read`, `billing`, `hospitalization` |
| 入院管理 | `medical`, `medical_read`, `hospitalization` |
| トリミング | `trimming`, `billing` |
| 検査管理 | `medical`, `medical_read` |
| 会計 | `billing`, `medical`, `medical_read` |
| 予防接種 | `medical`, `medical_read`, `reception` |
| 定期健診 | `medical`, `medical_read`, `reception` |
| 在庫管理 | `inventory` |
| シフト管理 | （常時表示、編集は `shift_admin` 必要） |
| マスタ設定 | `master_admin`, `account_admin` |

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

-- 権限
CREATE TYPE permission_type AS ENUM (
  'account_admin',
  'medical',
  'medical_read',
  'trimming',
  'billing',
  'reception',
  'hospitalization',
  'master_admin',
  'shift_admin',
  'inventory'
);

-- アカウントステータス
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');
```

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

#### `user_permissions` — ユーザー権限（クリニックスコープ）

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGINT` | `PK DEFAULT -` | 権限レコードID |
| `user_id` | `BIGINT` | `FK → user_accounts.id NOT NULL` | ユーザーID |
| `clinic_id` | `BIGINT` | `FK → clinics.id NOT NULL` | クリニックID |
| `permission` | `permission_type` | `NOT NULL` | 権限種別 |
| `granted_by` | `BIGINT` | `FK → user_accounts.id` | 権限付与者 |
| `granted_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 付与日時 |

```sql
CREATE TABLE user_permissions (
  id          BIGINT PRIMARY KEY DEFAULT -,
  user_id     BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  clinic_id   BIGINT NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
  permission  permission_type NOT NULL,
  granted_by  BIGINT REFERENCES user_accounts(id) ON DELETE SET NULL,
  granted_at  TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, clinic_id, permission)
);
```

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

-- ===== user_permissions =====
-- UNIQUE(user_id, clinic_id, permission) が複合インデックスを暗黙作成
CREATE INDEX idx_user_permissions_clinic ON user_permissions(clinic_id);

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

-- 指定クリニックでの権限チェック
CREATE OR REPLACE FUNCTION auth.has_permission(
  p_clinic_id BIGINT,
  p_permission permission_type
)
RETURNS BOOLEAN AS $$
  SELECT EXISTS (
    SELECT 1 FROM user_accounts ua
    WHERE ua.id = auth.uid()
    AND (
      ua.user_type = 'system_admin'
      OR ua.user_type = 'clinic_admin' AND EXISTS (
        SELECT 1 FROM user_clinic_memberships ucm
        WHERE ucm.user_id = ua.id AND ucm.clinic_id = p_clinic_id
      )
      OR EXISTS (
        SELECT 1 FROM user_permissions up
        WHERE up.user_id = ua.id
        AND up.clinic_id = p_clinic_id
        AND up.permission = p_permission
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

-- カルテ: medical 権限が必要
CREATE POLICY "medical_write" ON medical_records
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'medical'));

CREATE POLICY "medical_update" ON medical_records
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'medical'));

-- 会計: billing 権限が必要
CREATE POLICY "billing_write" ON accountings
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'billing'));

CREATE POLICY "billing_update" ON accountings
  FOR UPDATE TO authenticated
  USING (auth.has_permission(clinic_id, 'billing'));

-- マスタ: master_admin 権限が必要
CREATE POLICY "master_write" ON master_items
  FOR ALL TO authenticated
  USING (auth.has_permission(clinic_id, 'master_admin'))
  WITH CHECK (auth.has_permission(clinic_id, 'master_admin'));

-- 入院: hospitalization または medical 権限が必要
CREATE POLICY "hospitalization_write" ON hospitalizations
  FOR INSERT TO authenticated
  WITH CHECK (
    auth.has_permission(clinic_id, 'hospitalization')
    OR auth.has_permission(clinic_id, 'medical')
  );

-- トリミング: trimming 権限が必要
CREATE POLICY "trimming_write" ON trimming_records
  FOR INSERT TO authenticated
  WITH CHECK (auth.has_permission(clinic_id, 'trimming'));

-- シフト: shift_admin 権限が必要
CREATE POLICY "shift_write" ON shift_entries
  FOR ALL TO authenticated
  USING (auth.has_permission(clinic_id, 'shift_admin'))
  WITH CHECK (auth.has_permission(clinic_id, 'shift_admin'));

-- 在庫: inventory 権限が必要
CREATE POLICY "inventory_write" ON inventory_items
  FOR ALL TO authenticated
  USING (auth.has_permission(clinic_id, 'inventory'))
  WITH CHECK (auth.has_permission(clinic_id, 'inventory'));
```

---

## フロントエンド実装方針

### 新規ファイル構成

```
/features/auth/
├── api/
│   ├── index.ts              # barrel re-export
│   ├── mockData.ts           # モックユーザー・権限データ
│   ├── login.ts              # ログインAPI
│   ├── logout.ts             # ログアウトAPI
│   └── refreshToken.ts       # トークンリフレッシュ
├── components/
│   ├── LoginForm.tsx          # ログインフォーム
│   ├── ClinicSwitcher.tsx     # クリニック切替コンポーネント（サイドバー用）
│   └── PermissionGate.tsx     # 権限ガードコンポーネント（children を条件付きレンダリング）
├── hooks/
│   ├── useAuth.ts             # 認証状態管理フック（AuthContext）
│   └── usePermission.ts      # 権限チェックフック
├── routes/
│   ├── index.ts              # barrel re-export
│   └── Login.tsx             # ログインページ
└── types/
    └── index.ts              # 認証関連型定義
```

### 型定義 (`features/auth/types/index.ts`)

```typescript
// ユーザー種別
export const USER_TYPE_VALUES = ["system_admin", "clinic_admin", "staff"] as const;
export type UserType = (typeof USER_TYPE_VALUES)[number];

export const USER_TYPE_LABELS: Record<UserType, string> = {
  system_admin: "運営管理者",
  clinic_admin: "医院管理者",
  staff: "スタッフ",
};

// 職種
export const JOB_TITLE_VALUES = ["veterinarian", "nurse", "trimmer", "reception", "general_staff"] as const;
export type JobTitle = (typeof JOB_TITLE_VALUES)[number];

export const JOB_TITLE_LABELS: Record<JobTitle, string> = {
  veterinarian: "医師",
  nurse: "看護師",
  trimmer: "トリマー",
  reception: "受付",
  general_staff: "職員",
};

// 権限
export const PERMISSION_VALUES = [
  "account_admin",
  "medical",
  "medical_read",
  "trimming",
  "billing",
  "reception",
  "hospitalization",
  "master_admin",
  "shift_admin",
  "inventory",
] as const;
export type Permission = (typeof PERMISSION_VALUES)[number];

export const PERMISSION_LABELS: Record<Permission, string> = {
  account_admin: "アカウント管理者",
  medical: "医師",
  medical_read: "カルテ閲覧",
  trimming: "トリマー",
  billing: "会計",
  reception: "受付",
  hospitalization: "入院管理",
  master_admin: "マスタ管理",
  shift_admin: "シフト管理",
  inventory: "在庫管理",
};

// アカウントステータス
export const ACCOUNT_STATUS_VALUES = ["active", "inactive", "locked"] as const;
export type AccountStatus = (typeof ACCOUNT_STATUS_VALUES)[number];

// ログインユーザー情報
export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  userType: UserType;
  jobTitle: JobTitle | null;
  avatarUrl: string | null;
  mainClinicId: string;
  clinics: ClinicMembership[];
  permissions: ClinicPermissions;
}

export interface ClinicMembership {
  clinicId: string;
  clinicName: string;
  branchName: string;
  isMain: boolean;
}

// クリニック別権限マップ: { [clinicId]: Permission[] }
export type ClinicPermissions = Record<string, readonly Permission[]>;

// 認証コンテキスト
export interface AuthContextValue {
  user: AuthUser | null;
  currentClinicId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  switchClinic: (clinicId: string) => void;
  hasPermission: (permission: Permission) => boolean;
  hasAnyPermission: (permissions: readonly Permission[]) => boolean;
}
```

### 権限チェックパターン

```typescript
// usePermission フック
export function usePermission() {
  const { user, currentClinicId } = useAuth();

  function hasPermission(permission: Permission): boolean {
    if (!user || !currentClinicId) return false;
    // system_admin は全権限
    if (user.userType === "system_admin") return true;
    // clinic_admin は所属クリニック内全権限
    if (user.userType === "clinic_admin") {
      return user.clinics.some((c) => c.clinicId === currentClinicId);
    }
    // staff は明示的権限チェック
    const clinicPerms = user.permissions[currentClinicId];
    if (!clinicPerms) return false;
    return clinicPerms.includes(permission);
  }

  return { hasPermission, hasAnyPermission };
}
```

```tsx
// PermissionGate コンポーネント
export function PermissionGate({
  permission,
  anyOf,
  fallback = null,
  children,
}: PermissionGateProps) {
  const { hasPermission, hasAnyPermission } = usePermission();

  if (permission && !hasPermission(permission)) return fallback;
  if (anyOf && !hasAnyPermission(anyOf)) return fallback;

  return <>{children}</>;
}

// 使用例: サイドバーのメニュー項目
<PermissionGate anyOf={["medical", "medical_read", "billing", "hospitalization"]}>
  <SidebarMenuItem path="/medical-records" label="カルテ" />
</PermissionGate>
```

### ルーティング統合

```typescript
// ProtectedRoute ラッパー
export function ProtectedRoute({
  permission,
  anyOf,
  children,
}: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();
  const { hasPermission, hasAnyPermission } = usePermission();

  if (isLoading) return <RouteFallback />;
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (permission && !hasPermission(permission)) return <Navigate to="/" replace />;
  if (anyOf && !hasAnyPermission(anyOf)) return <Navigate to="/" replace />;

  return <>{children}</>;
}

// routes.ts での使用
createBrowserRouter([
  {
    path: "/login",
    Component: Login,
  },
  {
    path: "/",
    Component: Root, // AuthProvider + Sidebar ラップ
    children: [
      { index: true, Component: Dashboard },
      {
        path: "medical-records",
        Component: () => (
          <ProtectedRoute anyOf={["medical", "medical_read", "billing"]}>
            <MedicalRecords />
          </ProtectedRoute>
        ),
      },
      // ...
    ],
  },
]);
```

### 新規ルート

| ルート | コンポーネント | 説明 |
|---|---|---|
| `/login` | `Login` | ログインページ |

> ルート総数: 41 → 42 (+1)

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

### 数値整合

| 項目 | 現在（Phase 0） | 認証実装後（Phase 4完了） |
|---|---|---|
| ルート総数 | 41 (+`/dev/tests`) | 42 (+`/login`) |
| Feature数 | 15 | 16 (+`auth`) |
| エンティティ数 | 31 | 35 (+`UserAccount`, `Clinic`, `UserClinicMembership`, `UserPermission`) |
| テーブル数 | 27 | 31 (+`clinics`, `user_accounts`, `user_clinic_memberships`, `user_permissions`) |
| lazy-loaded コンポーネント数 | 29 | 30 (+`Login`) |

---

## 備考

1. **`staff_role` enum の移行**: 既存の `staff_role` (`veterinarian`, `nurse`, `trimmer`, `reception`, `manager`) は `job_title` enum に段階的に移行する。`manager` は `UserType.clinic_admin` に対応するため、`job_title` enum では `general_staff` に置換。移行期間中は両方の enum が共存する。

2. **パスワードハッシュ**: `user_accounts` テーブルにはパスワードカラムを持たない。パスワード管理は認証基盤（自前の場合は別テーブル `user_credentials`、Supabase の場合は `auth.users`）に委譲する。

3. **運営管理者のクリニック所属**: `system_admin` は `user_clinic_memberships` にレコードを持たなくてもよい（全クリニックアクセス可能）。ただし UI の利便性のため、メインクリニックの設定は推奨。

4. **マスタデータのクリニックスコープ**: `master_items` に `clinic_id` を追加すると、クリニックごとにマスタデータが独立する。法人全体で共有すべきマスタ項目がある場合は、`clinic_id = NULL` をグローバルマスタとして扱い、RLS ポリシーで「自クリニック OR NULL」を許可する設計も検討。

5. **職種テンプレートの自動適用**: アカウント作成時に `job_title` を選択すると、対応するデフォルト権限テンプレート（2.2 節参照）が自動的に `user_permissions` に挿入される。作成後に個別編集可能。

6. **既存モックデータとの互換**: Phase 1 ではモック認証（`mockData.ts` に定義したユーザーでログイン）で動作させ、既存の Mock Data 環境を維持する。バックエンド接続時にリアル認証に切り替え。

7. **WCAG AA 準拠**: ログインフォームは既存のフォームアクセシビリティパターン（`FormFieldError` + `aria-describedby`、`NavigationBlocker` 不要）に準拠。パスワード入力は `type="password"` + 表示/非表示トグル（`aria-label` 付き）。

8. **印刷機能との関係**: 帳票（領収書・処方箋等）にクリニック情報を表示する箇所は、`currentClinicId` に基づく `clinics` テーブルのデータを使用するよう更新が必要。

---

## 関連ドキュメント

| ドキュメント | パス | 関連箇所 |
|---|---|---|
| **仕様定義書** | `docs/SPECIFICATION.md` | Feature一覧、ルーティング構成、ロードマップ |
| **画面仕様書** | `docs/SCREENS.md` | 全ルートの画面仕様（`/login` §15 追加済み） |
| **ER図** | `docs/ERD.md` | エンティティ一覧（4エンティティ追加）、リレーション追加 |
| **ER図** | `docs/ERD.md` | ENUM型追加、テーブル追加、`clinic_id` カラム追加（v5.0 最新版） |
| **デザインシステム** | `docs/DESIGN_SYSTEM.md` | ログインページのUI仕様（§15 Login参照） |
