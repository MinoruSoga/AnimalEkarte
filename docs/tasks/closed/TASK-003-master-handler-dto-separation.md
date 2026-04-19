# TASK-003: マスタ系ハンドラ層でのモデル直接構築を Service Input DTO に移行

## 概要

マスタ系ハンドラ（ExamType / CheckupType / ChiefComplaintType / TrimmingCourse / TrimmingOption）の Create / Update ハンドラが、`*model.Xxx` を直接組み立ててサービスに渡している。`ClinicID` のセットや ENUM 型キャスト (`model.TargetSize` など) がハンドラ層に漏れており、`handler → service（Input DTO）→ repository` のレイヤー規約に違反している。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/handler/exam_type_handler.go` | L61-69（Create でモデル直接構築） |
| `backend/internal/handler/checkup_type_handler.go` | L61-72（Create でモデル直接構築） |
| `backend/internal/handler/chief_complaint_handler.go` | L61-67（Create でモデル直接構築） |
| `backend/internal/handler/trimming_master_handler.go` | L61-79, L201-211（Course/Option でモデル直接構築） |

## 規約違反

`.claude/CLAUDE.md` / `.claude/rules/go-language.md`:
> handler は Request → service Input DTO に変換するだけ。モデルオブジェクトを組み立ててはならない。

## 修正方針

1. 各サービス層に `CreateXxxInput` / `UpdateXxxInput` struct を定義する
2. `ClinicID` のセット・ENUM 変換をサービス層に移動する
3. ハンドラはリクエスト DTO → Input DTO の変換のみ行う

### 例（ExamType）

```go
// service/exam_type_service.go
type CreateExamTypeInput struct {
    Name        string
    Price       int64
    IsActive    bool
    Description string
    ParentID    *uint64
    SortOrder   int
}

func (s *examTypeService) Create(ctx context.Context, clinicID uint64, input CreateExamTypeInput) (*model.ExaminationType, error) {
    exType := &model.ExaminationType{
        ClinicID:    clinicID,
        Name:        input.Name,
        Price:       input.Price,
        IsActive:    input.IsActive,
        Description: input.Description,
        ParentID:    input.ParentID,
        SortOrder:   input.SortOrder,
    }
    return s.repo.Create(ctx, exType)
}

// handler/exam_type_handler.go
_, err = h.svc.ExaminationType.Create(c.Request.Context(), clinicID, service.CreateExamTypeInput{
    Name:     req.Name,
    Price:    req.Price,
    IsActive: req.IsActive,
    // ...
})
```

## 対象エンティティ

- `ExaminationType` (exam_type_handler.go / exam_type_service.go)
- `CheckupType` (checkup_type_handler.go / checkup_type_service.go)
- `ChiefComplaintType` (chief_complaint_handler.go / chief_complaint_service.go)
- `TrimmingCourse` (trimming_master_handler.go / trimming_master_service.go)
- `TrimmingOption` (trimming_master_handler.go / trimming_master_service.go)

## テスト

- 各エンティティの Create / Update をユニットテストで検証
- ハンドラテスト: Input DTO への変換が正しいか
- サービステスト: ClinicID / ENUM のセットが正しいか
