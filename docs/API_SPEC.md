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

**レスポンス (200 OK)**:
```json
{
  "access_token": "new-jwt-token"
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
- `POST /api/v1/login` — ログイン
- `POST /api/v1/logout` — ログアウト
- `POST /api/v1/auth/refresh` — トークン更新
- `POST /api/v1/auth/forgot-password` — パスワード忘却
- `POST /api/v1/auth/reset-password` — パスワード再設定

### 飼主・ペット
- `GET /api/v1/owners` — 飼主一覧
- `GET /api/v1/owners/:id` — 飼主詳細
- `POST /api/v1/owners` — 飼主作成
- `PATCH /api/v1/owners/:id` — 飼主更新
- `DELETE /api/v1/owners/:id` — 飼主削除
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
