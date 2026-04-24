---
status: open
---

# [hospitalization] GET /v1/hospitalizations の FindAll に Doctor/CarePlanItems の Preload 欠落

## 背景

入院一覧ページでは担当医・ケアプランを表示するが、
`hospitalization_repository.FindAll` はこれらを Preload していない。
詳細ページ（FindByID）との差が大きく、一覧では表示できる情報が限られている。

## 問題

```go
// FindAll: 最低限のみ
.Preload("Pet").Preload("Owner").Preload("Cage")

// FindByID: 詳細データあり
.Preload("Pet").
.Preload("Owner").
.Preload("Cage").
.Preload("Doctor").           // ❌ FindAll にない
.Preload("CarePlanItems").    // ❌ FindAll にない
.Preload("DailyRecords").     // ❌ FindAll にない
.Preload("TreatmentPlans")    // ❌ FindAll にない
```

一覧で担当医が表示できない。

## 修正方針

一覧表示に必要な最低限の Preload を追加する。
`DailyRecords` や `TreatmentPlans` はデータ量が多い可能性があるため一覧には不要。

```go
// FindAll に追加
.Preload("Pet").Preload("Owner").Preload("Cage").Preload("Doctor")
```

`CarePlanItems` / `DailyRecords` / `TreatmentPlans` は FindByID のみで OK。

## 完了条件

- [ ] `hospitalization_repository.FindAll` に `Preload("Doctor")` 追加
- [ ] `GET /v1/hospitalizations` レスポンスに `doctor` フィールドが含まれる
- [ ] 入院一覧で担当医名が表示される
- [ ] `docker compose exec backend go test ./... -v` がパス
