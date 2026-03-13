---
status: open
---

# [pet] GET /v1/pets に Owner/AnimalSpecies/Insurance の Preload 追加・飼主名検索・レスポンスネスト対応

## 背景

「飼主・ペット一覧」ページはペットを主体に表示し、各行に飼主情報（氏名・電話番号）を
合わせて表示する設計。しかし現状 `GET /v1/pets` は関連データなしで返るため、
フロントエンドが `GET /v1/owners` を呼んでクライアント側でフラット化するという
迂回路を使っている。これによりページネーションが owner 単位になり pet 単位にならない。

## 問題

### 1. `pet_repository.FindAll` に Preload が一切ない

```go
// 現状: 関連データなし
q.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&pets)

// FindByID には Preload あり
r.db.Preload("Owner").Preload("Insurance").First(...)
```

`GET /v1/pets` の結果は `owner_id` だけが返り、飼主名・電話番号は空になる。

### 2. `petResponse` に Owner/AnimalSpecies/Insurance のネスト情報がない

```go
// 現状: ID のみ
type petResponse struct {
    OwnerID         uint64 `json:"owner_id"`
    AnimalSpeciesID uint64 `json:"animal_species_id"`
    InsuranceID     *uint64 `json:"insurance_id,omitempty"`
    // owner_name, animal_species.name, insurance.name がない
}
```

### 3. 検索がペット名のみ（飼主名で検索できない）

```go
// 現状: name, pet_name_kana のみ
q.Where(`(name ILIKE ? OR pet_name_kana ILIKE ?)`, pattern, pattern)
// 飼主名（owner_name）で検索できない
```

## 修正方針

### pet_repository.go

`FindAll` に Preload を追加する:

```go
if err := q.
    Preload("Owner").
    Preload("AnimalSpecies").
    Preload("Insurance").
    Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").
    Find(&pets).Error; err != nil {
```

検索条件に飼主名を追加:

```go
q = q.Joins("LEFT JOIN owners ON owners.id = pets.owner_id AND owners.deleted_at IS NULL").
    Where(
        `(pets.name ILIKE ? ESCAPE '\' OR pets.pet_name_kana ILIKE ? ESCAPE '\' OR owners.owner_name ILIKE ? ESCAPE '\')`,
        pattern, pattern, pattern,
    )
```

### pet_response.go

ネスト型を追加し `toPetResponse` で設定する:

```go
type petOwnerNested struct {
    ID        uint64 `json:"id"`
    OwnerName string `json:"owner_name"`
    Phone     string `json:"phone"`
}

type petAnimalSpeciesNested struct {
    ID   uint64 `json:"id"`
    Name string `json:"name"`
}

type petInsuranceNested struct {
    ID           uint64  `json:"id"`
    Name         string  `json:"name"`
    CoverageRate float64 `json:"coverage_rate"`
}

type petResponse struct {
    // ... 既存フィールド ...
    Owner         *petOwnerNested         `json:"owner,omitempty"`
    AnimalSpecies *petAnimalSpeciesNested  `json:"animal_species,omitempty"`
    Insurance     *petInsuranceNested     `json:"insurance,omitempty"`
}
```

### フロントエンド（別途対応）

`frontend/issues/` に対応チケットを起票する。
`owners/loaders.ts` の迂回路を廃止し `GET /v1/pets` を直接使う。

## 完了条件

- [ ] `pet_repository.FindAll` に `Preload("Owner").Preload("AnimalSpecies").Preload("Insurance")` 追加
- [ ] `pet_repository.FindAll` の search に `owners.owner_name` ILIKE 検索追加
- [ ] `petResponse` に `owner`, `animal_species`, `insurance` ネスト追加
- [ ] `toPetResponse` でネストデータを設定
- [ ] `GET /v1/pets` レスポンスに `owner.owner_name`, `animal_species.name` が含まれる
- [ ] `docker compose exec backend go test ./... -v` がパス
