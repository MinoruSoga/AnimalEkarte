# API ロードマップ（未実装設計）

## 概要

本ドキュメントは**今後実装予定のAPI設計**です。
実装済みAPIは Swagger UI を参照してください。

**Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## 実装状況

| 状態 | API |
|------|-----|
| ✅ 実装済 | Pet CRUD (`/pets`), Health (`/health`) → Swagger参照 |
| 📋 設計のみ | 以下すべて |

---

## 共通仕様

### Base URL
```
/api/v1
```

### 認証（将来実装）
```
Authorization: Bearer <token>
```

### レスポンス形式

**成功:**
```json
{
  "data": { ... },
  "meta": { "total": 100, "page": 1, "per_page": 20 }
}
```

**エラー:**
```json
{
  "error": { "code": "NOT_FOUND", "message": "Resource not found" }
}
```

---

## 未実装エンドポイント

### 飼い主 (Owners)

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/owners` | 飼い主一覧取得 |
| GET | `/owners/:id` | 飼い主詳細取得 |
| POST | `/owners` | 飼い主作成 |
| PUT | `/owners/:id` | 飼い主更新 |
| DELETE | `/owners/:id` | 飼い主削除 |
| GET | `/owners/:id/pets` | 飼い主のペット一覧 |

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

## 詳細設計

各APIの詳細（リクエスト/レスポンス例）は実装時に Swagger アノテーションで定義し、
自動生成されるドキュメントで管理します。
