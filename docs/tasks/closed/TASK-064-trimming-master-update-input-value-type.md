# TASK-064: trimming_master_service — UpdateInput が値型（TASK-050 対象漏れ）

## 優先度

LOW

---

## 概要

TASK-050 で `UpdateXxxInput` の値型 → ポインタ型移行が 5 サービス（vaccine / insurance / exam_type / cage / hospitalization_plan）を対象としたが、`trimming_master_service.go` の 2 つの Update Input が対象から漏れていた。

参照実装 `medicine_service.go` では `Update` の引数が `*UpdateMedicineInput`（ポインタ）で統一されている。

---

## 問題箇所

### backend/internal/service/trimming_master_service.go

```go
// ❌ 値型（コピーコスト・統一性の問題）
// Interface
Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)

// 実装
func (s *trimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)

// TrimmingOption も同様
Update(ctx context.Context, clinicID, id uint64, input UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
```

---

## 修正方針

Interface・実装・handler 呼び出し側の 3 箇所を変更する。

```go
// ✅ 修正後
// Interface
Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)

// 実装
func (s *trimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)
```

handler 側（`trimming_master_handler.go`）では `&svcInput` でポインタを渡すよう変更。

---

## 影響ファイル

- `backend/internal/service/trimming_master_service.go`（Interface + 実装 × 2 = 4 箇所）
- `backend/internal/handler/trimming_master_handler.go`（UpdateTrimmingCourse / UpdateTrimmingOption の呼び出し × 2 箇所）

---

## 参照実装

`medicine_service.go` の `UpdateMedicineInput` → `*UpdateMedicineInput` パターン。
