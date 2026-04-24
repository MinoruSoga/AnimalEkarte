# TASK-250: service — Delete/Update に FindByID 事前確認欠落（新規ドメイン群）

## 優先度
High

## 対象ファイル
- `backend/internal/service/trimming_course_service.go`
- `backend/internal/service/trimming_option_service.go`
- `backend/internal/service/shift_template_service.go`
- `backend/internal/service/reservation_type_group_service.go`
- `backend/internal/service/reservation_type_liff_service.go`

## 問題概要
TASK-243（7ドメイン）に続き、新たに発見された5ドメインでも
`Delete`（および `Update`）メソッドが FindByID 事前確認なしで
FK 依存チェックや UpdateFields を実行している。

規約: **Delete は `FindByID` → FK 依存チェック → `repo.Delete` の順で実行する。**

## 各ファイルの具体的問題

### trimming_course_service.go（行153〜166付近）
```go
func (s *trimmingCourseService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountUsageByCourseID(ctx, clinicID, id)  // ❌ FindByID なし
    // ...
}
```

### trimming_option_service.go（行150〜163付近）
```go
func (s *trimmingOptionService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountUsageByOptionID(ctx, clinicID, id)  // ❌ FindByID なし
    // ...
}
```

### shift_template_service.go — Update（行135〜168付近）
```go
func (s *shiftTemplateService) Update(...) (*model.ShiftTemplate, error) {
    fields := buildShiftTemplateUpdateFields(input)
    if len(fields) == 0 && input.Breaks == nil {
        existing, err := s.repo.FindByID(...)  // no-op 分岐のみ FindByID
        // ...
    }
    result, err = s.repo.UpdateFields(ctx, clinicID, id, fields)  // ❌ 通常パスは FindByID なし
```

### shift_template_service.go — Delete（行170〜176付近）
```go
func (s *shiftTemplateService) Delete(ctx context.Context, clinicID, id uint64) error {
    if err := s.repo.Delete(ctx, clinicID, id); err != nil {  // ❌ FindByID なし
        return apperrors.Wrap(err, "failed to delete shift template")
    }
}
```

### reservation_type_group_service.go（行137〜152付近）
```go
func (s *reservationTypeGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
    count, err := s.repo.CountReservationTypesByGroupID(ctx, clinicID, id)  // ❌ FindByID なし
    // ...
}
```

### reservation_type_liff_service.go — Delete（行125〜140付近）
```go
func (s *reservationTypeLiffService) Delete(ctx context.Context, clinicID, id uint64) error {
    exists, err := s.resRepo.ExistsByReservationTypeID(ctx, clinicID, id)  // ❌ FindByID なし
    // ...
}
```

## 正しい参照実装（chief_complaint_service.go）
```go
func (s *chiefComplaintTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {  // ✅ FindByID 先行
        return apperrors.Wrap(err, "failed to get chief complaint type")
    }
    count, err := s.repo.CountUsageByChiefComplaintTypeID(ctx, clinicID, id)
    // ...
}
```

## 完了条件
- [ ] 上記5ファイルの Delete（および shift_template の Update）先頭に `FindByID` を追加
- [ ] FindByID エラー時は `apperrors.Wrap(err, "failed to get {entity}")` を返す
- [ ] `go test ./backend/internal/...` がパス
