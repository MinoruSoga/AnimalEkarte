# BE-047: medical_record_service に owner_id 変更バリデーション追加

**Status**: Open
**Priority**: High
**Affects**: medical-records API（PATCH /api/v1/medical-records/:id）
**Date Created**: 2026-03-19
**Related**: TASK-022, FE-077

## Summary

`medical_record_service.Update()` で `owner_id` / `pet_id` 変更時に、指定されたリソースが同一クリニックに所属しているかバリデーションする。現在はバリデーションなしで `repo.Update()` に直接渡しており、外部キー制約エラーでしか検知できない。

## 現状のコード

```go
// backend/internal/service/medical_record_service.go:43-44
func (s *medicalRecordService) Update(ctx context.Context, record *model.MedicalRecord) error {
	return s.repo.Update(ctx, record)
}
```

バリデーションが一切ない。`pet_service.go:226-231` にある owner_id バリデーションが参照実装：

```go
// backend/internal/service/pet_service.go:226-231（参照実装）
if input.OwnerID != nil {
	if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
		return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
	}
}
```

## 必要な変更

### 1. Service インターフェース・構造体に OwnerRepository / PetRepository 追加

```go
// backend/internal/service/medical_record_service.go

type medicalRecordService struct {
	repo      repository.MedicalRecordRepository
	ownerRepo repository.OwnerRepository  // 追加
	petRepo   repository.PetRepository    // 追加
}

func NewMedicalRecordService(
	repo repository.MedicalRecordRepository,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
) MedicalRecordService {
	return &medicalRecordService{
		repo:      repo,
		ownerRepo: ownerRepo,
		petRepo:   petRepo,
	}
}
```

### 2. Update メソッドにバリデーション追加

```go
// backend/internal/service/medical_record_service.go
func (s *medicalRecordService) Update(ctx context.Context, record *model.MedicalRecord) error {
	// owner_id 変更時: クリニック所属確認
	if record.OwnerID != nil {
		if _, err := s.ownerRepo.FindByID(ctx, record.ClinicID, *record.OwnerID); err != nil {
			return apperrors.WrapInvalidInput("owner not found in this clinic")
		}
	}

	// pet_id 変更時: クリニック所属確認
	if record.PetID != nil {
		if _, err := s.petRepo.FindByID(ctx, record.ClinicID, *record.PetID); err != nil {
			return apperrors.WrapInvalidInput("pet not found in this clinic")
		}
	}

	return s.repo.Update(ctx, record)
}
```

### 3. DI 配線更新

```go
// backend/cmd/api/main.go（DI 配線箇所）
// Before:
medicalRecordService := service.NewMedicalRecordService(medicalRecordRepo)

// After:
medicalRecordService := service.NewMedicalRecordService(medicalRecordRepo, ownerRepo, petRepo)
```

## API レスポンス形式（エラー時）

```json
{
  "error": "owner not found in this clinic"
}
```

HTTP Status: 400 Bad Request

## フロントエンド影響

- FE-077 で飼主変更時に 400 エラーを適切にハンドリングする必要がある
- `handleApiError()` で既にカバーされている

## 完了条件

- [ ] `medicalRecordService` に `ownerRepo` / `petRepo` を DI
- [ ] `Update()` で owner_id 変更時に clinic 所属確認
- [ ] `Update()` で pet_id 変更時に clinic 所属確認
- [ ] `cmd/api/main.go` の DI 配線更新
- [ ] 既存テスト・ビルドが通る
- [ ] 存在しない owner_id を送信した場合に 400 が返る
