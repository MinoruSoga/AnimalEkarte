# BE-069: accounting_service が *model.Billing を素通しし service DTO がない

**Status**: Closed
**Priority**: Medium
**Affects**: `backend/internal/service/accounting_service.go`
**Date Created**: 2026-03-26
**Related**: BE-068

## Summary

`accounting_service.go` の `Create`/`Update` メソッドが `*model.Billing` を引数に取り、バリデーションや変換なしにそのまま repository に渡している。handler が直接 model を組み立てており、service がただのパススルー層になっている。プロジェクト規約では service 層に `CreateXxxInput` / `UpdateXxxInput` DTO を定義し、バリデーションを service 内で行うべきとされている。

## 現状のコード

```go
// backend/internal/service/accounting_service.go（概略）
func (s *accountingService) Create(ctx context.Context, billing *model.Billing) (*model.Billing, error) {
	// バリデーションなし、そのまま渡す
	return s.repo.Create(ctx, billing)
}

func (s *accountingService) Update(ctx context.Context, clinicID, billingID uint64, billing *model.Billing) (*model.Billing, error) {
	// バリデーションなし、そのまま渡す
	return s.repo.Update(ctx, clinicID, billingID, billing)
}
```

## 必要な変更

### 1. CreateAccountingInput DTO を定義

```go
// backend/internal/service/accounting_service.go

type CreateAccountingInput struct {
	ClinicID    uint64
	PetID       uint64
	VisitDate   time.Time
	Notes       string
	Items       []CreateBillingItemInput
}

type CreateBillingItemInput struct {
	Name     string
	Quantity int
	UnitPrice int64
	Category string
}
```

### 2. service 内でバリデーションを実装

```go
func (s *accountingService) Create(ctx context.Context, input CreateAccountingInput) (*model.Billing, error) {
	if input.PetID == 0 {
		return nil, apperrors.WrapInvalidInput("pet_id is required")
	}
	if input.VisitDate.IsZero() {
		return nil, apperrors.WrapInvalidInput("visit_date is required")
	}
	// model 組み立ては service 層で行う
	billing := &model.Billing{
		ClinicID:  input.ClinicID,
		PetID:     input.PetID,
		VisitDate: input.VisitDate,
		Notes:     input.Notes,
		Status:    model.BillingStatusWaiting,
	}
	return s.repo.Create(ctx, billing)
}
```

### 3. handler を DTO 経由に変更

handler は `CreateAccountingInput` を組み立てて service に渡すだけにする（model を直接組み立てない）。

## 依存関係

BE-068 と同時対応推奨（Update 側の DTO は BE-068 で定義する）。

## 完了条件

- [ ] `CreateAccountingInput` / `UpdateAccountingInput` DTO を service 層に定義
- [ ] service 内でバリデーション（必須フィールド・範囲チェック）を実装
- [ ] handler は DTO を組み立てて service に渡すだけにする
- [ ] `docker compose exec backend go test ./... -v` がパス
