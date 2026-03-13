---
status: open
---

# [medical-record] GET /v1/medical-records/:id の FindByID に Pet.AnimalSpecies の Preload 欠落

## 背景

カルテ詳細ページ（カルテフォーム）でペットの動物種を表示するが、
`medical_record_repository.FindByID` が `Pet` を Preload しているものの
`Pet.AnimalSpecies` まで辿っていないため種名が空になる。

## 問題

```go
// FindAll: Pet.AnimalSpecies あり（正しい）
.Preload("Owner").Preload("Pet.AnimalSpecies").Preload("Doctor").Preload("Inquiry")

// FindByID: Pet のみ（AnimalSpecies が欠落）
.Preload("Treatments").
.Preload("Vitals").
.Preload("Doctor").
.Preload("Owner").
.Preload("Pet")  // ← Pet.AnimalSpecies がない
```

カルテ詳細ページで種（犬・猫等）が表示されない。

## 修正方針

`medical_record_repository.FindByID` の `Pet` Preload を `Pet.AnimalSpecies` に変更:

```go
.Preload("Treatments").
.Preload("Vitals").
.Preload("Doctor").
.Preload("Owner").
.Preload("Pet.AnimalSpecies")  // ← 変更
```

## 完了条件

- [ ] `medical_record_repository.FindByID` の `Preload("Pet")` を `Preload("Pet.AnimalSpecies")` に変更
- [ ] `GET /v1/medical-records/:id` レスポンスの `pet.animal_species.name` が正しく返る
- [ ] カルテ詳細フォームで動物種が表示される
- [ ] `docker compose exec backend go test ./... -v` がパス
