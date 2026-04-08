# 動物病院管理システム 認証・認可設計書

本ドキュメントは、システムの認証（Authentication）および認可（Authorization）の設計を定義します。
マルチクリニック対応のRBAC（Role-Based Access Control）モデルを採用し、ユーザー種別・職種・権限の3層構造で柔軟なアクセス制御を実現します。

**バージョン**: v7.0 | **最終更新**: 2026-04-06

> **ドキュメント方針**: 本ドキュメントは「あるべき姿（ベストプラクティス）」を記述します。
> 現在の実装が本設計と乖離している箇所は `> ⚠️ 現在の実装:` として注記しています。

---

## 目次

1. [設計方針](#設計方針)
2. [ユーザーモデル（3層構造）](#ユーザーモデル3層構造)
   - [2.1 ユーザー種別 (UserType)](#21-ユーザー種別-usertype)
   - [2.2 職種 (StaffRole / Occupation)](#22-職種-staffrole--jobtitle)
   - [2.3 権限グループ (PermissionGroup)](#23-権限グループ-permissiongroup)
3. [マルチクリニック設計](#マルチクリニック設計)
4. [認証フロー](#認証フロー)
5. [権限マトリクス](#権限マトリクス)
6. [DB設計](#db設計)
   - [6.1 ENUM型](#61-enum型)
   - [6.2 テーブル](#62-テーブル)
   - [6.3 既存テーブルへの影響](#63-既存テーブルへの影響)
   - [6.4 インデックス](#64-インデックス)
7. [アプリケーション層認可設計](#アプリケーション層認可設計)
8. [フロントエンド実装方針](#フロントエンド実装方針)
9. [既存システムとの統合](#既存システムとの統合)

---

## 設計方針

| 項目 | 方針 |
|------|------|
| **3層モデル** | ユーザー種別（システムレベル）→ 職種（表示・テンプレート用）→ 権限グループ（リソース×CRUD制御）の3層で分離 |
| **権限グループの複数保持** | 1ユーザーが複数の権限グループに所属可能（例: 一般 + 管理者） |
| **companyスコープ** | 権限グループは company 単位で管理される（クリニック横断で共通の権限設定）。データの clinic_id 分離は維持 |
| **職種と権限の分離** | 職種（何者であるか）と権限（何ができるか）を明確に分離。職種は権限テンプレートの初期値として使用 |
| **最小権限の原則** | デフォルトでは最小限の権限のみ付与。必要に応じて権限を追加 |
| **認証方式** | メールアドレス + パスワードによるログイン。将来的にSSO/SAML対応を想定 |
| **セッション管理** | 短命アクセストークン (15分 JWT) + 長命リフレッシュトークン (7日 opaque) の dual-token。両方 httpOnly Cookie で管理。リフレッシュトークンは DB に保存しサーバー側で無効化可能にする |
| **監査ログ** | ログイン・ログアウト・権限変更・重要リソース操作を `audit_logs` テーブルに記録 |

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
│  Layer 2: 職種 (StaffRole / Occupation)                               │
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

### 2.2 職種 (StaffRole / Occupation)

業務上の肩書き。UI表示・シフト管理でのフィルタリング・権限テンプレートの初期値決定に使用。

**現在の実装（2層構造）:**

| カラム / テーブル | 型 | 場所 | 説明 |
|---|---|---|---|
| `staffs.staff_role` | `staff_role` ENUM | `staffs` テーブル | 職種分類（`veterinarian|nurse|trimmer|reception|manager`） |
| `staffs.occupation_id` | bigint FK → `occupations` | `staffs` テーブル | 表示名・ラベル用の職種マスタ |
| `user_accounts.occupation_id` | bigint FK → `occupations` | `user_accounts` テーブル | ユーザー直接の職種（`staffs` 非紐付けの場合） |
| `user_accounts.staff_id` | bigint FK → `staffs` | `user_accounts` テーブル | スタッフマスタへの紐付け |

`AuthUser.staffRole` はバックエンドが `user_accounts → staffs.staff_role` を JOIN して返す。

| 職種 | `staff_role` 値 | 説明 |
|---|---|---|
| **医師** | `veterinarian` | 獣医師。診療・カルテ・処方等の医療行為全般 |
| **看護師** | `nurse` | 動物看護師。医師の補助、入院管理、バイタル記録 |
| **トリマー** | `trimmer` | トリミング施術担当 |
| **受付** | `reception` | 受付業務。予約・会計・飼主/ペット情報管理 |
| **管理職** | `manager` | 医院管理者に準ずるロール（`clinic_admin` ユーザーに対応） |

> **注**: `general_staff` 値は migration に**存在しない**。`manager` は現在も `staff_role` ENUM に残存。
> `occupation` ENUM 型も migration に**存在しない**（設計上の `Occupation` ENUM 移行は未実施）。
> デフォルト権限テンプレートはアカウント作成時の初期値であり、作成後に個別調整可能。

### 2.3 権限グループ (PermissionGroup)

**グループベース RBAC（実装済み）**: `permission_groups` テーブルで定義されたグループを通じてアクセスを制御。1ユーザーが複数のグループに所属可能。

#### 権限モデル

各権限グループは **23リソース × CRUD（4操作）** の細粒度アクセス制御を持つ。

**業務リソース（14種）:**

| リソース | キー | 説明 |
|---|---|---|
| 当日の受付 | `reception` | 当日の受付（カンバンボード） |
| 飼主・ペット | `owners` | 飼主・ペット情報 |
| 予約管理 | `reservations` | 予約登録・変更 |
| カルテ | `medical-records` | カルテ作成・確定 |
| 入院・ホテル | `hospitalization` | 入院・退院処理 |
| トリミング | `trimming` | トリミング記録 |
| 検査管理 | `examinations` | 検査記録 |
| 会計管理 | `accounting` | 会計・入金処理 |
| 予防接種 | `vaccinations` | ワクチン記録 |
| 定期健診 | `checkups` | 健診記録 |
| 在庫管理 | `inventory` | 在庫品目管理 |
| 見積書 | `estimates` | 見積書管理 |
| シフト管理 | `shifts` | シフト登録 |
| 医院 | `hospital-settings` | クリニック基本情報 |

**マスタ設定リソース（9種）:**

| リソース | キー | 説明 |
|---|---|---|
| 動物種類 | `master-animal-species` | 動物種類マスタ |
| カルテ関連 | `master-medical` | 診療項目・診断病名・問診設定・薬剤マスタ |
| 診療サービス | `master-service-type` | 予約区分マスタ |
| 入院・ケージ | `master-hospitalization` | 入院料金・ケージマスタ |
| トリミングマスタ | `master-trimming` | トリミングコース・オプションマスタ |
| 権限グループ | `master-permission` | 権限グループの管理（一般スタッフは非表示） |
| スタッフ管理 | `master-staff` | スタッフ情報・権限割当 |
| 保険マスタ | `master-insurance` | ペット保険会社マスタ |
| 物販・フード | `master-merchandise` | 販売品目マスタ |

#### シードデータ：3つの権限グループ

| グループ名 | カラー | 主な用途 |
|---|---|---|
| **管理者** | `#EF4444` | 全23リソースフルアクセス |
| **執行** | `#6366F1` | 全リソース閲覧 + 業務・マスタの作成・編集（権限・スタッフ管理は閲覧のみ） |
| **一般** | `#10B981` | 基本業務操作 + マスタ閲覧（権限グループ管理は非表示） |

> `system_admin` / `clinic_admin` は全リソースへの暗黙的フルアクセス（グループ不要）。
> `master-permission` は「一般」グループで `canView=false` のため、サイドナビから非表示になる。

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
│  occupation_id│       │  joined_at            │       │  ...         │
│  ...         │       └──────────────────────┘       └──────────────┘
└─────────────┘
         │
         │ N   N  （staff_permission_groups 中間テーブル）
         ▼
┌─────────────────────┐       ┌──────────────────────────┐
│  PermissionGroup     │ 1   N │  PermissionGroupRule      │
│                      │───────│                           │
│  id                  │       │  group_id (FK)            │
│  company_id (FK)     │       │  resource (TEXT)          │
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
| **権限の companyスコープ** | `permission_groups` は company 単位で管理（TASK-049）。グループ設定は全クリニック共通。データアクセスの clinic_id 分離とは独立して管理 |
| **運営管理者の例外** | `system_admin` はクリニック所属に関係なく全クリニックにアクセス可能 |

### クリニック切替

- サイドナビバーのヘッダー部分に現在のクリニック名を表示
- 所属クリニックが複数ある場合、クリックでドロップダウンメニューを展開
- クリニック切替時:
  1. `currentClinicId` をコンテキストで更新
  2. 権限は company 単位のため追加 API コール不要（既存の `permissions` はそのまま有効）
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
  │                           │  POST /v1/login             │
  │                           │  { email, password }        │
  │                           │ ────────────────────────────>│
  │                           │                             │  1. FindByEmail()
  │                           │                             │  2. bcrypt.Compare()
  │                           │                             │  3. GetMemberships()
  │                           │                             │  4. CalcEffectivePermissions()
  │                           │                             │  5. JWT(15分) 生成
  │                           │                             │  6. RefreshToken(opaque,7日) 生成
  │                           │                             │  7. refresh_tokens テーブルに保存
  │                           │                             │  8. Set-Cookie: access_token (15分)
  │                           │                             │     Set-Cookie: refresh_token (7日)
  │                           │  200 OK                     │
  │                           │  { user: {                  │
  │                           │      id, email, userType,   │
  │                           │      staffRole, avatarUrl,  │
  │                           │      mainClinicId,          │
  │                           │      clinics: [...],        │
  │                           │      permissions: {         │
  │                           │        resource → CRUD      │
  │                           │      }                      │
  │                           │    }                        │
  │                           │  }                          │
  │                           │ <────────────────────────────│
  │                           │                             │
  │                           │  AuthContext にセット         │
  │                           │  currentClinicId =          │
  │                           │    mainClinicId             │
  │                           │                             │
  │  メインクリニックの         │                             │
  │  ダッシュボード (`/`) へ遷移 │                             │
  │ <─────────────────────────│                             │
```

> ✅ 実装済み: dual-token（`access_token` 15分 JWT + `refresh_token` 7日 opaque token）。`refresh_tokens` テーブルでサーバー側無効化対応済み。

### トークンリフレッシュフロー

```
フロントエンド (Axiosインターセプター)       バックエンド
  │                                          │
  │  API リクエスト (access_token 期限切れ)  │
  │ ────────────────────────────────────────>│ 401 Unauthorized
  │ <────────────────────────────────────────│
  │                                          │
  │  POST /v1/auth/refresh                   │
  │  (Cookie: refresh_token 自動送信)         │
  │ ────────────────────────────────────────>│
  │                                          │  1. refresh_tokens テーブルで検証
  │                                          │  2. 期限・無効化フラグ確認
  │                                          │  3. 旧トークンを無効化 (ローテーション)
  │                                          │  4. 新 access_token (15分) 生成
  │                                          │  5. 新 refresh_token (7日) 生成・DB保存
  │                                          │  Set-Cookie: access_token (新)
  │  200 OK                                  │  Set-Cookie: refresh_token (新)
  │ <────────────────────────────────────────│
  │                                          │
  │  元のリクエストをリトライ                  │
  │ ────────────────────────────────────────>│
```

**リフレッシュトークン ローテーション**: リフレッシュのたびに旧トークンを無効化し新トークンを発行する。
これにより盗まれたリフレッシュトークンの使用を1回のみに限定できる（検知も可能）。

### ログイン後の初期遷移

| 条件 | 遷移先 |
|---|---|
| 通常ログイン | メインクリニックのダッシュボード (`/`) = 当日の受付ページ |
| セッション復帰（リフレッシュ） | 前回表示していたページ |
| パスワードリセット後 | パスワード変更完了画面 → ダッシュボード |

### セッション管理

| 項目 | 仕様 |
|---|---|
| **アクセストークン** | JWT (HS256)、有効期限 **15分**、httpOnly Cookie (`access_token`) |
| **リフレッシュトークン** | Opaque token (crypto.randomBytes(32) → hex)、有効期限 **7日**、httpOnly Cookie (`refresh_token`)、`refresh_tokens` テーブルに保存 |
| **Cookie 設定** | `HttpOnly: true`、`Secure: true` (本番)、`SameSite: None` (本番・クロスドメイン) / `Lax` (開発)、`Path: /` |
| **トークンリフレッシュ** | Axios レスポンスインターセプターが 401 を検知 → `POST /v1/auth/refresh` → 元リクエストをリトライ |
| **トークンローテーション** | リフレッシュのたびに旧リフレッシュトークンを無効化し新トークンを発行（再使用検知のため） |
| **ログアウト** | `DELETE /v1/auth/refresh` でサーバー側の refresh_tokens レコードを削除 + 両 Cookie を MaxAge=-1 でクリア |
| **強制ログアウト** | `DELETE /v1/users/:id/sessions` で `refresh_tokens` テーブルの全レコードを削除（パスワード変更・アカウント停止時） |
| **同時セッション** | 1ユーザーにつき最大 **5デバイス**まで（超過時は最古セッションを無効化） |

> ✅ 実装済み: dual-token（`access_token` 15分 + `refresh_token` 7日）、httpOnly Cookie、`refresh_tokens` テーブルによるサーバー側無効化。本番は `SameSite=None`（Vercel ↔ CloudFront クロスドメイン対応）。

#### Cookie 設定の根拠

| 設定 | 目的 |
|------|------|
| `HttpOnly` | JavaScript から Cookie にアクセス不可 → XSS でトークン盗取を防止 |
| `Secure` | HTTPS のみ送信 → 中間者攻撃を防止 |
| `SameSite=None` (本番) | Vercel (フロントエンド) ↔ CloudFront (API) のクロスドメイン通信で Cookie を送信するために必要。`Secure=true` とセット必須 |
| トークンを localStorage に保存しない | XSS で `document.cookie` にアクセスできなくても localStorage は読み取り可能なため |

---

## 権限マトリクス

### 権限グループ別リソースアクセス（実装済みシードデータ）

凡例: ✓=可、−=不可。操作列は View / Create / Edit / Delete の順。

**業務リソース:**

| リソース | 管理者 | 執行 | 一般 |
|---|---|---|---|
| `reception` | ✓/−/−/− | ✓/−/−/− | ✓/−/−/− |
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
| `hospital-settings` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/−/−/− |

**マスタ設定リソース:**

| リソース | 管理者 | 執行 | 一般 |
|---|---|---|---|
| `master-animal-species` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `master-medical` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `master-service-type` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `master-hospitalization` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `master-trimming` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `master-permission` | ✓/✓/✓/✓ | ✓/−/−/− | **−/−/−/−** |
| `master-staff` | ✓/✓/✓/✓ | ✓/−/−/− | ✓/−/−/− |
| `master-insurance` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |
| `master-merchandise` | ✓/✓/✓/✓ | ✓/✓/✓/− | ✓/−/−/− |

> **重要**: `master-permission` は「一般」グループで全操作不可（`canView=false`）。サイドナビから非表示になる。
> `system_admin` / `clinic_admin` は全リソースへの暗黙的フルアクセス（権限グループ不要）。

### サイドナビバー表示制御

`usePermission(resource).canView` が true のリソースに応じてフィルタリング。

**業務メニュー:**

| メニュー項目 | 必要なリソース（`canView` が true） |
|---|---|
| 当日の受付 | `reception` |
| 飼主・ペット | `owners` |
| 予約管理 | `reservations` |
| カルテ | `medical-records` |
| 検査管理 | `examinations` |
| トリミング | `trimming` |
| 予防接種 | `vaccinations` |
| 定期健診 | `checkups` |
| 会計管理 | `accounting` |
| 入院・ホテル | `hospitalization` |
| 在庫管理 | `inventory` |
| シフト管理 | `shifts` |

**マスタ設定メニュー（各サブ項目が個別に制御）:**

| メニュー項目 | 必要なリソース（`canView` が true） |
|---|---|
| 医院 | `hospital-settings` |
| 動物種類 | `master-animal-species` |
| カルテ関連 | `master-medical` |
| 診療サービス | `master-service-type` |
| 入院・ケージ | `master-hospitalization` |
| トリミング | `master-trimming` |
| 権限グループ | `master-permission` |
| スタッフ管理 | `master-staff` |
| 保険マスタ | `master-insurance` |
| 物販・フード | `master-merchandise` |

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

### 6.1 ENUM型

```sql
-- ユーザー種別
CREATE TYPE user_type AS ENUM ('system_admin', 'clinic_admin', 'staff');

-- 職種（staffs テーブルで使用）
CREATE TYPE staff_role AS ENUM ('veterinarian', 'nurse', 'trimmer', 'reception', 'manager');

-- アカウントステータス
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');
```

> ⚠️ `occupation` ENUM は migration に**存在しない**。設計ドキュメントで言及される `general_staff` 値も未実装。
> `occupations` は FK テーブルとして管理（`staffs.occupation_id`, `user_accounts.occupation_id`）。

> **廃止済み**: `permission_type` ENUM（`account_admin`, `medical` 等の10値）は**実装されていない**。
> 権限制御は `permission_groups` + `permission_group_rules` テーブルで行う（§6.2参照）。

### 6.2 テーブル

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
| `id` | `BIGSERIAL` | `PK` | ユーザーID |
| `email` | `text` | `NOT NULL UNIQUE` | メールアドレス（ログインID） |
| `display_name` | `text` | `NOT NULL` | 表示名 |
| `display_name_kana` | `text` | `NOT NULL DEFAULT ''` | 表示名カナ |
| `user_type` | `user_type` | `NOT NULL DEFAULT 'staff'` | ユーザー種別 |
| `occupation_id` | `bigint` | `FK → occupations(id) SET NULL` | 職種マスタ（`system_admin` は NULL 可） |
| `status` | `account_status` | `DEFAULT 'active'` | アカウントステータス |
| `avatar_url` | `text` | `NOT NULL DEFAULT ''` | アバター画像URL |
| `staff_id` | `bigint` | `FK → staffs(id) SET NULL` | スタッフマスタへの紐付け（`staff_role` 取得に使用） |
| `password_hash` | `text` | `NOT NULL DEFAULT ''` | bcrypt ハッシュ |
| `created_at` | `timestamptz` | `NOT NULL DEFAULT now()` | 作成日時 |
| `updated_at` | `timestamptz` | `NOT NULL DEFAULT now()` | 更新日時 |
| `deleted_at` | `timestamptz` | | 論理削除（NULL = 有効） |

```sql
CREATE TABLE user_accounts (
    id                BIGSERIAL      PRIMARY KEY,
    email             text           NOT NULL UNIQUE,
    display_name      text           NOT NULL,
    display_name_kana text           NOT NULL DEFAULT '',
    user_type         user_type      NOT NULL DEFAULT 'staff',
    occupation_id      bigint                  REFERENCES occupations(id) ON DELETE SET NULL,
    status            account_status          DEFAULT 'active',
    avatar_url        text           NOT NULL DEFAULT '',
    staff_id          bigint                  REFERENCES staffs(id) ON DELETE SET NULL,
    password_hash     text           NOT NULL DEFAULT '',
    created_at        timestamptz    NOT NULL DEFAULT now(),
    updated_at        timestamptz    NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);
```

> **`staff_id`**: `staffs` テーブルへの FK。`staffs.staff_role` を通じて `AuthUser.staffRole` を取得。
> `occupation` ENUM 列は存在しない。職種は `occupations` FK テーブルで管理。

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

#### `permission_groups` — 権限グループ

> ⚠️ **TASK-049**: `clinic_id` → `company_id` への移行により、権限グループが全クリニック共通になる予定。
> **現在の実装**は `clinic_id` で clinic スコープ管理。

**現在の実装（`clinic_id` ベース）:**

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGSERIAL` | `PK` | グループID |
| `clinic_id` | `bigint` | `FK → clinics.id NOT NULL` | クリニックID |
| `name` | `varchar(100)` | `NOT NULL` | グループ名 |
| `description` | `text` | `NOT NULL DEFAULT ''` | グループ説明 |
| `color` | `varchar(7)` | `NOT NULL DEFAULT '#6B7280'` | 表示カラー（HEX） |
| `created_at` | `timestamptz` | `NOT NULL DEFAULT now()` | 作成日時 |
| `updated_at` | `timestamptz` | `NOT NULL DEFAULT now()` | 更新日時 |
| `deleted_at` | `timestamptz` | | 論理削除（NULL = 有効） |

```sql
-- 現在の実装
CREATE TABLE permission_groups (
    id          BIGSERIAL    PRIMARY KEY,
    clinic_id   bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name        varchar(100) NOT NULL,
    description text         NOT NULL DEFAULT '',
    color       varchar(7)   NOT NULL DEFAULT '#6B7280',
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- TASK-049 完了後（company_id へ移行）
CREATE TABLE permission_groups (
    id          BIGSERIAL    PRIMARY KEY,
    company_id  bigint       NOT NULL REFERENCES company(id) ON DELETE CASCADE,
    name        varchar(100) NOT NULL,
    description text         NOT NULL DEFAULT '',
    color       varchar(7)   NOT NULL DEFAULT '#6B7280',
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now(),
    deleted_at  timestamptz
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
    id         BIGSERIAL   PRIMARY KEY,
    group_id   bigint      NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    resource   varchar(50) NOT NULL,
    can_view   boolean     NOT NULL DEFAULT false,
    can_create boolean     NOT NULL DEFAULT false,
    can_edit   boolean     NOT NULL DEFAULT false,
    can_delete boolean     NOT NULL DEFAULT false,
    CONSTRAINT uk_permission_group_rules UNIQUE (group_id, resource)
);
```

---

#### `staff_permission_groups` — スタッフ・グループ中間テーブル

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `staff_id` | `BIGINT` | `FK → staffs.id NOT NULL` | スタッフID |
| `group_id` | `BIGINT` | `FK → permission_groups.id NOT NULL` | グループID |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 作成日時 |

```sql
CREATE TABLE staff_permission_groups (
  staff_id  BIGINT NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
  group_id  BIGINT NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
  PRIMARY KEY (staff_id, group_id),
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_staff_permission_groups_staff ON staff_permission_groups(staff_id);
CREATE INDEX idx_staff_permission_groups_group ON staff_permission_groups(group_id);
```

> **廃止済み**: 旧 `user_permission_groups`（user_accounts ベース）は `staff_permission_groups`（staffs ベース）に置換。
> 権限制御は `permission_groups` + `permission_group_rules` + `staff_permission_groups` の3テーブルで行う。

---

#### `refresh_tokens` — リフレッシュトークン管理

| カラム | 型 | 制約 | 説明 |
|--------|-----|------|------|
| `id` | `BIGSERIAL` | `PK` | レコードID |
| `user_id` | `BIGINT` | `FK → user_accounts.id NOT NULL` | ユーザーID |
| `token_hash` | `TEXT` | `NOT NULL UNIQUE` | SHA-256 ハッシュ（平文は Cookie のみに存在） |
| `expires_at` | `TIMESTAMPTZ` | `NOT NULL` | 有効期限 |
| `revoked_at` | `TIMESTAMPTZ` | `NULL` | 無効化日時（NULL = 有効） |
| `user_agent` | `TEXT` | | ブラウザ・デバイス識別用 |
| `ip_address` | `INET` | | 発行時の IP アドレス |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` | 発行日時 |

```sql
CREATE TABLE refresh_tokens (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,
  user_agent  TEXT,
  ip_address  INET,
  created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
-- 期限切れ・失効済みトークンの自動削除（バッチ or pg_cron）
CREATE INDEX idx_refresh_tokens_cleanup ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
```

> **平文トークンの扱い**: Cookie には平文を格納、DB には `SHA-256(token)` のみ保存。
> DB が漏洩してもトークン値は利用不可（レインボーテーブル対策として十分なエントロピーが前提）。

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
| スタッフは `master_items` テーブルの `category='staff'` レコードとして管理 | `user_accounts` テーブルが認証の主テーブル。`staff_id` で `staffs` テーブルと紐付け（実装済み） |
| `staff_role` ENUM で `staffs` テーブルの職種を管理（実装済み） | `occupations` FK テーブルで職種名を管理（実装済み）。`occupation` ENUM への置換は未実施 |
| シフト管理は `master_items.id` を `shift_entries.staff_id` として参照 | 段階的に `user_accounts.id` ベースに移行。移行期間中は `staffs.id`（`user_accounts.staff_id`）経由で互換性維持 |

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
CREATE INDEX idx_user_accounts_staff_id ON user_accounts(staff_id);
CREATE INDEX idx_user_accounts_occupation_id ON user_accounts(occupation_id);

-- ===== user_clinic_memberships =====
-- UNIQUE(user_id, clinic_id) が複合インデックスを暗黙作成
CREATE INDEX idx_user_clinic_memberships_clinic ON user_clinic_memberships(clinic_id);

-- ===== permission_groups =====
-- TASK-049 完了後: clinic_id → company_id
CREATE INDEX idx_permission_groups_company ON permission_groups(company_id);

-- ===== permission_group_rules =====
-- UNIQUE(group_id, resource) が複合インデックスを暗黙作成

-- ===== staff_permission_groups =====
-- PRIMARY KEY(staff_id, group_id) が複合インデックスを暗黙作成
CREATE INDEX idx_staff_permission_groups_staff ON staff_permission_groups(staff_id);
CREATE INDEX idx_staff_permission_groups_group ON staff_permission_groups(group_id);

-- ===== 既存テーブルの clinic_id インデックス =====
CREATE INDEX idx_owners_clinic ON owners(clinic_id);
CREATE INDEX idx_pets_clinic ON pets(clinic_id);
CREATE INDEX idx_medical_records_clinic ON medical_records(clinic_id);
CREATE INDEX idx_hospitalizations_clinic ON hospitalizations(clinic_id);
CREATE INDEX idx_reservation_appointments_clinic ON reservation_appointments(clinic_id);
CREATE INDEX idx_trimming_records_clinic ON trimming_records(clinic_id);
CREATE INDEX idx_accountings_clinic ON accountings(clinic_id);
-- master_items テーブルは廃止済み（16専用マスタテーブルに分割）
CREATE INDEX idx_inventory_items_clinic ON inventory_items(clinic_id);
CREATE INDEX idx_shift_entries_clinic ON shift_entries(clinic_id);
```

---

## アプリケーション層認可設計

本システムは Go/Gin + PostgreSQL 構成であり、認可はアプリケーション層（Gin ミドルウェア + ハンドラー）で実施する。
DB 直接アクセスがないため PostgreSQL RLS は採用しない。

### 認証ミドルウェア (`internal/middleware/auth.go`)

全保護エンドポイントに適用。JWT を検証し `user_id`, `clinic_id`, `user_type` を Gin コンテキストに格納する。

```go
type JWTClaims struct {
    UserID   string `json:"user_id"`
    ClinicID string `json:"clinic_id"`
    UserType string `json:"user_type"`
    jwt.RegisteredClaims
}

func Auth(secret string) gin.HandlerFunc {
    key := []byte(secret)
    return func(c *gin.Context) {
        var tokenStr string

        // 1. access_token Cookie を優先して読む（XSS耐性あり）
        if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
            tokenStr = cookie
        }

        // 2. 後方互換: 旧Cookie名 auth_token にフォールバック
        if tokenStr == "" {
            if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
                tokenStr = cookie
            }
        }

        // 3. Cookie がなければ Authorization Bearer ヘッダにフォールバック
        if tokenStr == "" {
            if authHeader := c.GetHeader("Authorization"); authHeader != "" {
                parts := strings.SplitN(authHeader, " ", 2)
                if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
                    tokenStr = parts[1]
                }
            }
        }

        if tokenStr == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
            c.Abort()
            return
        }

        claims := &JWTClaims{}
        if _, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return key, nil
        }); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }

        c.Set("user_id",   claims.UserID)
        c.Set("clinic_id", claims.ClinicID)
        c.Set("user_type", claims.UserType)
        c.Next()
    }
}
```

### マルチテナント強制

全ハンドラーで `clinic_id` を JWT から取得し、クエリの WHERE 句に含める。
**リクエストボディや URL パラメータの `clinic_id` は信頼しない。**

```go
// ✅ 正しい: JWT から取得
clinicID := c.GetString("clinic_id")
owners, err := h.service.ListOwners(ctx, clinicID)

// ❌ 危険: クライアント入力を信頼
clinicID := c.Query("clinic_id")  // 他クリニックの clinic_id を指定可能
```

### リソース・アクション別認可ミドルウェア

`staff` ユーザーに対してリソース×アクション単位でアクセス可否を検証する。

```go
// RequirePermission: resource と action を指定したミドルウェア
func RequirePermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userType := c.GetString("user_type")
        // system_admin / clinic_admin は全権限バイパス
        if userType == "system_admin" || userType == "clinic_admin" {
            c.Next()
            return
        }
        // staff: DB から実効権限を取得して確認（権限は company スコープ）
        userID := c.GetString("user_id")
        allowed, err := checkPermission(c.Request.Context(), userID, resource, action)
        if err != nil || !allowed {
            RespondError(c, fmt.Errorf("%s/%s: %w", resource, action, ErrForbidden))
            c.Abort()
            return
        }
        c.Next()
    }
}

// ルート定義での使用例
protected.GET("/medical-records",
    RequirePermission("medical-records", "view"),
    handler.ListMedicalRecords,
)
protected.POST("/medical-records",
    RequirePermission("medical-records", "create"),
    handler.CreateMedicalRecord,
)
```

### 実効権限の計算

`staff` ユーザーは複数の権限グループに所属できる。実効権限は `bool_or()` で UNION する。

```sql
-- TASK-049 完了後: company スコープでフラットに実効権限を計算
SELECT
    pgr.resource,
    bool_or(pgr.can_view)   AS can_view,
    bool_or(pgr.can_create) AS can_create,
    bool_or(pgr.can_edit)   AS can_edit,
    bool_or(pgr.can_delete) AS can_delete
FROM staff_permission_groups spg
JOIN permission_groups pg ON pg.id = spg.group_id
    AND pg.deleted_at IS NULL
    AND pg.is_active = true
JOIN permission_group_rules pgr ON pgr.group_id = pg.id
WHERE spg.staff_id = $1
GROUP BY pgr.resource
```

### 認可チェックの責務分離

| 層 | 責務 |
|---|---|
| **Middleware** | JWT 検証、`clinic_id` / `user_type` の Gin コンテキスト格納 |
| **RequirePermission** | リソース×アクション単位の事前チェック（ルート定義時に宣言） |
| **Handler** | パスパラメータと JWT の `clinic_id` 一致検証（サブリソース保護） |
| **Repository** | 全クエリに `clinic_id` を必須フィルタとして適用 |

```go
// Handler でのサブリソース保護例 (GET /owners/:id の所有権確認)
func (h *OwnerHandler) Get(c *gin.Context) {
    ownerID  := c.Param("id")
    clinicID := c.GetString("clinic_id")  // JWT から取得

    owner, err := h.service.GetOwner(c.Request.Context(), clinicID, ownerID)
    if err != nil {
        RespondError(c, err)  // clinic_id 不一致 → ErrNotFound として処理
        return
    }
    c.JSON(http.StatusOK, toOwnerResponse(owner))
}
```

> **設計判断**: clinic_id 不一致時は `403 Forbidden` ではなく `404 Not Found` を返す。
> 他クリニックのリソース存在の有無を推測させないための情報漏洩対策。

---

## フロントエンド実装方針

### 実装済みファイル構成

```
features/auth/
├── api/
│   ├── login.ts              # ログインAPI（POST /v1/login）
│   ├── logout.ts             # ログアウトAPI（POST /v1/logout）
│   ├── refresh-token.ts      # セッション復元（GET /v1/me）
│   └── types.ts              # API 型定義
├── components/
│   └── LoginForm.tsx          # ログインフォーム
├── hooks/
│   ├── use-auth.tsx           # AuthContext + useAuth() フック
│   └── use-permission.ts      # usePermission(resource) フック
└── routes/
    └── Login.tsx              # ログインページ
```

### 型定義（実装済み）

```typescript
// ユーザー種別（models.ts の定数を使用）
export const USER_TYPE_VALUES = ["system_admin", "clinic_admin", "staff"] as const;
export type UserType = (typeof USER_TYPE_VALUES)[number];

// 職種（staff_role ENUM に対応。スタッフマスタが紐づく場合のみ非null）
export const STAFF_ROLE_VALUES = ["veterinarian", "nurse", "trimmer", "reception", "manager"] as const;
export type StaffRole = (typeof STAFF_ROLE_VALUES)[number];

// CRUD アクション（ResourcePermission のキーと一致）
export type ResourceAction = "view" | "create" | "edit" | "delete";

// 1リソースの CRUD 権限（フィールド名は view/create/edit/delete）
export interface ResourcePermission {
  view: boolean;
  create: boolean;
  edit: boolean;
  delete: boolean;
}

// resource → CRUD マップ
export type ResourcePermissions = Record<string, ResourcePermission>;

// clinicId → resource → CRUD（現在の実装。TASK-049 完了後フラット化予定）
export type ClinicEffectivePermissions = Record<string, ResourcePermissions>;

// クリニック所属情報
export interface ClinicMembership {
  clinicId: string;
  clinicName: string;
  isMain: boolean;
}

// ログインユーザー情報（/me レスポンス対応）
// permissions はバックエンドで実効権限を計算済みの状態で渡される
// フロントエンドで permissionGroups を再計算しない
export interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  userType: UserType;
  /** staff_role ENUM 値。スタッフマスタが紐づく場合のみ非null */
  staffRole: StaffRole | null;
  avatarUrl: string | null;
  mainClinicId: string;
  /** メイン医院の詳細情報 */
  clinic: AuthClinic | null;
  clinics: ClinicMembership[];
  // 現在: { clinicId → { resource → CRUD } }（TASK-049 完了後フラット化）
  permissions: ClinicEffectivePermissions;
}

// 認証コンテキスト
export interface AuthContextValue {
  user: AuthUser | null;
  currentClinicId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isSwitchingClinic: boolean;  // クリニック切替中フラグ
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  switchClinic: (clinicId: string) => Promise<void>;
  hasPermission: (resource: string, action: ResourceAction) => boolean;
}

// ⚠️ 廃止予定: バックエンドのグループ構造をフロントで持ち回す設計は非推奨
// 実効権限は permissions フィールドで参照すること
// export interface PermissionGroupWithRules { ... }
```

### 権限チェックパターン

`hasPermission` は company スコープのフラット `permissions` を直接参照する（TASK-049）。
`currentClinicId` によるスコープは不要 — 権限グループは全クリニック共通であるため。

```typescript
// use-auth.tsx 内の hasPermission 実装（TASK-049 完了後）
function hasPermission(
  resource: string,
  action: "view" | "create" | "edit" | "delete"
): boolean {
  if (!user) return false;
  // system_admin / clinic_admin は全権限バイパス
  if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
  // staff: company スコープの実効権限を参照（/me レスポンスで計算済み）
  const resourcePerms = user.permissions[resource];
  if (!resourcePerms) return false;
  switch (action) {
    case "view":   return resourcePerms.view;
    case "create": return resourcePerms.create;
    case "edit":   return resourcePerms.edit;
    case "delete": return resourcePerms.delete;
    default:       return false;
  }
}
```

> ⚠️ 現在の実装（TASK-049 未完了）: `permissions` が `{ clinicId → { resource → CRUD } }` の
> ネスト構造になっており `currentClinicId` でスコープしている。
> TASK-049（BE-082 + FE-139）完了後にフラット構造へ移行する。

**バックエンドの `/me` レスポンス構造（TASK-049 完了後）:**

```typescript
// permissions は { resource → CRUD } のフラット構造で返す（company スコープ）
// フロント側でグループを再計算しない（実効権限はバックエンドで計算済みを渡す）
interface AuthUser {
  id: string;
  email: string;
  displayName: string;
  userType: UserType;
  staffRole: StaffRole | null;
  avatarUrl: string | null;
  mainClinicId: string;
  clinic: AuthClinic | null;
  clinics: ClinicMembership[];
  permissions: ResourcePermissions; // TASK-049完了後: { resource → CRUD } フラット構造
}
```

### 認証ハイドレーション (React 19)

FOUC (Flash of Unauthenticated Content) を防ぐため、React 19 の `use()` フックを使用した Suspense ベースのハイドレーションを採用しています。

```typescript
// features/auth/hooks/use-auth.tsx
const initialAuthPromise = refreshToken().catch(() => null);

export function AuthProvider({ children }: AuthProviderProps) {
  // 初期チェックが完了するまでレンダリングをサスペンド
  const initialResult = use(initialAuthPromise);
  
  // ...初期値を state にセット
}
```

このパターンにより、アプリケーション起動時に「一瞬未ログイン画面が見える」といった現象を完全に排除しています。

### `usePermission(resource)` フック

```typescript
// features/auth/hooks/use-permission.ts
import { useAuth } from "@/features/auth/hooks/use-auth";

// usePermission の戻り値型（ResourcePermission とは別に can- prefix で公開）
export interface UsePermissionResult {
  canView: boolean;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

export function usePermission(resource: string): UsePermissionResult {
  const { hasPermission } = useAuth();
  return {
    canView:   hasPermission(resource, "view"),
    canCreate: hasPermission(resource, "create"),
    canEdit:   hasPermission(resource, "edit"),
    canDelete: hasPermission(resource, "delete"),
  };
}

// 使用例
const { canView, canCreate, canEdit, canDelete } = usePermission("medical-records");

// 閲覧権限チェック
if (!canView) return <AccessDenied />;

// 権限に応じてボタン表示/非表示（&& は禁止、三項演算子を使う）
{canCreate ? <Button onClick={handleCreate}>新規登録</Button> : null}
{canDelete ? <Button onClick={handleDelete}>削除</Button> : null}
```

### 権限ガード（コンポーネントレベル）

`RequirePermission` コンポーネントでページ・セクション単位のアクセス制御を宣言的に記述する。

```typescript
// components/shared/RequirePermission.tsx
interface RequirePermissionProps {
  resource: string;
  action?: "view" | "create" | "edit" | "delete";
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export function RequirePermission({
  resource,
  action = "view",
  children,
  fallback,
}: RequirePermissionProps) {
  const { hasPermission } = useAuth();
  if (!hasPermission(resource, action)) {
    return fallback ? <>{fallback}</> : <AccessDenied />;
  }
  return <>{children}</>;
}

// ルートへの適用（app/router.tsx）
{
  path: "medical-records",
  element: (
    <RequirePermission resource="medical-records">
      <Outlet />
    </RequirePermission>
  ),
  children: [...],
}
```

### ルート定義と必要権限一覧

| ルート | コンポーネント | 必要権限 |
|---|---|---|
| `/login` | `Login` | 認証不要 |
| `/` | `Reception` | `reception.view` |
| `/reservations` | `ReservationsList` | `reservations.view` |
| `/owners` | `OwnersList` | `owners.view` |
| `/owners/new` | `OwnerFormPage` | `owners.create` |
| `/owners/:id` | `OwnerFormPage` | `owners.view` |
| `/medical-records` | `MedicalRecordsList` | `medical-records.view` |
| `/medical-records/:id` | `MedicalRecordsForm` | `medical-records.view` |
| `/hospitalization` | `HospitalizationList` | `hospitalization.view` |
| `/trimming` | `TrimmingList` | `trimming.view` |
| `/examinations` | `ExaminationsList` | `examinations.view` |
| `/accounting` | `AccountingList` | `accounting.view` |
| `/vaccinations` | `VaccinationsList` | `vaccinations.view` |
| `/checkups` | `CheckupsList` | `checkups.view` |
| `/inventory` | `InventoryList` | `inventory.view` |
| `/estimates` | `EstimatesList` | `estimates.view` |
| `/shifts` | `ShiftCalendar` | `shifts.view` |
| `/settings/master` | `MasterSettingsIndex` | — （各サブページで個別制御） |
| `/settings/animal-species` | `AnimalSpeciesSettings` | `master-animal-species.view` |
| `/settings/treatment-items` | `TreatmentPlanMaster` | `master-medical.view` |
| `/settings/diagnosis` | `DiagnosisSettings` | `master-medical.view` |
| `/settings/inquiry-templates` | `InterviewTemplateSettings` | `master-medical.view` |
| `/settings/medicine` | `MedicineSettings` | `master-medical.view` |
| `/settings/service-type` | `ServiceTypeSettings` | `master-service-type.view` |
| `/settings/hospitalization` | `HospitalizationSettings` | `master-hospitalization.view` |
| `/settings/trimming-*` | `TrimmingSettings` | `master-trimming.view` |
| `/settings/permission-groups` | `PermissionGroupSettings` | `master-permission.view` |
| `/settings/staff` | `StaffSettings` | `master-staff.view` |
| `/settings/insurance` | `InsuranceSettings` | `master-insurance.view` |
| `/settings/merchandise-items` | `MerchandiseItemSettings` | `master-merchandise.view` |
| `/settings/clinic` | `ClinicMasterSettings` | `hospital-settings.view` |

---

## 実装状態と残課題

### 実装済み

| 項目 | 状態 |
|---|---|
| JWT + httpOnly Cookie ログイン / ログアウト | ✅ 実装済み |
| `/me` API によるセッション復元 | ✅ 実装済み |
| RBAC 3層モデル（UserType / StaffRole / PermissionGroup） | ✅ 実装済み |
| 権限グループ CRUD（マスタ設定画面） | ✅ 実装済み |
| `RequirePermission` コンポーネント | ✅ 実装済み |
| `usePermission` フック | ✅ 実装済み |
| サイドバーメニューの権限フィルタリング | ✅ 実装済み |
| マルチクリニック所属・切替 | ✅ 実装済み |
| 認証関連テーブル（6テーブル） | ✅ 実装済み |
| デモアカウント・シードデータ（7アカウント） | ✅ 実装済み |

### 残課題（セキュリティ優先度順）

| 課題 | チケット | 優先度 | 説明 |
|---|---|---|---|
| **論理削除ユーザーがログイン可能** | BUG-063 | 🔴 高 | `FindByEmail` に `deleted_at IS NULL` フィルタなし。削除済みユーザーが認証通過する |
| **アカウント停止後も JWT が有効期限まで通過** | BUG-061 | 🔴 高 | ミドルウェアが DB の `account_status` を確認しない。停止後最大 24h アクセス継続 |
| **dual-token 移行** | BUG-055 / BE-078 | 🔴 高 | 現在の 24h 単一 JWT をアクセス(15分) + リフレッシュ(7日) に分離。`refresh_tokens` テーブル追加が必要 |
| **パスワード変更エンドポイント未実装** | BUG-062 | 🔴 高 | `PUT /v1/users/me/password` がない。初期パスワードを使い続けるリスク |
| **権限スコープ company 単位への移行** | TASK-049 / BE-082 / FE-139 | 🔴 高 | `permission_groups.clinic_id` → `company_id`。`permissions` レスポンスのフラット化 |
| **バックエンド認可ミドルウェア未実装** | BUG-056 / BE-080 | 🔴 高 | `RequirePermission` ミドルウェアが未適用。全エンドポイントが `staff` に対して無認可でアクセス可能 |
| **パスワードリセットフロー未実装** | BUG-060 / BE-081 / FE-138 | 🟡 中 | `forgot-password` / `reset-password` エンドポイントなし |
| **権限変更の即時反映なし** | BUG-057 / FE-137 | 🟡 中 | 権限グループ変更後、セッション中のユーザーに即時反映されない。`useMe` refetchInterval 5分 + refetchOnWindowFocus で対応予定 |
| **`RequirePermission` をバックエンドと二重チェック** | — | 🟡 中 | フロントのガードは UX 目的。バックエンドの `RequirePermission` ミドルウェアが主防衛線 |
| **ログイン試行回数制限** | — | 🟡 中 | ブルートフォース対策。5回失敗で 15分ロック（`account_status = 'locked'`） |
| **監査ログ** | BE-079 | 🟢 低 | ログイン・ログアウト・権限変更イベントを別テーブルに記録 |

---

## 備考・設計判断

1. **`staff_role` と `occupations` の関係**: `user_accounts.occupation_id` は `occupations` テーブルへの FK（ENUM ではなく DB テーブル管理）。`staff_role` ENUM（`veterinarian|nurse|trimmer|reception|manager`）は `staffs.staff_role` カラムに格納（`user_accounts` には存在しない）。`user_accounts.staff_id` FK で `staffs` レコードと紐付け、`staffs.staff_role` を `/me` レスポンスの `staffRole` フィールドとして返す。`manager` ロールは現在も migration に存在するが、`user_type = clinic_admin` に対応するロール。一般スタッフは `staff_role` + 権限グループで権限管理。

2. **パスワードハッシュ**: `user_accounts.password_hash` に bcrypt (cost=10) で保存。シードデータは `$2a$10$...` 形式。bcrypt は今後も推奨（argon2id はメモリ要件が高く Docker 環境でリスクあり）。

3. **権限グループの companyスコープ（TASK-049）**: `permission_groups` は `company_id` で company 単位に管理する方針（現在は `clinic_id` で実装中）。権限グループ設定は全クリニック共通となり、クリニックごとに異なるグループ割り当ては廃止。データアクセスの clinic_id 分離（マルチテナント）は維持。`staff_permission_groups` は staff_id + group_id のみで管理。

4. **マスタ設定の個別リソース分割**: 旧 `master`（単一リソース）は `master-animal-species`, `master-medical`, `master-service-type`, `master-hospitalization`, `master-trimming`, `master-permission`, `master-staff`, `master-insurance`, `master-merchandise` の9個に分割済み。サイドナビの各マスタ項目に個別の `canView` 権限チェックが適用される。`master-permission` は「一般」グループで `canView=false`（権限グループ管理画面は非表示）。

5. **`system_admin` / `clinic_admin` の暗黙的全権限**: フロント・バックエンド両方で `userType` チェックにより全リソース全アクションをバイパス。`clinic_admin` は所属クリニック内のみ。

6. **`404 vs 403` の設計判断**: 他クリニックリソースへのアクセス時は `403 Forbidden` ではなく `404 Not Found` を返す。リソースの存在有無を他クリニックに推測させない情報漏洩対策。

7. **WCAG AA 準拠**: ログインフォームは `aria-describedby` でエラー紐付け、パスワード入力は表示/非表示トグル付き。

8. **印刷・帳票とクリニック情報**: 領収書・処方箋等は `currentClinicId` ベースの `clinics` テーブルデータを使用。クリニック切替後は帳票の発行元も自動的に切り替わる。

9. **`switchClinic` 後の権限リロード不要**: 権限は company スコープのフラット構造のため、クリニック切替時に追加 API コールは不要。`currentClinicId` の変更は UI 表示切替のみに影響し、`hasPermission` の結果は変わらない（TASK-049 完了後）。

---

## 関連ドキュメント

| ドキュメント | パス | 関連箇所 |
|---|---|---|
| **仕様定義書** | `docs/SPECIFICATION.md` | Feature一覧、ルーティング構成、ロードマップ |
| **画面仕様書** | `docs/screens/README.md` | 全ルートの画面仕様（`/login` §21 参照） |
| **ER図** | `docs/ERD.md` | 全テーブル定義・ENUM型・インデックス（v27.0、57テーブル） |
| **デザインシステム** | `docs/DESIGN_SYSTEM.md` | ログインページのUI仕様（§15 Login参照） |
