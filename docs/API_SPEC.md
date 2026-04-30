# Animal Ekarte API 仕様書

> **バージョン**: 2.0  
> **最終更新**: 2026-04-23  
> **ステータス**: ✅ TIER 3 — API ドキュメント統合フェーズ

---

## 📋 概要

Animal Ekarte は、動物病院向け電子カルテシステムです。本 API ドキュメントは、バックエンド Go/Gin API の主要エンドポイントを記述します。

### 環境構成

| 環境 | Base URL |
|------|----------|
| **ローカル開発** | `http://localhost:8080` |
| **ステージング** | `https://stg.noah-karte.com` |
| **本番環境** | `https://api.noah-karte.com` |

### API バージョン

- **Current**: `/api/v1`
- **認証**: JWT (HttpOnly Cookie) + RBAC
- **マルチテナント切替**: `X-Clinic-ID` ヘッダで所属クリニック切替（保護エンドポイントのみ評価。詳細は OpenAPI `components/parameters/XClinicIDHeader` を参照）

---

## 🔐 認証

### ログイン

```http
POST /api/v1/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "password"
}
```

**レスポンス (200 OK)**:
```json
{
  "id": "user-uuid",
  "email": "admin@example.com",
  "first_name": "太郎",
  "last_name": "山田",
  "role_name": "一般",
  "clinic_id": 1,
  "clinic_name": "Noah 動物病院",
  "permissions": [
    "owners:read",
    "owners:create",
    "owners:edit",
    "owners:delete",
    ...
  ]
}
```

**レスポンス (401 Unauthorized)**:
```json
{
  "error": "メールアドレスまたはパスワードが違います"
}
```

**レート制限**: 5回/分（BUG-130 ブルートフォース対策）

---

### ログアウト

```http
POST /api/v1/logout
```

**レスポンス (200 OK)**:
```json
{
  "message": "logged out"
}
```

---

### トークン更新

```http
POST /api/v1/auth/refresh
```

`refresh_token` Cookie を検証し、新しい `access_token`（15分）と `refresh_token`（7日）を発行する。Token Rotation により使用済み refresh_token は破棄される。

**レスポンス (200 OK)**:
```json
{
  "message": "token refreshed"
}
```

**レスポンス (401 Unauthorized)** — refresh_token 不在 / 検証失敗 / staff 無効:
```json
{
  "error": "invalid or expired refresh token"
}
```

---

### パスワード忘却

```http
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

リセットトークンを発行しメール送信する。アカウント不存在でも 200 を返す（メール存在有無の漏洩防止）。

**レスポンス (200 OK)**:
```json
{
  "message": "If the email exists, a reset link has been sent."
}
```

**レート制限**: 3回/分

---

### パスワード再設定

```http
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "raw-reset-token",
  "password": "NewP@ssw0rd"
}
```

`token` を検証し、新しいパスワードに更新する。`password` は **8文字以上 / 英字+数字を含む** 必要がある（BUG-139）。

**レスポンス (200 OK)**:
```json
{
  "message": "Password has been reset successfully."
}
```

**レスポンス (400 Bad Request)** — トークン不正 / 期限切れ / パスワード複雑性違反:
```json
{
  "error": "パスワードは英字と数字の両方を含めてください"
}
```

**レート制限**: 3回/分

---

### 自分のパスワード変更

```http
PUT /api/v1/users/me/password
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "current_password": "CurrentP@ss",
  "new_password": "NewP@ssw0rd"
}
```

認証済みユーザーが自分のパスワードを変更する（BUG-148）。`current_password` を bcrypt で検証してから `new_password` に更新する。`new_password` は **8文字以上 / 英字+数字を含む** 必要がある（BUG-139）。

**レスポンス (200 OK)**:
```json
{
  "message": "パスワードを変更しました"
}
```

**レスポンス (401 Unauthorized)** — `current_password` 不一致:
```json
{
  "error": "現在のパスワードが正しくありません"
}
```

**レスポンス (400 Bad Request)** — 新パスワード複雑性違反:
```json
{
  "error": "パスワードは8文字以上で入力してください"
}
```

---

### ログインユーザー情報取得

```http
GET /api/v1/me
Authorization: Bearer {jwt_token}
```

JWT で認証されているユーザーの情報・所属クリニック・実効権限を返す。

**レスポンス (200 OK)**:
```json
{
  "id": "1",
  "email": "admin@example.com",
  "display_name": "山田 太郎",
  "is_system_admin": false,
  "occupation": "獣医師",
  "main_clinic_id": "1",
  "clinic": { "id": "1", "name": "Noah 動物病院" },
  "clinics": [
    { "clinic_id": "1", "clinic_name": "Noah 動物病院", "is_main": true }
  ],
  "permissions": {
    "owners": { "view": true, "create": true, "edit": true, "delete": true }
  }
}
```

---

## 🏥 主要リソース

### 1. 飼主管理 (Owners)

#### 飼主一覧取得

```http
GET /api/v1/owners?page=1&limit=20
Authorization: Bearer {jwt_token}
```

**クエリパラメータ**:

| パラメータ | 型 | 説明 |
|----------|-----|------|
| `page` | integer | ページ番号 (デフォルト: 1) |
| `limit` | integer | 1ページあたりの件数 (デフォルト: 20) |
| `search` | string | 飼主名またはペット名で検索 |

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": 1,
      "owner_name": "山田 太郎",
      "owner_kana": "ヤマダ タロウ",
      "phone": "09012345678",
      "email": "yamada@example.com",
      "address": "東京都渋谷区",
      "pets": [
        {
          "id": "pet-001",
          "pet_name": "ポチ",
          "species": "dog",
          "breed": "柴犬",
          "birth_date": "2018-05-15",
          "weight": 12.5
        }
      ],
      "danger_level": "high"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20
}
```

**エラー (401 Unauthorized)**:
```json
{
  "error": "authentication required"
}
```

---

#### 飼主詳細取得

```http
GET /api/v1/owners/:owner_id
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "id": 1,
  "owner_name": "山田 太郎",
  "owner_kana": "ヤマダ タロウ",
  "phone": "09012345678",
  "email": "yamada@example.com",
  "home_address": "東京都渋谷区道玄坂1-2-3",
  "company_name": "ABC株式会社",
  "company_phone": "0312345678",
  "company_address": "東京都渋谷区セルリアンタワー",
  "birth_date": "1980-05-15",
  "is_danger": false,
  "notes": "いつも優しい飼主さん",
  "discount_rate": 5,
  "membership_type": "member",
  "pets": [...]
}
```

---

#### 飼主作成

```http
POST /api/v1/owners
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "owner_name": "山田 太郎",
  "owner_kana": "ヤマダ タロウ",
  "phone": "09012345678",
  "email": "yamada@example.com",
  "home_address": "東京都渋谷区",
  "birth_date": "1980-05-15",
  "is_danger": false,
  "notes": "特記事項",
  "discount_rate": 5,
  "membership_type": "member"
}
```

**必須フィールド**: `owner_name`, `owner_kana`, `phone`

**レスポンス (201 Created)**:
```json
{
  "id": 1,
  "owner_name": "山田 太郎",
  ...
}
```

**エラー (400 Bad Request)**:
```json
{
  "error": "owner_name is required"
}
```

**権限チェック**: `owners:create` (BUG-125 CRUD 個別ガード)

---

#### 飼主更新

```http
PATCH /api/v1/owners/:owner_id
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "owner_name": "山田 次郎",
  "discount_rate": 10
}
```

**レスポンス (200 OK)**:
```json
{
  "id": 1,
  "owner_name": "山田 次郎",
  ...
}
```

**権限チェック**: `owners:edit`

---

#### 飼主削除

```http
DELETE /api/v1/owners/:owner_id
Authorization: Bearer {jwt_token}
```

**レスポンス (204 No Content)**

**エラー (409 Conflict)** — 関連データ存在:
```json
{
  "error": "cannot delete: 3 pets are associated with this owner"
}
```

**権限チェック**: `owners:delete`

---

#### 顧客集計一覧取得

```http
GET /api/v1/clinics/:clinic_id/owners/aggregations?sort=annual_amount&order=desc&page=1&per_page=50
Authorization: Bearer {jwt_token}
```

飼い主単位で年間売上、来院回数、最終来院日を集計する一覧。

**主なクエリパラメータ**:

| パラメータ | 型 | 説明 |
|----------|-----|------|
| `metric` | string | `annual_sales` / `visit_count` / `last_visit` |
| `sort` | string | `annual_amount` / `period_visit_count` / `last_visit_date` |
| `order` | string | `asc` / `desc` |
| `amount_basis` | string | `gross_total_amount` / `paid_amount` / `net_paid_amount` |
| `min_amount` | integer | 年間売上下限 |
| `max_amount` | integer | 年間売上上限 |
| `search` | string | 飼い主名部分一致 |
| `year` | integer | 年間売上の対象年 |
| `from` | date | 集計開始日 |
| `to` | date | 集計終了日 |
| `period_preset` | string | `last_3_months` / `last_6_months` / `last_12_months` / `calendar_year` |
| `min_visit_count` | integer | 累計来院回数下限 |
| `max_visit_count` | integer | 累計来院回数上限 |
| `last_visit_bucket` | string | `within_3m` / `over_3m` / `over_6m` / `over_1y` / `no_visit` |
| `min_days_since_last_visit` | integer | 経過日数下限 |
| `max_days_since_last_visit` | integer | 経過日数上限 |
| `include_zero` | boolean | 0円・0回を含める |
| `include_no_visit` | boolean | 来院なしを含める |
| `page` | integer | ページ番号 |
| `per_page` | integer | 1ページ件数 |

**レスポンス (200 OK)**:
```json
{
  "total": 142,
  "page": 1,
  "per_page": 50,
  "owners": [
    {
      "owner_id": "1",
      "owner_name": "山田 太郎",
      "annual_amount": 156000,
      "billing_count": 12,
      "period_visit_count": 4,
      "total_visit_count": 12,
      "last_visit_date": "2026-03-10",
      "days_since_last_visit": 47,
      "last_visit_bucket": "over_3m",
      "first_visit_date": "2022-05-01"
    }
  ]
}
```

詳細仕様は [CUSTOMER_AGGREGATION_SPEC.md](./CUSTOMER_AGGREGATION_SPEC.md) を参照。

---

### 2. ペット管理 (Pets)

#### ペット一覧取得（飼主配下）

```http
GET /api/v1/owners/:owner_id/pets
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": "pet-001",
      "pet_no": "1-001",
      "pet_name": "ポチ",
      "species": "dog",
      "breed": "柴犬",
      "birth_date": "2018-05-15",
      "weight": 12.5,
      "microchip": "123456789",
      "is_alive": true,
      "notes": "気が強い"
    }
  ]
}
```

---

### 3. 予約管理 (Reservations)

#### 予約一覧取得

```http
GET /api/v1/appointments?clinic_id=1&date=2026-04-24&limit=100
Authorization: Bearer {jwt_token}
```

**クエリパラメータ**:

| パラメータ | 型 | 説明 |
|----------|-----|------|
| `clinic_id` | integer | 医院ID |
| `date` | string (YYYY-MM-DD) | 予約日付（この日付のみを返す） |
| `limit` | integer | 最大件数 |

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": "reservation-001",
      "clinic_id": 1,
      "pet_id": "pet-001",
      "owner_id": 1,
      "owner_name": "山田 太郎",
      "pet_name": "ポチ",
      "reservation_type": "初診",
      "staff_id": "staff-001",
      "staff_name": "倉田 春香",
      "start_time": "2026-04-24T10:00:00Z",
      "end_time": "2026-04-24T10:30:00Z",
      "notes": "特記事項",
      "status": "confirmed"
    }
  ],
  "total": 25
}
```

---

### 4. 医療記録 (Medical Records)

#### 医療記録一覧取得

```http
GET /api/v1/medical-records?page=1&limit=20
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": "record-001",
      "pet_id": "pet-001",
      "pet_name": "ポチ",
      "owner_name": "山田 太郎",
      "created_at": "2026-04-23T14:30:00Z",
      "status": "draft",
      "chief_complaint": "嘔吐",
      "diagnosis": "胃炎",
      "treatments": [
        {
          "treatment_item_id": 1,
          "name": "点滴",
          "quantity": 1,
          "unit_price": 5000
        }
      ]
    }
  ],
  "total": 340,
  "page": 1,
  "limit": 20
}
```

---

#### 医療記録詳細取得

```http
GET /api/v1/medical-records/:record_id
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "id": "record-001",
  "pet_id": "pet-001",
  "pet_name": "ポチ",
  "clinic_id": 1,
  "created_at": "2026-04-23T14:30:00Z",
  "status": "draft",
  "chief_complaint": "嘔吐",
  "examination": {
    "body_weight": 12.5,
    "temperature": 38.5,
    "pulse": 80,
    "findings": "腹部圧痛あり"
  },
  "diagnosis": "胃炎",
  "diagnosis_code": "ICD-001",
  "treatments": [
    {
      "id": "treatment-001",
      "treatment_item_id": 1,
      "name": "点滴",
      "quantity": 1,
      "unit_price": 5000,
      "total_price": 5000
    }
  ],
  "medicines": [
    {
      "medicine_id": 1,
      "name": "セファレキシン",
      "dosage": "500mg",
      "frequency": "1日2回",
      "days": 5
    }
  ],
  "total_amount": 15000,
  "created_by": "staff-001",
  "updated_at": "2026-04-23T15:00:00Z"
}
```

---

### 5. マスタ設定 (Masters)

#### 予約種別一覧取得

```http
GET /api/v1/settings/reservation-type
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "初診",
      "color": "#FF6B6B",
      "sort_order": 1,
      "is_active": true
    },
    {
      "id": 2,
      "name": "再診",
      "color": "#4ECDC4",
      "sort_order": 2,
      "is_active": true
    }
  ]
}
```

---

#### 処置項目一覧取得

```http
GET /api/v1/settings/treatment-items?category=診察
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": 1,
      "name": "初診料",
      "category": "診察",
      "unit_price": 3000,
      "tax_rate": 10,
      "is_active": true
    },
    {
      "id": 2,
      "name": "再診料",
      "category": "診察",
      "unit_price": 2000,
      "tax_rate": 10,
      "is_active": true
    }
  ]
}
```

---

#### スタッフ一覧取得

```http
GET /api/v1/settings/staff?clinic_id=1
Authorization: Bearer {jwt_token}
```

**レスポンス (200 OK)**:
```json
{
  "data": [
    {
      "id": "staff-001",
      "first_name": "春香",
      "last_name": "倉田",
      "email": "kurata@example.com",
      "occupation": "獣医師",
      "clinic_assignments": [
        {
          "clinic_id": 1,
          "clinic_name": "Noah 動物病院",
          "is_primary": true
        }
      ],
      "is_active": true
    }
  ]
}
```

---

### 6. Lステップ連携 — 健診対象者抽出 (Checkup Sync)

> **対象**: ISSUE-005 / ISSUE-008（健診抽出における除外理由対応）
> **ルート前提**: バックエンドは `/api/v1/...`、フロントエンドの axios baseURL は `/api` を含むため `/v1/...` で参照される。

#### 健診対象者プレビュー取得

```http
GET /api/v1/clinics/:clinic_id/lstep/checkup-sync/preview?checkup_type=annual&species=dog&last_visit_before=2025-10-01
Authorization: Bearer {jwt_token}
```

健診タグ一括付与の対象者候補を抽出し、送信可否（`exclusion_reason`）と集計サマリーを返す。

**クエリパラメータ**:

| パラメータ | 型 | 説明 |
|----------|-----|------|
| `checkup_type` | string | `annual` / `dental` / `blood` / `skin` / `cancer` / `other`（監査ログ用） |
| `species` | string | 動物種フィルタ |
| `last_visit_before` | date (YYYY-MM-DD) | 最終来院日がこの日以前 |
| `last_visit_after` | date (YYYY-MM-DD) | 最終来院日がこの日以降 |
| `min_age_years` | integer | **(ISSUE-009)** 生存ペットの最小年齢以上（少なくとも1匹該当） |
| `max_age_years` | integer | **(ISSUE-009)** 生存ペットの最大年齢以下（少なくとも1匹該当） |
| `has_chronic_condition` | boolean | **(ISSUE-009)** アクティブ慢性疾患の有無 |
| `cpm_stage` | string | **(ISSUE-009)** `cpm_encounter` / `cpm_growing` / `cpm_core` / `cpm_spot` / `cpm_noah` / `cpm_dormant` |
| `min_total_amount` | integer | **(ISSUE-009)** 累計診療費（円）以上（completed billings 合計） |
| `min_annual_visit_count` | integer | **(ISSUE-009)** 年間来院回数（過去365日 distinct visit）以上 |
| `last_checkup_before` | date (YYYY-MM-DD) | **(ISSUE-009)** 最終健診実施日がこの日以前 |
| `last_checkup_after` | date (YYYY-MM-DD) | **(ISSUE-009)** 最終健診実施日がこの日以降 |

**権限チェック**: `owners:view`

**レスポンス (200 OK)**:
```json
{
  "owners": [
    {
      "owner_id": "1",
      "owner_name": "山田 太郎",
      "pet_names": ["ポチ", "タマ"],
      "last_visit_date": "2025-09-15",
      "has_line": true,
      "is_opt_out": false,
      "has_living_pet": true,
      "exclusion_reason": null,
      "current_tags": ["健診_年次"]
    },
    {
      "owner_id": "2",
      "owner_name": "佐藤 花子",
      "pet_names": ["モモ"],
      "last_visit_date": "2025-08-01",
      "has_line": true,
      "is_opt_out": true,
      "has_living_pet": true,
      "exclusion_reason": "Lステップ配信停止中",
      "current_tags": []
    },
    {
      "owner_id": "3",
      "owner_name": "鈴木 一郎",
      "pet_names": ["クロ"],
      "last_visit_date": "2024-12-20",
      "has_line": true,
      "is_opt_out": false,
      "has_living_pet": false,
      "exclusion_reason": "生存ペットなし",
      "current_tags": []
    },
    {
      "owner_id": "4",
      "owner_name": "高橋 二郎",
      "pet_names": ["シロ"],
      "last_visit_date": "2025-07-10",
      "has_line": false,
      "is_opt_out": false,
      "has_living_pet": true,
      "exclusion_reason": "LINE未連携",
      "current_tags": []
    }
  ],
  "total_count": 4,
  "eligible_count": 1,
  "line_linked_count": 3,
  "opt_out_count": 1,
  "no_living_pet_count": 1
}
```

**フィールド仕様（オーナー単位）**:

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `owner_id` | string | 飼い主ID（数値の文字列表現） |
| `owner_name` | string | 飼い主名 |
| `pet_names` | string[] | 関連ペット名（生存・死亡を含む） |
| `last_visit_date` | string \| null | 最終来院日（YYYY-MM-DD）。履歴なしは `null` |
| `has_line` | boolean | LINE User ID 登録済みか |
| `is_opt_out` | boolean | Lステップ配信停止中か |
| `has_living_pet` | boolean | **(ISSUE-005)** 生存中ペットが1件以上あるか |
| `exclusion_reason` | string \| null | **(ISSUE-005)** 除外理由。`null` なら送信可能 |
| `current_tags` | string[] | 現在の Lステップタグ。LINE未連携時は `[]` |
| `min_pet_age_years` | integer \| null | **(ISSUE-009)** 生存ペットの最小年齢（years）。誕生日未登録時は `null` |
| `max_pet_age_years` | integer \| null | **(ISSUE-009)** 生存ペットの最大年齢（years）。誕生日未登録時は `null` |
| `has_chronic_condition` | boolean | **(ISSUE-009)** アクティブな慢性疾患の有無（生存ペット由来） |
| `cpm_stage` | string | **(ISSUE-009)** CPM ステージ（`CalculateCPMStage` と同一ロジック） |
| `total_amount` | integer | **(ISSUE-009)** 累計診療費（円、completed billings 合計） |
| `annual_visit_count` | integer | **(ISSUE-009)** 年間来院回数（過去365日 distinct visit） |
| `last_checkup_date` | string \| null | **(ISSUE-009)** 最終健診実施日。実績なしは `null` |

**フィールド仕様（全体集計）**:

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `total_count` | integer | 抽出条件マッチ総件数 |
| `eligible_count` | integer | **(ISSUE-005)** 送信可能件数（`exclusion_reason == null`） |
| `line_linked_count` | integer | LINE 連携済み件数 |
| `opt_out_count` | integer | **(ISSUE-005)** opt-out 中の件数 |
| `no_living_pet_count` | integer | **(ISSUE-005)** 生存ペットなしの件数 |

**`exclusion_reason` の取り得る値（優先順位順）**:

| 値 | 意味 | 優先度 |
|----|------|--------|
| `Lステップ配信停止中` | `is_opt_out == true` | 1 (最優先) |
| `生存ペットなし` | `has_living_pet == false` | 2 |
| `LINE未連携` | `has_line == false` | 3 |
| `null` | 送信可能（上記いずれにも該当せず） | — |

> **判定ロジック**: 上から順に評価し、最初に該当した理由が採用される（`backend/internal/service/checkup_sync_service.go::deriveExclusionReason`）。
> 例えば opt-out かつ LINE未連携の飼い主は `Lステップ配信停止中` のみが返される。

---

#### 健診タグ一括付与

```http
POST /api/v1/clinics/:clinic_id/lstep/checkup-sync
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "checkup_type": "annual",
  "owner_ids": ["1", "5", "12"],
  "tag_name": "健診_年次_2026"
}
```

**リクエスト**:

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|-----|------|
| `checkup_type` | string | ✅ | `annual` / `dental` / `blood` / `skin` / `cancer` / `other` |
| `owner_ids` | string[] | ✅ | 対象飼い主ID（最低1件） |
| `tag_name` | string | ✅ | `^[A-Za-z0-9_-]{1,100}$`。自動管理タグ名は不可 |

**権限チェック**: `owners:edit`

**レスポンス (200 OK)**:
```json
{
  "success_count": 2,
  "skipped_count": 1,
  "failed_count": 0,
  "failed_owner_ids": []
}
```

**スキップ判定**: プレビューの `exclusion_reason` と同じ優先度（`opt-out` > `生存ペットなし` > `LINE未連携`）で判定する。これにより API を直接叩かれた場合でも誤配信を防ぐ。

**エラー (400 Bad Request)** — `tag_name` パターン違反 or 自動管理タグ:
```json
{
  "error": "tag_name は英数字・アンダースコア・ハイフンのみ使用可能です（1〜100文字）"
}
```

---

## ⚙️ エラーハンドリング

### 標準エラーレスポンス

```json
{
  "error": "error message",
  "code": "ERROR_CODE",
  "request_id": "req-123456"
}
```

### エラーコード一覧

| Code | Status | 説明 |
|------|--------|------|
| `AUTHENTICATION_REQUIRED` | 401 | JWT トークンが無効または期限切れ |
| `PERMISSION_DENIED` | 403 | RBAC 権限不足 |
| `NOT_FOUND` | 404 | リソースが見つからない |
| `CONFLICT` | 409 | FK 参照制約違反（削除不可） |
| `INVALID_REQUEST` | 400 | バリデーションエラー |
| `RATE_LIMIT_EXCEEDED` | 429 | レート制限超過 |
| `INTERNAL_ERROR` | 500 | サーバー内部エラー |

---

## 🔗 エンドポイント一覧（サマリー）

### 認証関連
- `POST /api/v1/login` — ログイン（5回/分レート制限）
- `POST /api/v1/logout` — ログアウト
- `POST /api/v1/auth/refresh` — トークン更新（refresh_token Cookie 検証 + Rotation）
- `POST /api/v1/auth/forgot-password` — パスワード忘却（3回/分レート制限）
- `POST /api/v1/auth/reset-password` — パスワード再設定（3回/分レート制限）
- `GET /api/v1/me` — ログインユーザー情報・所属クリニック・実効権限取得
- `PUT /api/v1/users/me/password` — 自分のパスワード変更（BUG-148）

### 飼主・ペット
- `GET /api/v1/owners` — 飼主一覧
- `GET /api/v1/owners/:id` — 飼主詳細
- `POST /api/v1/owners` — 飼主作成
- `PATCH /api/v1/owners/:id` — 飼主更新
- `DELETE /api/v1/owners/:id` — 飼主削除
- `GET /api/v1/clinics/:clinic_id/owners/aggregations` — 顧客集計一覧
- `GET /api/v1/owners/:id/pets` — ペット一覧

### 予約
- `GET /api/v1/appointments` — 予約一覧
- `GET /api/v1/appointments/:id` — 予約詳細
- `POST /api/v1/appointments` — 予約作成
- `PATCH /api/v1/appointments/:id` — 予約更新
- `DELETE /api/v1/appointments/:id` — 予約削除

### 医療記録
- `GET /api/v1/medical-records` — 医療記録一覧
- `GET /api/v1/medical-records/:id` — 医療記録詳細
- `POST /api/v1/medical-records` — 医療記録作成
- `PATCH /api/v1/medical-records/:id` — 医療記録更新
- `DELETE /api/v1/medical-records/:id` — 医療記録削除

### 入院管理
- `GET /api/v1/hospitalizations` — 入院一覧
- `GET /api/v1/hospitalizations/:id` — 入院詳細
- `POST /api/v1/hospitalizations` — 入院作成
- `PATCH /api/v1/hospitalizations/:id` — 入院更新
- `DELETE /api/v1/hospitalizations/:id` — 入院削除
- `POST /api/v1/hospitalizations/:id/discharge-with-billing` — 退院・会計

### Lステップ連携 — 健診対象者抽出
- `GET /api/v1/clinics/:clinic_id/lstep/checkup-sync/preview` — 健診対象者プレビュー（`exclusion_reason` / 集計サマリー含む）
- `POST /api/v1/clinics/:clinic_id/lstep/checkup-sync` — 健診タグ一括付与

### マスタ設定
- `GET /api/v1/settings/reservation-type` — 予約種別
- `GET /api/v1/settings/treatment-items` — 処置項目
- `GET /api/v1/settings/staff` — スタッフ
- `GET /api/v1/settings/permission-groups` — 権限グループ
- `GET /api/v1/settings/animal-species` — 動物種
- `GET /api/v1/settings/medicine` — 医薬品
- `GET /api/v1/settings/cage` — 入院ケージ
- `GET /api/v1/settings/merchandise-items` — 物販商品

---

## 📊 パフォーマンス目標

### レスポンスタイム

| エンドポイント | 目標 | 説明 |
|-------------|------|------|
| 認証系 | < 300ms | ログイン・トークン更新 |
| 一覧取得 | < 500ms | ページネーション (limit: 20) |
| 詳細取得 | < 200ms | ID 指定取得 |
| 作成・更新 | < 500ms | POST/PATCH 処理 |
| 削除 | < 500ms | DELETE + FK チェック |

### 負荷テスト基準

| シナリオ | ユーザー数 | p95 | p99 | エラー率 |
|--------|---------|-----|-----|--------|
| 通常運用 | 10-50 | < 500ms | < 1000ms | < 10% |
| スパイク | 100 | < 2000ms | < 5000ms | < 20% |

---

## 🔄 レート制限

| エンドポイント | 制限 | 説明 |
|-------------|------|------|
| `/login` | 5回/分 | BUG-130 ブルートフォース対策 |
| `/auth/forgot-password` | 3回/分 | パスワード忘却メール |
| `/auth/reset-password` | 3回/分 | パスワード再設定 |
| その他 | 無制限 | JWT 認証済みのみ |

---

## 📝 注記

- すべての日時は UTC (ISO 8601) で返却
- 金額はセント（最小単位）で保存
- `clinic_id` はマルチテナント対応（BUG-145 clinic_id 強制）
- すべての write 操作に RBAC 権限チェック適用（BUG-125）
- NULL バイト除去済み（BUG-067）

---

## 🔗 関連ドキュメント

- [バックエンド開発ガイド](./architecture.md)
- [E2E テスティング](./E2E_TESTING_GUIDE.md)
- [負荷テスト](../load-tests/README.md)
- [パフォーマンスプロファイリング](./PERFORMANCE_PROFILING.md)
