# BUG-424: 複数マスタサービスの Update メソッドで存在確認（FindByID）が欠落

## 概要

BUG-420（trimming_service）・BUG-422（削除パターン不統一）に続いて発見された追加インスタンス。
以下のサービスで Update メソッドが `FindByID(ctx, clinicID, id)` による存在確認・テナント検証を
行わずに直接 `UpdateFields` を呼び出している。

## 問題箇所

### 1. `reservation_type_service.go:266-279`

```go
func (s *reservationTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    fields := buildReservationTypeUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
    }
    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← 存在確認なし
    // ...
}
```

### 2. `trimming_master_service.go:82-96`（TrimmingCourse）

```go
func (s *trimmingCourseService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
    // バリデーションのみ
    fields := buildTrimmingCourseUpdateFields(input)
    course, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← 存在確認なし
}
```

### 3. `trimming_master_service.go:238-251`（TrimmingOption）

```go
func (s *trimmingOptionService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
    fields := buildTrimmingOptionUpdateFields(input)
    option, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← 存在確認なし
}
```

### 4. `diagnosis_service.go:261-282`（DiagnosisName）

```go
func (s *diagnosisNameService) Update(ctx context.Context, clinicID, id uint64, input *UpdateDiagnosisNameInput) (*model.DiagnosisName, error) {
    // FK検証はあるが本体の存在確認がない
    if input.DiagnosisTypeID != nil {
        if _, err := s.typeRepo.FindByID(ctx, clinicID, *input.DiagnosisTypeID); err != nil {
            return nil, apperrors.WrapInvalidInput("診断カテゴリが見つかりません")
        }
    }
    fields := buildDiagnosisNameUpdateFields(input)
    name, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← id の存在確認なし
}
```

## 標準パターン（cage_service.go）

```go
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    // ✅ Step 1: 存在確認（テナント検証含む）
    existing, err := s.repo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to get cage")
    }
    // Step 2: フィールド更新 ...
}
```

## 修正方針

各 Update メソッドの先頭に存在確認を追加する。

```go
// reservation_type_service.go:266 に追加
if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
    return nil, apperrors.Wrap(err, "failed to get reservation type")
}

// trimming_master_service.go（TrimmingCourse）
if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
    return nil, apperrors.Wrap(err, "failed to get trimming course")
}

// trimming_master_service.go（TrimmingOption）
if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
    return nil, apperrors.Wrap(err, "failed to get trimming option")
}

// diagnosis_service.go（DiagnosisName）
if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
    return nil, apperrors.Wrap(err, "failed to get diagnosis name")
}
```

## 影響ファイル

- `backend/internal/service/reservation_type_service.go` — 行 266-279
- `backend/internal/service/trimming_master_service.go` — 行 82-96, 238-251
- `backend/internal/service/diagnosis_service.go` — 行 261-282

## 優先度

**High** — マルチテナント境界の保護が不完全。他クリニックのリソースを上書きできる可能性。

## 関連チケット

- BUG-420（trimming_service.Update の同種問題）
- BUG-422（Delete 前の存在確認パターン不統一）
