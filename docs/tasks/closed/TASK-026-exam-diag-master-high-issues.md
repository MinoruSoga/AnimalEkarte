# TASK-026: 検査・診断系マスタ HIGH 問題 3件

## 優先度

HIGH

---

## 問題 1: checkup_type ハンドラが Response DTO 変換をバイパス

### ファイル
`backend/internal/handler/checkup_type_handler.go:31, 45, 77, 113`

### 問題
`toCheckupTypeResponse()` が `checkup_type_response.go` に定義されているにもかかわらず、GetByID / List / Create / Update の全レスポンスで `*model.CheckupType` / `[]model.CheckupType` を直接 `c.JSON` に渡している。exam_type / diagnosis / chief_complaint の3ドメインは全て Response DTO 経由で返しており、checkup_type のみ非対称。GORM モデルが JSON に直接露出するため、DB カラム追加ごとに API コントラクトが変化する。

### 修正案
```go
// GetCheckupType
c.JSON(http.StatusOK, toCheckupTypeResponse(checkupType))

// ListCheckupTypes
c.JSON(http.StatusOK, mapSlice(checkupTypes, toCheckupTypeResponse))

// CreateCheckupType
c.JSON(http.StatusCreated, toCheckupTypeResponse(checkupType))

// UpdateCheckupType
c.JSON(http.StatusOK, toCheckupTypeResponse(checkupType))
```

---

## 問題 2: chief_complaint_type に Reorder が未実装（sort_order カラムは存在）

### ファイル
- `backend/internal/service/chief_complaint_type_service.go`
- `backend/internal/repository/chief_complaint_type_repository.go`
- `backend/internal/handler/chief_complaint_handler.go`

### 問題
`sort_order` カラムが存在するにもかかわらず、exam_type / checkup_type / diagnosis_type の3ドメインが Reorder を実装済みなのに chief_complaint_type だけ欠落している。フロントエンドから並び順を変更できない。

### 修正案
```go
// repository interface に追加
Reorder(ctx context.Context, clinicID uint64, ids []uint64) error

// repository 実装（reorderByClinicID ヘルパーを使用）
func (r *chiefComplaintTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    return reorderByClinicID(ctx, r.db, &model.ChiefComplaintType{}, "chief_complaint_type", clinicID, ids)
}

// service interface に追加
Reorder(ctx context.Context, clinicID uint64, ids []uint64) error

// service 実装
func (s *chiefComplaintTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if len(ids) == 0 {
        return apperrors.WrapInvalidInput("ids must not be empty")
    }
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder chief complaint categories")
    }
    slog.InfoContext(ctx, "chief_complaint_types reordered",
        slog.Uint64("clinic_id", clinicID),
        slog.Int("count", len(ids)))
    return nil
}

// handler に追加 + routes 登録
```

---

## 問題 3: exam_type / checkup_type の Delete で子ノード存在チェックなし

### ファイル
- `backend/internal/service/exam_type_service.go:69-81`
- `backend/internal/service/checkup_type_service.go`（相当箇所）
- `backend/internal/repository/exam_type_repository.go`

### 問題
`exam_type` と `checkup_type` は `parent_id` による木構造を持つ。親を削除すると孤立した子レコードが残る。`diagnosis_type` の Delete では `CountNamesByCategoryID` で子の存在チェックを実装済みであり、exam_type / checkup_type だけが欠落している。

### 修正案
```go
// repository インターフェースに追加
CountChildrenByParentID(ctx context.Context, parentID uint64) (int64, error)

// repository 実装
func (r *examTypeRepository) CountChildrenByParentID(ctx context.Context, parentID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.ExaminationType{}).
        Where("parent_id = ?", parentID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "examination_type", "")
    }
    return count, nil
}

// service Delete の先頭に追加
childCount, err := s.repo.CountChildrenByParentID(ctx, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check exam type children")
}
if childCount > 0 {
    return apperrors.WrapConflict("この検査種別にはサブ種別が登録されているため削除できません")
}
// 既存の CountUsageByExamTypeID チェックに続く
```

checkup_type も同様のパターンで実装すること。
