# API ドキュメント

## 概要

Animal Ekarte（動物病院電子カルテシステム）のREST API仕様書です。

**Base URL:** `/api/v1`

> **Note**: 本ドキュメントは**目標設計（Target Design）**です。
> 現在実装されているAPIは Swagger UI (`http://localhost:8080/swagger/index.html`) を参照してください。

### 実装状況

| 状態 | API |
|------|-----|
| ✅ 実装済 | Pet CRUD (`/pets`), Health (`/health`) |
| 📋 設計のみ | その他（Owners, Medical Records, Reservations, etc.） |

---

## 認証

> 注: 認証機能は今後のロードマップに含まれています。

```
Authorization: Bearer <token>
```

---

## 共通レスポンス形式

### 成功レスポンス

```json
{
  "data": { ... },
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20
  }
}
```

### エラーレスポンス

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

### HTTPステータスコード

| コード | 説明 |
|--------|------|
| 200 | OK - 成功 |
| 201 | Created - 作成成功 |
| 204 | No Content - 削除成功 |
| 400 | Bad Request - リクエスト不正 |
| 401 | Unauthorized - 認証エラー |
| 403 | Forbidden - 権限エラー |
| 404 | Not Found - リソース不存在 |
| 409 | Conflict - 競合エラー |
| 422 | Unprocessable Entity - バリデーションエラー |
| 500 | Internal Server Error - サーバーエラー |

---

## エンドポイント一覧

### 飼い主 (Owners)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/owners` | 飼い主一覧取得 |
| GET | `/owners/:id` | 飼い主詳細取得 |
| POST | `/owners` | 飼い主作成 |
| PUT | `/owners/:id` | 飼い主更新 |
| DELETE | `/owners/:id` | 飼い主削除 |
| GET | `/owners/:id/pets` | 飼い主のペット一覧 |

### ペット (Pets)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/pets` | ペット一覧取得 |
| GET | `/pets/:id` | ペット詳細取得 |
| POST | `/pets` | ペット作成 |
| PUT | `/pets/:id` | ペット更新 |
| DELETE | `/pets/:id` | ペット削除 |
| GET | `/pets/search` | ペット検索 |

### 電子カルテ (Medical Records)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/medical-records` | カルテ一覧取得 |
| GET | `/medical-records/:id` | カルテ詳細取得 |
| POST | `/medical-records` | カルテ作成 |
| PUT | `/medical-records/:id` | カルテ更新 |
| DELETE | `/medical-records/:id` | カルテ削除 |
| POST | `/medical-records/:id/finalize` | カルテ確定 |
| GET | `/pets/:petId/medical-records` | ペットのカルテ一覧 |

### 予約 (Reservations)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/reservations` | 予約一覧取得 |
| GET | `/reservations/:id` | 予約詳細取得 |
| POST | `/reservations` | 予約作成 |
| PUT | `/reservations/:id` | 予約更新 |
| DELETE | `/reservations/:id` | 予約削除 |
| POST | `/reservations/:id/check-in` | 受付処理 |
| POST | `/reservations/:id/cancel` | 予約キャンセル |

### 入院 (Hospitalizations)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/hospitalizations` | 入院一覧取得 |
| GET | `/hospitalizations/:id` | 入院詳細取得 |
| POST | `/hospitalizations` | 入院作成 |
| PUT | `/hospitalizations/:id` | 入院更新 |
| DELETE | `/hospitalizations/:id` | 入院削除 |
| POST | `/hospitalizations/:id/discharge` | 退院処理 |

### ケアプラン (Care Plans)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/hospitalizations/:id/care-plans` | ケアプラン一覧 |
| POST | `/hospitalizations/:id/care-plans` | ケアプラン追加 |
| PUT | `/care-plans/:id` | ケアプラン更新 |
| DELETE | `/care-plans/:id` | ケアプラン削除 |

### デイリーレコード (Daily Records)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/hospitalizations/:id/daily-records` | デイリーレコード一覧 |
| GET | `/daily-records/:id` | デイリーレコード詳細 |
| POST | `/daily-records/:id/vitals` | バイタル記録追加 |
| POST | `/daily-records/:id/care-logs` | ケアログ追加 |
| POST | `/daily-records/:id/notes` | スタッフメモ追加 |

### ワクチン (Vaccinations)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/vaccinations` | ワクチン記録一覧 |
| GET | `/vaccinations/:id` | ワクチン記録詳細 |
| POST | `/vaccinations` | ワクチン記録作成 |
| PUT | `/vaccinations/:id` | ワクチン記録更新 |
| DELETE | `/vaccinations/:id` | ワクチン記録削除 |
| GET | `/vaccinations/due` | 接種予定一覧 |

### トリミング (Trimmings)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/trimmings` | トリミング一覧 |
| GET | `/trimmings/:id` | トリミング詳細 |
| POST | `/trimmings` | トリミング作成 |
| PUT | `/trimmings/:id` | トリミング更新 |
| DELETE | `/trimmings/:id` | トリミング削除 |

### 検査 (Examinations)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/examinations` | 検査一覧 |
| GET | `/examinations/:id` | 検査詳細 |
| POST | `/examinations` | 検査作成 |
| PUT | `/examinations/:id` | 検査更新 |
| DELETE | `/examinations/:id` | 検査削除 |
| POST | `/examinations/:id/complete` | 検査完了 |

### 会計 (Accountings)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/accountings` | 会計一覧 |
| GET | `/accountings/:id` | 会計詳細 |
| POST | `/accountings` | 会計作成 |
| PUT | `/accountings/:id` | 会計更新 |
| DELETE | `/accountings/:id` | 会計削除 |
| POST | `/accountings/:id/complete` | 会計完了 |
| GET | `/accountings/:id/receipt` | 領収書取得 |
| GET | `/accountings/:id/invoice` | 請求書取得 |

### マスタ (Master)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/master/items` | マスタ項目一覧 |
| GET | `/master/items/:id` | マスタ項目詳細 |
| POST | `/master/items` | マスタ項目作成 |
| PUT | `/master/items/:id` | マスタ項目更新 |
| DELETE | `/master/items/:id` | マスタ項目削除 |
| GET | `/master/categories` | カテゴリ一覧 |

### 在庫 (Inventory)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/inventory` | 在庫一覧 |
| GET | `/inventory/:id` | 在庫詳細 |
| POST | `/inventory` | 在庫作成 |
| PUT | `/inventory/:id` | 在庫更新 |
| DELETE | `/inventory/:id` | 在庫削除 |
| POST | `/inventory/:id/restock` | 入荷処理 |
| GET | `/inventory/low-stock` | 在庫不足一覧 |

### クリニック設定 (Clinic)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/clinic` | クリニック情報取得 |
| PUT | `/clinic` | クリニック情報更新 |

### ダッシュボード (Dashboard)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/dashboard/kanban` | カンバンボードデータ |
| PUT | `/dashboard/kanban/:appointmentId` | ステータス更新 |
| GET | `/dashboard/stats` | 統計情報 |

---

## API詳細

### 飼い主 API

#### GET /owners

飼い主一覧を取得します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| page | integer | ページ番号（デフォルト: 1） |
| per_page | integer | 1ページあたりの件数（デフォルト: 20） |
| search | string | 名前・電話番号での検索 |
| sort | string | ソート項目（name, created_at） |
| order | string | ソート順（asc, desc） |

**レスポンス:**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "山田 太郎",
      "name_kana": "ヤマダ タロウ",
      "phone": "090-1234-5678",
      "email": "yamada@example.com",
      "address": "東京都...",
      "pets_count": 2,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20
  }
}
```

#### POST /owners

飼い主を作成します。

**リクエストボディ:**

```json
{
  "name": "山田 太郎",
  "name_kana": "ヤマダ タロウ",
  "phone": "090-1234-5678",
  "email": "yamada@example.com",
  "address": "東京都..."
}
```

---

### ペット API

#### GET /pets

ペット一覧を取得します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| page | integer | ページ番号 |
| per_page | integer | 1ページあたりの件数 |
| owner_id | uuid | 飼い主IDでフィルタ |
| species | string | 種別でフィルタ（犬, 猫, etc.） |
| status | string | ステータスでフィルタ（生存, 死亡） |
| search | string | 名前・患者番号での検索 |

**レスポンス:**

```json
{
  "data": [
    {
      "id": "uuid",
      "owner_id": "uuid",
      "owner_name": "山田 太郎",
      "pet_number": "30042-008",
      "name": "ポチ",
      "species": "犬",
      "breed": "トイプードル",
      "gender": "オス",
      "birth_date": "2020-01-15",
      "weight": 5.2,
      "status": "生存",
      "insurance_name": "アニコム損保",
      "last_visit": "2024-03-01",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": { ... }
}
```

#### POST /pets

ペットを作成します。

**リクエストボディ:**

```json
{
  "owner_id": "uuid",
  "name": "ポチ",
  "species": "犬",
  "breed": "トイプードル",
  "gender": "オス",
  "birth_date": "2020-01-15",
  "weight": 5.2,
  "microchip_id": "123456789012345",
  "environment": "室内（散歩あり）",
  "insurance_name": "アニコム損保",
  "insurance_details": "70%プラン"
}
```

---

### 電子カルテ API

#### GET /medical-records

カルテ一覧を取得します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| page | integer | ページ番号 |
| per_page | integer | 1ページあたりの件数 |
| pet_id | uuid | ペットIDでフィルタ |
| owner_id | uuid | 飼い主IDでフィルタ |
| status | string | ステータス（作成中, 確定済） |
| date_from | date | 診察日（開始） |
| date_to | date | 診察日（終了） |
| doctor_id | uuid | 担当医でフィルタ |

**レスポンス:**

```json
{
  "data": [
    {
      "id": "uuid",
      "record_no": "M-2024-001",
      "pet_id": "uuid",
      "pet_name": "ポチ",
      "owner_id": "uuid",
      "owner_name": "山田 太郎",
      "species": "犬",
      "visit_date": "2024-03-20T10:00:00Z",
      "visit_type": "再診",
      "chief_complaint": "食欲不振",
      "doctor": "医師A",
      "status": "作成中",
      "created_at": "2024-03-20T10:00:00Z",
      "updated_at": "2024-03-20T10:30:00Z"
    }
  ],
  "meta": { ... }
}
```

#### GET /medical-records/:id

カルテ詳細を取得します（SOAPS形式）。

**レスポンス:**

```json
{
  "data": {
    "id": "uuid",
    "record_no": "M-2024-001",
    "pet": {
      "id": "uuid",
      "name": "ポチ",
      "species": "犬",
      "breed": "トイプードル",
      "weight": 5.2
    },
    "owner": {
      "id": "uuid",
      "name": "山田 太郎",
      "phone": "090-1234-5678"
    },
    "visit_date": "2024-03-20T10:00:00Z",
    "visit_type": "再診",
    "chief_complaint": "食欲不振、嘔吐が続く",
    "subjective": "3日前から食欲がない。嘔吐は1日2回程度。",
    "objective": "体温: 38.5℃、心拍数: 120bpm、呼吸数: 24/min",
    "assessment": "急性胃腸炎の疑い",
    "plan": "制吐剤投与、絶食指示、経過観察",
    "surgery_notes": null,
    "diagnosis": "急性胃腸炎",
    "treatment": "セレニア注射、補液",
    "prescription": "セレニア錠 1錠×5日分",
    "notes": "3日後に再診予定",
    "doctor_id": "uuid",
    "doctor_name": "医師A",
    "status": "確定済",
    "created_at": "2024-03-20T10:00:00Z",
    "updated_at": "2024-03-20T11:00:00Z"
  }
}
```

#### POST /medical-records

カルテを作成します。

**リクエストボディ:**

```json
{
  "pet_id": "uuid",
  "visit_type": "再診",
  "chief_complaint": "食欲不振",
  "doctor_id": "uuid"
}
```

#### PUT /medical-records/:id

カルテを更新します。

**リクエストボディ:**

```json
{
  "chief_complaint": "食欲不振、嘔吐",
  "subjective": "3日前から食欲がない",
  "objective": "体温: 38.5℃",
  "assessment": "急性胃腸炎の疑い",
  "plan": "制吐剤投与",
  "diagnosis": "急性胃腸炎",
  "treatment": "セレニア注射",
  "prescription": "セレニア錠 1錠×5日分",
  "notes": "3日後に再診"
}
```

#### POST /medical-records/:id/finalize

カルテを確定します。確定後は編集不可になります。

---

### 予約 API

#### GET /reservations

予約一覧を取得します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| date | date | 日付でフィルタ |
| date_from | date | 開始日 |
| date_to | date | 終了日 |
| status | string | ステータスでフィルタ |
| service_type | string | サービス種別でフィルタ |
| doctor_id | uuid | 担当者でフィルタ |

**レスポンス:**

```json
{
  "data": [
    {
      "id": "uuid",
      "pet_id": "uuid",
      "pet_name": "ポチ",
      "owner_id": "uuid",
      "owner_name": "山田 太郎",
      "start_time": "2024-03-20T10:00:00Z",
      "end_time": "2024-03-20T10:30:00Z",
      "visit_type": "revisit",
      "service_type": "診療",
      "doctor_id": "uuid",
      "doctor_name": "医師A",
      "is_designated": true,
      "status": "confirmed",
      "notes": "定期検診"
    }
  ],
  "meta": { ... }
}
```

#### POST /reservations

予約を作成します。

**リクエストボディ:**

```json
{
  "pet_id": "uuid",
  "start_time": "2024-03-20T10:00:00Z",
  "end_time": "2024-03-20T10:30:00Z",
  "visit_type": "revisit",
  "service_type": "診療",
  "doctor_id": "uuid",
  "is_designated": true,
  "notes": "定期検診"
}
```

#### POST /reservations/:id/check-in

予約を受付処理します。

**レスポンス:**

```json
{
  "data": {
    "id": "uuid",
    "status": "checked_in",
    "checked_in_at": "2024-03-20T09:55:00Z"
  }
}
```

---

### 入院 API

#### GET /hospitalizations

入院一覧を取得します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| status | string | ステータス（入院中, 退院済, 予約） |
| type | string | 種別（入院, ホテル） |
| date_from | date | 入院開始日（開始） |
| date_to | date | 入院開始日（終了） |

**レスポンス:**

```json
{
  "data": [
    {
      "id": "uuid",
      "hospitalization_no": "H-2024-001",
      "pet_id": "uuid",
      "pet_name": "ポチ",
      "owner_id": "uuid",
      "owner_name": "山田 太郎",
      "species": "犬",
      "type": "入院",
      "start_date": "2024-03-20",
      "end_date": "2024-03-25",
      "status": "入院中",
      "cage_id": "uuid",
      "cage_code": "A01"
    }
  ],
  "meta": { ... }
}
```

#### GET /hospitalizations/:id

入院詳細を取得します（ケアプラン、デイリーレコード含む）。

**レスポンス:**

```json
{
  "data": {
    "id": "uuid",
    "hospitalization_no": "H-2024-001",
    "pet": { ... },
    "owner": { ... },
    "type": "入院",
    "start_date": "2024-03-20",
    "end_date": "2024-03-25",
    "status": "入院中",
    "cage": {
      "id": "uuid",
      "code": "A01",
      "name": "大型犬用ケージA1"
    },
    "owner_request": "食事は少量ずつ",
    "staff_notes": "投薬時は注意",
    "care_plans": [
      {
        "id": "uuid",
        "type": "food",
        "name": "ロイヤルカナン消化器サポート",
        "description": "30g / ふやかして",
        "timing": ["morning", "night"],
        "status": "active",
        "unit_price": 150
      },
      {
        "id": "uuid",
        "type": "medicine",
        "name": "アンピシリン",
        "description": "1錠",
        "timing": ["morning", "night"],
        "status": "active",
        "unit_price": 500
      }
    ],
    "daily_records": [
      {
        "id": "uuid",
        "date": "2024-03-20",
        "vitals": [ ... ],
        "care_logs": [ ... ],
        "staff_notes": [ ... ]
      }
    ],
    "cost_summary": {
      "days": 5,
      "room_charge": 15000,
      "food_charge": 1500,
      "medicine_charge": 5000,
      "treatment_charge": 7500,
      "total": 29000
    }
  }
}
```

#### POST /hospitalizations

入院を作成します。

**リクエストボディ:**

```json
{
  "pet_id": "uuid",
  "type": "入院",
  "start_date": "2024-03-20",
  "end_date": "2024-03-25",
  "cage_id": "uuid",
  "owner_request": "食事は少量ずつ",
  "staff_notes": "投薬時は注意"
}
```

#### POST /hospitalizations/:id/discharge

退院処理を行います。

**リクエストボディ:**

```json
{
  "discharge_date": "2024-03-25",
  "discharge_notes": "経過良好、投薬継続"
}
```

---

### ケアプラン API

#### POST /hospitalizations/:id/care-plans

ケアプランを追加します。

**リクエストボディ:**

```json
{
  "type": "medicine",
  "name": "アンピシリン",
  "description": "1錠",
  "timing": ["morning", "night"],
  "master_id": "uuid",
  "unit_price": 500,
  "notes": "投薬後の様子を観察"
}
```

---

### デイリーレコード API

#### POST /daily-records/:id/vitals

バイタルを記録します。

**リクエストボディ:**

```json
{
  "time": "09:00",
  "temperature": 38.5,
  "heart_rate": 120,
  "respiration_rate": 24,
  "weight": 5.2,
  "notes": "元気あり"
}
```

#### POST /daily-records/:id/care-logs

ケアログを記録します。

**リクエストボディ:**

```json
{
  "time": "09:00",
  "type": "food",
  "status": "completed",
  "value": "100%",
  "notes": "完食"
}
```

---

### 会計 API

#### GET /accountings/:id

会計詳細を取得します。

**レスポンス:**

```json
{
  "data": {
    "id": "uuid",
    "medical_record_id": "uuid",
    "pet": { ... },
    "owner": { ... },
    "scheduled_date": "2024-03-20",
    "status": "waiting",
    "items": [
      {
        "id": "uuid",
        "code": "EX001",
        "category": "examination",
        "name": "再診料",
        "unit_price": 800,
        "quantity": 1,
        "tax_rate": 0.1,
        "is_insurance_applicable": true,
        "source": "medical_record"
      },
      {
        "id": "uuid",
        "code": "MD001",
        "category": "medicine",
        "name": "セレニア錠",
        "unit_price": 500,
        "quantity": 5,
        "tax_rate": 0.1,
        "is_insurance_applicable": true,
        "source": "medical_record"
      }
    ],
    "payment": {
      "subtotal": 3300,
      "tax_total": 330,
      "total_amount": 3630,
      "insurance_name": "アニコム損保",
      "insurance_ratio": 0.7,
      "insurance_amount": -2541,
      "discount_amount": 0,
      "billing_amount": 1089,
      "received_amount": 0,
      "change_amount": 0,
      "method": null
    }
  }
}
```

#### POST /accountings/:id/complete

会計を完了します。

**リクエストボディ:**

```json
{
  "received_amount": 1100,
  "method": "cash"
}
```

**レスポンス:**

```json
{
  "data": {
    "id": "uuid",
    "status": "completed",
    "completed_at": "2024-03-20T11:00:00Z",
    "payment": {
      "billing_amount": 1089,
      "received_amount": 1100,
      "change_amount": 11,
      "method": "cash"
    }
  }
}
```

---

### ダッシュボード API

#### GET /dashboard/kanban

カンバンボードのデータを取得します。

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| date | date | 対象日（デフォルト: 今日） |

**レスポンス:**

```json
{
  "data": {
    "columns": [
      {
        "id": "reception_reserved",
        "title": "受付予約",
        "appointments": [
          {
            "id": "uuid",
            "time": "10:25",
            "owner_name": "山田 太郎",
            "pet_type": "犬",
            "pet_name": "ポチ",
            "pet_id": "uuid",
            "visit_type": "再診",
            "service_type": "診療",
            "is_designated": true,
            "doctor": "医師A"
          }
        ]
      },
      {
        "id": "checked_in",
        "title": "受付済",
        "appointments": [ ... ]
      },
      {
        "id": "in_consultation",
        "title": "診療中",
        "appointments": [ ... ]
      },
      {
        "id": "accounting_waiting",
        "title": "会計待ち",
        "appointments": [ ... ]
      },
      {
        "id": "completed",
        "title": "会計済",
        "appointments": [ ... ]
      }
    ]
  }
}
```

#### PUT /dashboard/kanban/:appointmentId

予約のステータスを更新します（カラム間移動）。

**リクエストボディ:**

```json
{
  "column_id": "in_consultation"
}
```

---

## バリデーションルール

### 飼い主

| フィールド | ルール |
|-----------|--------|
| name | 必須、100文字以内 |
| phone | 電話番号形式 |
| email | メールアドレス形式 |

### ペット

| フィールド | ルール |
|-----------|--------|
| owner_id | 必須、有効なUUID |
| name | 必須、100文字以内 |
| species | 必須、50文字以内 |
| birth_date | 日付形式（過去の日付） |
| weight | 0より大きい数値 |

### カルテ

| フィールド | ルール |
|-----------|--------|
| pet_id | 必須、有効なUUID |
| visit_type | 必須、初診/再診 |
| doctor_id | 必須、有効なUUID |

### 予約

| フィールド | ルール |
|-----------|--------|
| pet_id | 必須、有効なUUID |
| start_time | 必須、未来の日時 |
| end_time | 必須、start_timeより後 |
| service_type | 必須、有効なサービス種別 |

---

## Swagger/OpenAPI

Swagger UIは以下のURLで利用可能です：

```
http://localhost:8080/swagger/index.html
```

OpenAPI仕様書（JSON）:

```
http://localhost:8080/swagger/doc.json
```
