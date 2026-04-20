# CODE-QUALITY-223: Route URL パス不一致 2 件

## 概要

ルート登録・Location ヘッダの URL パスに不整合がある。
クライアントが Location ヘッダを参照してリソースにアクセスしようとすると 404 になる。

---

## 問題1: exam_type_handler の Location ヘッダ URL 不一致

**ファイル:** `backend/internal/handler/exam_type_handler.go:75`

### 現状

```go
// Create 成功時の Location ヘッダ
c.Header("Location", fmt.Sprintf("/v1/masters/exam-types/%d", examType.ID))
```

### 実際のルート登録（`staff_handler.go:521-526`）

```go
masters.GET("/examination-types", h.ListExaminationTypes)
masters.POST("/examination-types", ...)
masters.PATCH("/examination-types/reorder", ...)
masters.GET("/examination-types/:id", ...)
masters.PATCH("/examination-types/:id", ...)
masters.DELETE("/examination-types/:id", ...)
```

### 問題

Location ヘッダが `/v1/masters/exam-types/` を指しているが、
実際のエンドポイントは `/api/v1/masters/examination-types/`。
パスが違う（`exam-types` vs `examination-types`）ため、
Location ヘッダを辿ってアクセスすると 404 になる。

### 修正案

```go
c.Header("Location", fmt.Sprintf("/v1/masters/examination-types/%d", examType.ID))
```

---

## 問題2: payment_method_master が `/masters/` グループ外に登録されている

**ファイル:** `backend/internal/handler/handler.go:101` および
`backend/internal/handler/payment_method_master_handler.go:138-146`

### 現状

```go
// payment_method_master だけ masters グループ外に独立登録
RegisterPaymentMethodMasterRoutes(protected)

// 登録されているパス
/api/v1/payment-methods        ← masters グループ外
/api/v1/payment-methods/:id
/api/v1/payment-methods/reorder
```

### 他マスタの登録パス（参照）

```
/api/v1/masters/animal-species
/api/v1/masters/cages
/api/v1/masters/medicines
/api/v1/masters/trimming-courses
// ... 全マスタは /masters/ 以下
```

### 問題

- payment_method_master だけ `/api/v1/masters/` プレフィックスがない
- フロントエンドの認証ミドルウェア・権限チェックが masters グループに依存している場合、
  payment_method_master だけ適用されないリスクがある
- API パスの一貫性が崩れ、クライアント実装・ドキュメントが混乱する

### 修正案

`RegisterMasterRoutes()` 内に統合し、パスを `/api/v1/masters/payment-methods` に変更:

```go
// staff_handler.go の RegisterMasterRoutes() 内に追加
masters.GET("/payment-methods", h.ListPaymentMethodMasters)
masters.POST("/payment-methods", perm(...), h.CreatePaymentMethodMaster)
masters.PATCH("/payment-methods/reorder", perm(...), h.ReorderPaymentMethodMasters)
masters.GET("/payment-methods/:id", h.GetPaymentMethodMaster)
masters.PATCH("/payment-methods/:id", perm(...), h.UpdatePaymentMethodMaster)
masters.DELETE("/payment-methods/:id", perm(...), h.DeletePaymentMethodMaster)
```

**注意:** パス変更はフロントエンドの API 呼び出し箇所も合わせて変更が必要。
破壊的変更のため、移行期間を設けるか旧パスの redirect を検討すること。

## 優先度

- 問題1: HIGH — Location ヘッダが指す URL が 404 になる（機能バグ）
- 問題2: MEDIUM — 機能的には動作するが、URL設計の一貫性と権限適用の整合性リスク
