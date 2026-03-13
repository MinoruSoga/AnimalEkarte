# [BUG] owner_repository: Pets.AnimalSpecies / Pets.Insurance が Preload されていない

## 優先度: High

## 発見経路
コードレビュー（vercel-react-best-practices セッション）

---

## 問題

`owner_repository.go` の `FindAll` と `FindByID` で `Pets` の関連テーブルが一部 Preload されていなかった。

### FindAll（修正前）
```go
q.Preload("Pets").Preload("Pets.AnimalSpecies").Find(&owners)
// Pets.Insurance が抜けていた
```

### FindByID（修正前）
```go
r.db.Preload("Pets").First(&owner, ...)
// Pets.AnimalSpecies と Pets.Insurance が両方抜けていた
```

---

## 影響

| API | 症状 |
|-----|------|
| `GET /v1/owners` (一覧) | `insurance_name`, `insurance_details` が常に null |
| `GET /v1/owners/:id` (編集画面) | `species`（種別名）が空文字、`insurance_name` も null |

フロントエンドの `transformPet`:
```typescript
species: pet.animal_species?.name ?? "",  // → "" になる
insuranceName: pet.insurance?.name,       // → undefined になる
```

---

## 修正内容

`FindAll` と `FindByID` の両方に以下を追加:

```go
Preload("Pets.AnimalSpecies").Preload("Pets.Insurance")
```

### FindAll（修正後）
```go
q.Preload("Pets").
  Preload("Pets.AnimalSpecies").
  Preload("Pets.Insurance").
  Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").
  Find(&owners)
```

### FindByID（修正後）
```go
r.db.WithContext(ctx).
  Preload("Pets").
  Preload("Pets.AnimalSpecies").
  Preload("Pets.Insurance").
  First(&owner, "id = ? AND clinic_id = ?", id, clinicID)
```

---

## ステータス: ✅ 修正済み

コミット: 対応中セッションにて修正
