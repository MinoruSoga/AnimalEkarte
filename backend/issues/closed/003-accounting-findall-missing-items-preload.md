---
status: open
---

# [accounting] GET /v1/accounting の FindAll に BillingItems の Preload 欠落

## 背景

会計一覧ページでは請求明細（BillingItems）を参照して合計金額や内訳を表示するが、
`accounting_repository.FindAll` に `Items` の Preload がないため
一覧では明細が取得できない。詳細ページでのみ正しく表示される。

## 問題

```go
// accounting_repository.go FindAll: Items なし
.Preload("Owner").Preload("Pet").Preload("Payments")

// FindByID: Items あり
.Preload("Items").Preload("Payments").Preload("Owner").Preload("Pet")
```

一覧で合計金額・明細件数を表示したい場合、毎行 `GET /v1/accounting/:id` を
叩かなければならない（N+1）。

## 修正方針

`accounting_repository.FindAll` に `Items` Preload を追加:

```go
.Preload("Items").
Preload("Owner").Preload("Pet").Preload("Payments")
```

ただし Items が大量になる場合はパフォーマンスに注意。
一覧では件数のみ表示するなら `Preload` ではなくサブクエリ COUNT で代替する選択肢もある。

## 完了条件

- [ ] `accounting_repository.FindAll` に `Preload("Items")` 追加（またはCOUNTで代替）
- [ ] `GET /v1/accounting` レスポンスに `items` 配列が含まれる
- [ ] 会計一覧で明細件数・合計金額が表示できる
- [ ] `docker compose exec backend go test ./... -v` がパス
