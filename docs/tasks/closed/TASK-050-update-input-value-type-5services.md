# TASK-050: UpdateInput 値型パラメータ — 5 サービス

## 優先度

LOW

---

## 概要

参照実装 `medicine_service.go` は `Update` メソッドのシグネチャで `*UpdateMedicineInput`（ポインタ型）を使用している。以下の 5 サービスは値型（`UpdateXxxInput`）で受け取っており、インターフェース定義と実装が medicine 規約と不統一になっている。

---

## 対象

| ファイル | 概算行 | 現状シグネチャ |
|---------|------|------|
| `vaccine_service.go` | L31 | `Update(ctx, clinicID, id uint64, input UpdateVaccineInput)` |
| `insurance_service.go` | L29 | `Update(ctx, clinicID, id uint64, input UpdateInsuranceInput)` |
| `exam_type_service.go` | L29 | `Update(ctx, clinicID, id uint64, input UpdateExamTypeInput)` |
| `cage_service.go` | L30 | `Update(ctx, clinicID, id uint64, input UpdateCageInput)` |
| `hospitalization_plan_service.go` | L32 | `Update(ctx, clinicID, id uint64, input UpdateHospitalizationPlanInput)` |

---

## 問題

```go
// ❌ 値型（5サービス共通パターン）
type VaccineService interface {
    Update(ctx context.Context, clinicID, id uint64, input UpdateVaccineInput) (*model.Vaccine, error)
}

// ✅ medicine（参照実装）
type MedicineService interface {
    Update(ctx context.Context, clinicID, id uint64, input *UpdateMedicineInput) (*model.Medicine, error)
}
```

値型の場合、呼び出し側が大きな struct を毎回コピーするコスト（Go の値セマンティクス）があり、将来的に Input が大きくなるほど影響が増す。また medicine との統一性が損なわれる。

---

## 修正方針

Interface 定義と実装（func レシーバ）・ハンドラ呼び出し側の 3 箇所を同時に変更する。

```go
// 1. Interface
Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error)

// 2. Service 実装
func (s *vaccineService) Update(ctx context.Context, clinicID, id uint64, input *UpdateVaccineInput) (*model.Vaccine, error) {
    fields := buildVaccineUpdateFields(input)
    // ...
}

// 3. Handler 呼び出し
result, err := h.service.Update(ctx, clinicID, id, &svcInput)
```

---

## 備考

- 動作上の影響は軽微（nil チェックが必要になる程度）だが、将来の混乱を防ぐため medicine 規約に統一する。
- TASK-049 で procedure/cage の Input DTO を変更する際に同時対応すると効率的。
- 低優先度のため、他の HIGH/MEDIUM タスクを優先してよい。
