# BE-052: カルテ一覧API - 種別・主訴・会計IDが返されない

**Status**: Closed
**Closed At**: 2026-03-23

## クローズ情報

- **変更ファイル**:
  - `backend/internal/handler/pet_response.go` — `petSummaryResponse` に `AnimalSpecies` 追加、`toPetSummary` 更新
  - `backend/internal/handler/medical_record_response.go` — `Inquiry` フィールド追加、`toMedicalRecordResponse` に inquiry マッピング追加

---

## 概要

カルテ一覧 `GET /v1/medical-records` のレスポンスで以下のフィールドが欠落している:

1. `pet.animal_species` (動物種名) - `pet` は `{id, name}` のみ返却
2. `inquiry` (問診データ / chief_complaint) - `undefined` で返却
3. `accounting_id` - `undefined` で返却

## 影響

フロントエンドの `transformMedicalRecord()` が以下のフィールドをマッピングしているが、
API が返さないため一覧表示で常に空欄になる:

```typescript
species: record.pet?.animal_species?.name ?? "",   // 常に "" になる
chiefComplaint: record.inquiry?.chief_complaint ?? "",  // 常に "" になる
```

カルテ一覧画面の「種」「主訴」カラムが全件空欄表示となっている。

## 原因（推定）

`medical_records` の一覧取得ハンドラ/リポジトリで `Preload` が不足:

```go
// 必要な Preload
db.Preload("Pet.AnimalSpecies").   // 種別名取得に必要
   Preload("Inquiry").             // 主訴取得に必要
   Preload("Doctor").              // 担当医名（現在は返っている）
   ...
```

## 修正方針

`backend/internal/repository/medical_record_repository.go` の一覧取得クエリに以下を追加:

```go
query = query.Preload("Pet", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "name", "animal_species_id").
        Preload("AnimalSpecies", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name")
        })
}).Preload("Inquiry", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "medical_record_id", "chief_complaint")
})
```

また `accounting_id` が返っていない場合は、会計レコードとの JOIN/サブクエリも要確認。

## 確認方法

```
GET /v1/medical-records?limit=5
→ response[0].pet.animal_species.name が "犬" 等の文字列であること
→ response[0].inquiry.chief_complaint が文字列（空文字含む）であること
```

## 優先度

中（一覧表示上の欠損。詳細画面は別途確認要）
