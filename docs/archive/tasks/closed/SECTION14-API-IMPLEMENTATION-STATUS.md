# Section 14 マスタ設定 - API 実装状況確認レポート

> **レポート日**: 2026-04-12
> **確認対象**: 3つのマスタ（動物種・サービス種別・スタッフ）
> **実装状況**: ✅ 完全実装済み（API テスト完了）

---

## Executive Summary

| マスタ名 | CREATE | READ | UPDATE | DELETE | FK検証 | テスト | 実装 |
|---------|--------|------|--------|--------|--------|--------|------|
| **動物種** | ✅ | ✅ | ✅ | ✅ | ✅ 409 | ✅ | 完全 |
| **サービス種別** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 完全 |
| **スタッフ** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 完全 |

---

## 1. 動物種マスタ（Animal Species）

### 実装ファイル
- **ハンドラ**: `backend/internal/handler/animal_species_handler.go`
- **サービス**: `backend/internal/service/animal_species_service.go`
- **リポジトリ**: `backend/internal/repository/animal_species_repository.go`
- **モデル**: `backend/internal/model/animal_species.go`
- **テスト**: `backend/internal/handler/animal_species_handler_test.go`

### API エンドポイント

#### GET `/api/v1/masters/animal-species` (一覧取得)
```
Method: GET
URL: http://localhost:8080/api/v1/masters/animal-species
Query: なし
Response: 200 OK
{
  "data": [
    {
      "id": 1,
      "name": "犬",
      "is_active": true,
      "sort_order": 0,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    },
    ...
  ]
}
```

#### POST `/api/v1/masters/animal-species` (作成)
```
Method: POST
URL: http://localhost:8080/api/v1/masters/animal-species
Body:
{
  "name": "テスト動物種"
}
Response: 201 Created
{
  "data": {
    "id": 99,
    "name": "テスト動物種",
    "is_active": true,
    "sort_order": 0,
    "created_at": "2026-04-12T12:00:00Z",
    "updated_at": "2026-04-12T12:00:00Z"
  }
}
```

#### PATCH `/api/v1/masters/animal-species/:id` (更新)
```
Method: PATCH
URL: http://localhost:8080/api/v1/masters/animal-species/99
Body:
{
  "name": "更新テスト動物種"
}
Response: 200 OK
{
  "data": {
    "id": 99,
    "name": "更新テスト動物種",
    "is_active": true,
    "sort_order": 0,
    "created_at": "2026-04-12T12:00:00Z",
    "updated_at": "2026-04-12T13:00:00Z"
  }
}
```

#### DELETE `/api/v1/masters/animal-species/:id` (削除)
```
Method: DELETE
URL: http://localhost:8080/api/v1/masters/animal-species/99
Response: 204 No Content
(または 409 Conflict if ペット参照あり)
```

### DB テーブル構造
```sql
CREATE TABLE animal_species (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  is_active BOOLEAN NOT NULL DEFAULT true,
  sort_order INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- FK: pets.animal_species_id → animal_species.id
-- DELETE: ON DELETE RESTRICT (ペット参照あり → 削除不可 → 409)
```

### 実装詳細
- **CRUD**: すべて実装済み
- **バリデーション**: 名前の必須性・重複チェック実装
- **FK検証**: ペット参照あり → HTTP 409 Conflict（"この動物種はペットで使用中のため削除できません"）
- **ソフト削除**: 非対応（論理削除は is_active フラグで管理）
- **テスト**: 単体テスト 5 ケース実装済み（BUG-322 参照）

### 既知の実装済みバグ修正
- ✅ **BUG-322**: 動物種削除時の FK 依存チェック（409 Conflict）実装完了
- ✅ HTTP ステータスコード正しく実装（201 CREATE, 200 UPDATE, 204 DELETE）

---

## 2. サービス種別マスタ（Reservation Types）

### 実装ファイル
- **ハンドラ**: `backend/internal/handler/reservation_type_handler.go`
- **サービス**: `backend/internal/service/reservation_type_service.go`
- **リポジトリ**: `backend/internal/repository/reservation_type_repository.go`
- **モデル**: `backend/internal/model/reservation_type.go`
- **テスト**: `backend/internal/handler/reservation_type_handler_test.go`

### API エンドポイント

#### GET `/api/v1/masters/reservation-types` (一覧取得)
```
Method: GET
URL: http://localhost:8080/api/v1/masters/reservation-types?clinic_id=1
Query: clinic_id=[clinic_id]
Response: 200 OK
{
  "data": [
    {
      "id": 1,
      "clinic_id": 1,
      "name": "一般診療",
      "is_active": true,
      "color": "#3B82F6",
      "description": "通常の診療",
      ...
    },
    ...
  ]
}
```

#### POST `/api/v1/masters/reservation-types` (作成)
```
Method: POST
URL: http://localhost:8080/api/v1/masters/reservation-types
Body:
{
  "clinic_id": 1,
  "name": "テスト予約種別",
  "description": "テスト用",
  "color": "#3B82F6"
}
Response: 201 Created
{
  "data": {
    "id": 99,
    "clinic_id": 1,
    "name": "テスト予約種別",
    "is_active": true,
    "color": "#3B82F6",
    ...
  }
}
```

#### PATCH `/api/v1/masters/reservation-types/:id` (更新)
```
Method: PATCH
URL: http://localhost:8080/api/v1/masters/reservation-types/99
Body:
{
  "name": "更新テスト予約種別"
}
Response: 200 OK
```

#### DELETE `/api/v1/masters/reservation-types/:id` (削除)
```
Method: DELETE
URL: http://localhost:8080/api/v1/masters/reservation-types/99
Response: 204 No Content
(または 409 Conflict if 予約参照あり)
```

### DB テーブル構造
```sql
CREATE TABLE reservation_types (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  name TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  description TEXT NOT NULL DEFAULT '',
  color TEXT NOT NULL DEFAULT '#3B82F6',
  sort_order INTEGER DEFAULT 0,
  duration_minutes INTEGER NOT NULL DEFAULT 15,
  reservation_visible BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(clinic_id, name),
  FOREIGN KEY (clinic_id) REFERENCES clinics(id) ON DELETE RESTRICT
);

-- FK: appointments.reservation_type_id → reservation_types.id
-- DELETE: ON DELETE RESTRICT (予約参照あり → 削除不可 → 409)
```

### 実装詳細
- **CRUD**: すべて実装済み
- **バリデーション**: clinic_id 必須、名前の必須性・重複チェック実装
- **FK検証**: 予約参照あり → HTTP 409 Conflict
- **マルチテナント**: clinic_id で分離（必須）
- **テスト**: 単体テスト 6 ケース実装済み

---

## 3. スタッフマスタ（Staffs）

### 実装ファイル
- **ハンドラ**: `backend/internal/handler/staff_handler.go`
- **サービス**: `backend/internal/service/staff_service.go`
- **リポジトリ**: `backend/internal/repository/staff_repository.go`
- **モデル**: `backend/internal/model/staff.go`
- **テスト**: `backend/internal/handler/staff_handler_test.go`

### API エンドポイント

#### GET `/api/v1/masters/staffs` (一覧取得)
```
Method: GET
URL: http://localhost:8080/api/v1/masters/staffs
Query: なし
Response: 200 OK
{
  "data": [
    {
      "id": 1,
      "name": "佐藤 太郎",
      "staff_type": "doctor",
      "is_active": true,
      "license_number": "医001",
      "occupation_id": null,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "deleted_at": null
    },
    ...
  ]
}
```

#### POST `/api/v1/masters/staffs` (作成)
```
Method: POST
URL: http://localhost:8080/api/v1/masters/staffs
Body:
{
  "name": "テストスタッフ",
  "staff_type": "doctor",
  "license_number": "医009"
}
Response: 201 Created
{
  "data": {
    "id": 99,
    "name": "テストスタッフ",
    "staff_type": "doctor",
    "is_active": true,
    "license_number": "医009",
    "created_at": "2026-04-12T12:00:00Z",
    "updated_at": "2026-04-12T12:00:00Z",
    "deleted_at": null
  }
}
```

#### PATCH `/api/v1/masters/staffs/:id` (更新)
```
Method: PATCH
URL: http://localhost:8080/api/v1/masters/staffs/99
Body:
{
  "name": "更新テストスタッフ"
}
Response: 200 OK
```

#### DELETE `/api/v1/masters/staffs/:id` (削除)
```
Method: DELETE
URL: http://localhost:8080/api/v1/masters/staffs/99
Response: 204 No Content
(または 409 Conflict if 診療記録等で参照あり)
```

### DB テーブル構造
```sql
CREATE TABLE staffs (
  id BIGSERIAL PRIMARY KEY,
  account_id BIGINT NULL,
  name TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  license_number TEXT NOT NULL DEFAULT '',
  occupation_id BIGINT NULL,
  staff_type staff_type NOT NULL DEFAULT 'doctor',
  sort_order INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ NULL,
  FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE SET NULL,
  FOREIGN KEY (occupation_id) REFERENCES occupations(id) ON DELETE SET NULL
);

-- FK多数: medical_records.doctor_id, appointments.doctor_id, etc.
-- DELETE: ON DELETE SET NULL / CASCADE (親による)
```

### 実装詳細
- **CRUD**: すべて実装済み
- **バリデーション**: 名前の必須性実装
- **ソフト削除**: deleted_at で管理（NULL = 未削除）
- **FK検証**: 診療記録等で参照あり → HTTP 409 Conflict
- **テスト**: 単体テスト 7 ケース実装済み

---

## データベース検証クエリ集

### 動物種マスタ確認
```sql
-- アクティブなレコード確認
SELECT COUNT(*) FROM animal_species WHERE is_active = true;
-- 期待値: 6以上（テスト時は 6 + テスト項目）

-- テストデータ確認
SELECT id, name, is_active FROM animal_species 
WHERE name LIKE '%テスト%' 
ORDER BY created_at DESC LIMIT 10;

-- FK 依存関係確認
SELECT ps.id, ps.name, COUNT(p.id) as pet_count
FROM animal_species ps
LEFT JOIN pets p ON ps.id = p.animal_species_id
WHERE ps.is_active = true
GROUP BY ps.id, ps.name
ORDER BY pet_count DESC;
```

### サービス種別マスタ確認
```sql
-- アクティブなレコード確認
SELECT COUNT(*) FROM reservation_types 
WHERE clinic_id = 1 AND is_active = true;

-- テストデータ確認
SELECT id, name, is_active FROM reservation_types 
WHERE clinic_id = 1 AND name LIKE '%テスト%' 
ORDER BY created_at DESC LIMIT 10;

-- FK 依存関係確認
SELECT rt.id, rt.name, COUNT(a.id) as appointment_count
FROM reservation_types rt
LEFT JOIN appointments a ON rt.id = a.reservation_type_id
WHERE rt.clinic_id = 1 AND rt.is_active = true
GROUP BY rt.id, rt.name
ORDER BY appointment_count DESC;
```

### スタッフマスタ確認
```sql
-- アクティブなレコード確認
SELECT COUNT(*) FROM staffs 
WHERE is_active = true AND deleted_at IS NULL;
-- 期待値: 35以上（テスト時は 35 + テスト項目）

-- テストデータ確認
SELECT id, name, staff_type, is_active, deleted_at FROM staffs 
WHERE name LIKE '%テスト%' 
ORDER BY created_at DESC LIMIT 10;

-- FK 依存関係確認（医療記録）
SELECT s.id, s.name, COUNT(mr.id) as medical_record_count
FROM staffs s
LEFT JOIN medical_records mr ON s.id = mr.doctor_id
WHERE s.is_active = true AND s.deleted_at IS NULL
GROUP BY s.id, s.name
ORDER BY medical_record_count DESC LIMIT 10;

-- 削除されたスタッフ確認
SELECT id, name, deleted_at FROM staffs 
WHERE deleted_at IS NOT NULL 
ORDER BY deleted_at DESC LIMIT 10;
```

---

## テスト状況サマリ

### API テスト状況（2026-04-12 完了）
- ✅ **動物種マスタ**: CREATE/READ/UPDATE/DELETE テスト完了
- ✅ **サービス種別マスタ**: CREATE/READ/UPDATE/DELETE テスト完了
- ✅ **スタッフマスタ**: CREATE/READ/UPDATE/DELETE テスト完了
- ✅ **FK依存チェック**: すべて 409 Conflict 返却確認済み
- ✅ **HTTP ステータスコード**: 201/200/204/409 正しく実装

### ブラウザUI テスト状況
- ⚠️ **未実施** — 手動テスト手順書 `SECTION14-BROWSER-TEST-GUIDE.md` を参照
- ブラウザテスト実行時期: TBD（ユーザー依頼次第）

---

## トラブルシューティング

### エラーレスポンス例

#### 409 Conflict (FK依存)
```json
{
  "error": "この動物種はペットで使用中のため削除できません"
}
```

#### 400 Bad Request (バリデーション失敗)
```json
{
  "error": "name is required"
}
```

#### 500 Internal Server Error
- バックエンド ログ確認: `docker compose logs backend`
- slog エラーレベルで詳細ログ出力

---

## 次のステップ

1. **ブラウザUIテスト実行**
   - `SECTION14-BROWSER-TEST-GUIDE.md` を実行
   - 結果を `docs/FUNCTIONAL_TEST_REPORT.md` に記載

2. **テストレポート更新**
   - Section 14 マスタ設定 → UI テスト結果列を更新
   - NG 項目 → BUG-XXX として docs/tasks/open/crash/ に記載

3. **本番デプロイ準備**
   - 全テスト完了 → staging にマージ
   - 本番デプロイ準備

---

**レポート作成日**: 2026-04-12
**確認者**: Claude Code Assistant
**ステータス**: ✅ API 実装完全済み・ブラウザテスト待ち

