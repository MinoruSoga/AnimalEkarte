# TASK-025: マスタ系 Create ハンドラの *model.Xxx 直接構築 — 追加 8 ドメイン

## 概要

TASK-003 で ExamType / CheckupType / ChiefComplaintType / TrimmingCourse / TrimmingOption の5ドメインを起票済みだが、同一の規約違反がさらに8ドメインで確認された。Medicine と AnimalSpecies は正しく Input DTO を使用しており、その他9ドメインが非対称になっている。

## 優先度

HIGH（handler/service 責務境界が全マスタ系で崩れている）

## 追加対象ドメイン（TASK-003 との合算で合計 13 ドメイン）

| ドメイン | ファイル | 行 |
|---------|---------|-----|
| Vaccine | `handler/vaccine_handler.go` | L65-84 |
| Cage | `handler/cage_handler.go` | L64-79 |
| Insurance | `handler/insurance_handler.go` | L61-74 |
| Procedure | `handler/procedure_handler.go` | L69-90 |
| Consultation | `handler/consultation_handler.go` | L69-88 |
| InquiryTemplate | `handler/inquiry_template_handler.go` | L61-68 |
| Occupation | `handler/occupation_handler.go`（該当箇所） |
| TrimmingCourse | `handler/trimming_master_handler.go` | L61-79（TASK-003 記載済み） |
| TrimmingOption | `handler/trimming_master_handler.go` | L201-216（TASK-003 記載済み） |

## 規約違反

`.claude/CLAUDE.md`:
> handler は Request → service Input DTO に変換するだけで、モデルオブジェクトを組み立ててはならない。`ClinicID` のセット責任がハンドラに漏れている。

## 参照実装（正しいパターン）

`backend/internal/service/medicine_service.go` と `backend/internal/service/animal_species_service.go` が正しい Input DTO パターンを実装している。これを全ドメインに展開すること。

## 修正方針（Vaccine を例に）

```go
// service/vaccine_service.go — Input DTO を追加
type CreateVaccineInput struct {
    Name        string
    Price       *int64
    IsActive    bool
    Description string
    Species     *model.VaccineSpecies
    Interval    *string
    ParentID    *uint64
    SortOrder   int
}

// Service.Create のシグネチャ変更
func (s *vaccineService) Create(ctx context.Context, clinicID uint64, input CreateVaccineInput) (*model.Vaccine, error) {
    v := &model.Vaccine{
        ClinicID:    clinicID,
        Name:        input.Name,
        // ...
    }
    // ...
}

// handler/vaccine_handler.go — model 組み立てを削除
h.svc.Vaccine.Create(c.Request.Context(), clinicID, service.CreateVaccineInput{
    Name:     req.Name,
    Price:    req.Price,
    IsActive: req.IsActive,
    // ...
})
```

## 作業順序

1. service 層に `CreateXxxInput` struct を追加
2. `Create` メソッドシグネチャを `*model.Xxx` → `CreateXxxInput` に変更
3. model 組み立てを service 層に移動
4. handler 側をリクエスト→Input DTO 変換のみにリファクタ
5. `make lint-front` + `go vet ./...` で確認
