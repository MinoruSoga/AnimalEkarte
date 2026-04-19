# TASK-031: 臨床系ハンドラの責務分離違反 2件（handler→repository 直接参照 / model 直接構築）

## 優先度

HIGH

---

## 問題 1: treatment_handler / treatment_plan_handler が repository を直接呼んでいる

### ファイル
- `backend/internal/handler/treatment_handler.go:117`
- `backend/internal/handler/treatment_plan_handler.go:174`

### 問題
割引権限チェック（BUG-372 対応）のために handler が `h.repos.Treatment.FindByID` / `h.repos.TreatmentPlan.FindByID` を直接呼んでいる。

```go
// treatment_handler.go:117（規約違反）
existing, err := h.repos.Treatment.FindByID(c.Request.Context(), clinicID, treatmentID)

// treatment_plan_handler.go:174（規約違反）
existing, err := h.repos.TreatmentPlan.FindByID(c.Request.Context(), clinicID, planID)
```

### 修正案
```go
// service interface に追加
type TreatmentService interface {
    // 既存メソッド...
    GetByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
}

// handler では service 経由で取得
existing, err := h.svc.Treatment.GetByID(c.Request.Context(), clinicID, treatmentID)
```

---

## 問題 2: examination / vaccination / hospitalization の Create が *model.Xxx を直接構築

### ファイル
- `backend/internal/handler/examination_handler.go:102-125`
- `backend/internal/handler/vaccination_handler.go:116-133`
- `backend/internal/handler/hospitalization_handler.go:108-132`

### 問題
handler 内で `&model.Examination{...}` / `&model.Vaccination{...}` / `&model.Hospitalization{...}` を直接組み立てて service に渡している。TASK-003 / TASK-025 と同根の Input DTO 未使用問題。

### 参照実装
`treatment_handler.go` の `CreateTreatmentInput` DTO パターンが正しい実装例。

### 修正案（examination を例に）
```go
// service/examination_service.go — Input DTO 追加
type CreateExaminationInput struct {
    MedicalRecordID uint64
    ExamTypeID      uint64
    Status          model.ExaminationStatus
    Date            *time.Time
    Notes           string
    // ...
}

// service シグネチャ変更
Create(ctx context.Context, clinicID uint64, input CreateExaminationInput) (*model.Examination, error)

// handler 側: model 組み立てを削除して DTO のみ構築
```

vaccination / hospitalization も同様に `CreateVaccinationInput` / `CreateHospitalizationInput` を追加する。
