# Animal Ekarte API 使用例

## 概要

このドキュメントではAnimal Ekarte APIの主要なエンドポイントの使用例を示します。

## 認証

APIキー認証を使用します。リクエストヘッダーに以下を含めてください：

```
Authorization: Bearer YOUR_API_KEY
```

## 基本URL

```
http://localhost:8080/api/v1
```

## エンドポイント例

### 1. ヘルスチェック

```bash
curl -X GET "http://localhost:8080/health" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**レスポンス:**
```json
{
  "database": "connected",
  "message": "Animal Ekarte API is running",
  "status": "ok",
  "timestamp": "2026-01-25T20:48:57.347113513Z",
  "version": "1.0.0"
}
```

### 2. ペット管理

#### ペット一覧取得

```bash
curl -X GET "http://localhost:8080/api/v1/pets?page=1&limit=10&search=ポチ" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

#### ペット作成

```bash
curl -X POST "http://localhost:8080/api/v1/pets" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ポチ",
    "species": "犬",
    "breed": "柴犬",
    "gender": "男",
    "birth_date": "2020-01-15",
    "weight": 12.5,
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "pet_number": "P001"
  }'
```

#### ペット情報更新

```bash
curl -X PUT "http://localhost:8080/api/v1/pets/{pet_id}" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ポチ（更新）",
    "weight": 13.2
  }'
```

#### ペット削除

```bash
curl -X DELETE "http://localhost:8080/api/v1/pets/{pet_id}" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 3. 飼い主管理

#### 飼い主一覧取得

```bash
curl -X GET "http://localhost:8080/api/v1/owners?page=1&limit=10&search=田中" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

#### 飼い主作成

```bash
curl -X POST "http://localhost:8080/api/v1/owners" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "田中太郎",
    "name_kana": "タナカタロウ",
    "phone": "090-1234-5678",
    "email": "tanaka@example.com",
    "address": "東京都渋谷区1-2-3",
    "notes": "初回診察"
  }'
```

### 4. 電子カルテ管理

#### カルテ一覧取得（フィルタリング付き）

```bash
curl -X GET "http://localhost:8080/api/v1/medical-records/paginated?page=1&limit=10&pet_id={pet_id}&visit_type=初診&date_from=2026-01-01&date_to=2026-01-31" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

#### カルテ作成

```bash
curl -X POST "http://localhost:8080/api/v1/medical-records" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "pet_id": "87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd",
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "visit_date": "2026-01-25T10:00:00Z",
    "visit_type": "初診",
    "chief_complaint": "食欲不振",
    "subjective": "飼い主によると昨日から食欲がない",
    "objective": "体温正常、心拍正常、腹部に圧痛なし",
    "assessment": "軽度の消化器系問題",
    "plan": "経過観察、食事指導、再来診を1週間後に設定",
    "surgery_notes": "",
    "diagnosis": "軽度胃炎",
    "treatment": "消化剤処方、食事改善指導",
    "prescription": "ガスター錠 1錠 1日2回 7日分",
    "notes": "飼い主に経過観察の重要性を説明",
    "status": "確定済"
  }'
```

#### カルテ更新

```bash
curl -X PUT "http://localhost:8080/api/v1/medical-records/{record_id}" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "確定済",
    "treatment": "消化剤処方＋食事改善指導",
    "notes": "飼い主に経過観察の重要性を説明。再来診予約済み"
  }'
```

## エラーレスポンス

### バリデーションエラー (400)

```json
{
  "error": "invalid request body"
}
```

### 見つからない (404)

```json
{
  "error": "pet with id 550e8400-e29b-41d4-a716-446655440000 not found: resource not found"
}
```

### サーバーエラー (500)

```json
{
  "error": "internal server error"
}
```

## ページング情報

ページング対応のエンドポイントでは以下の情報が返されます：

```json
{
  "records": [...],
  "current_page": 1,
  "per_page": 10,
  "total": 25,
  "total_pages": 3,
  "has_next": true,
  "has_prev": false
}
```

## 日付形式

- **診察日**: ISO 8601形式 (`2026-01-25T10:00:00Z`)
- **生年月日**: YYYY-MM-DD形式 (`2020-01-15`)

## ステータス値

### 診察タイプ
- `初診`
- `再診`

### カルテステータス
- `作成中`
- `確定済`

### ペットステータス
- `生存`
- `死亡`

## 検索機能

### ペット検索
- 名前、種別、品種で検索可能

### 飼い主検索
- 名前、メール、電話番号で検索可能

### カルテ検索
- ペットID、飼い主ID、診察タイプ、ステータス、診察日範囲でフィルタリング可能

## 制限事項

- 1ページあたりの最大件数: 100件
- テキストフィールドの最大文字数は各モデルで定義
- APIキーは必須
- UUID形式のIDを使用
