# BE-055: 入院一覧API で Pet.AnimalSpecies を Preload する

## 種類
バグ修正（データ欠損）

## 関連バグ
OPEN-BUG-015: 入院・ホテル管理リストビューで「種」カラムが全件空欄

## 現状
`hospitalization_repository.go` の `List` メソッドで `Pet` を Preload しているが、
`Pet.AnimalSpecies` は Preload されていないため、API レスポンスに `animal_species.name` が含まれない。

```go
// 現在
q.Preload("Pet").Preload("Owner").Preload("Cage").Preload("Doctor")
```

## 修正内容

```go
// 修正後
q.Preload("Pet.AnimalSpecies").Preload("Owner").Preload("Cage").Preload("Doctor")
```

`"Pet"` を `"Pet.AnimalSpecies"` に変更することで、GORM がネスト Preload を実行する。

## 影響ファイル
- `backend/internal/repository/hospitalization_repository.go`
  - `List` メソッドの Preload 引数を `"Pet"` → `"Pet.AnimalSpecies"` に変更

## 注意
フロントエンドの `transforms.ts` はすでに `hosp.pet?.animal_species?.name` を参照している。
本 BE 修正により自動的に「種」カラムが表示されるようになる。
