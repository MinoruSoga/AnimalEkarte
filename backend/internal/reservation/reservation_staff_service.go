package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	colReservationStaffName               = "name"
	colReservationStaffStaffType          = "staff_type"
	colReservationStaffReservationVisible = "reservation_visible"
	colReservationStaffReservationComment = "reservation_comment"
	colReservationStaffSortOrder          = "sort_order"
)

func buildReservationStaffUpdate(input *UpdateReservationStaffInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colReservationStaffName] = *input.Name
	}
	if input.StaffType != nil {
		fields[colReservationStaffStaffType] = *input.StaffType
	}
	if input.ReservationVisible != nil {
		fields[colReservationStaffReservationVisible] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields[colReservationStaffReservationComment] = *input.ReservationComment
	}
	if input.SortOrder != nil {
		fields[colReservationStaffSortOrder] = *input.SortOrder
	}
	return fields
}

// CreateReservationStaffInput は予約スタッフ作成の入力データ
type CreateReservationStaffInput struct {
	Name               string
	StaffType          string
	ReservationVisible bool
	ReservationComment string
	SortOrder          int
}

// UpdateReservationStaffInput は予約スタッフ更新の入力データ（ポインタ型でゼロ値を区別）
type UpdateReservationStaffInput struct {
	Name               *string
	StaffType          *string
	ReservationVisible *bool
	ReservationComment *string
	SortOrder          *int
}
// ReservationStaffCoreService は予約スタッフの CRUD・ステータス・並び順操作
type ReservationStaffCoreService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error)
	PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
}

// ReservationStaffExclusionService は予約スタッフの除外コース操作
type ReservationStaffExclusionService interface {
	// ListExcludedByStaffIDs は複数スタッフの除外コースをバルク取得してスタッフID→除外コース一覧のマップを返す（N+1回避）
	// Stage B: values are derived from capabilities (compatibility facade).
	ListExcludedByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error)
}

// ReservationStaffCapabilityService は対応可能予約区分の bulk read。
type ReservationStaffCapabilityService interface {
	// ListCapableByStaffIDs は複数スタッフの対応可能コースをバルク取得する（N+1回避）
	ListCapableByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) (map[uint64][]model.StaffReservationCapability, error)
}

// ReservationStaffService は ReservationStaffCoreService / ReservationStaffExclusionService を統合したインターフェース。
type ReservationStaffService interface {
	ReservationStaffCoreService
	ReservationStaffExclusionService
	ReservationStaffCapabilityService
}

type reservationStaffService struct {
	repo         ReservationStaffRepository
	transactor   Transactor
	staffDeleter ReservationStaffDeleter
}

func NewReservationStaffService(
	repo ReservationStaffRepository,
	transactor Transactor,
	staffDeleter ReservationStaffDeleter,
) ReservationStaffService {
	return &reservationStaffService{
		repo:         repo,
		transactor:   transactor,
		staffDeleter: staffDeleter,
	}
}

func (s *reservationStaffService) List(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	staffs, err := s.repo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservation staffs", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list reservation staffs")
	}
	return staffs, nil
}

func (s *reservationStaffService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation staff", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation staff")
	}
	return staff, nil
}

func (s *reservationStaffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	// TASK-021 UNIT-021-A: request excluded_type_ids removed. Create always seeds full active
	// universe via inverse facade with empty exclusion (same as former empty-array path).
	staffType := model.StaffType(input.StaffType)
	if staffType == "" {
		staffType = model.StaffTypeDoctor
	}
	staff := &model.Staff{
		ClinicID:           clinicID,
		Name:               input.Name,
		IsActive:           true,
		SortOrder:          input.SortOrder,
		StaffType:          staffType,
		ReservationVisible: input.ReservationVisible,
		ReservationComment: input.ReservationComment,
	}
	var excluded []model.StaffReservationExclusion
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, staff, clinicID); err != nil {
			slog.ErrorContext(txCtx, "failed to create reservation staff", "error", err)
			return apperrors.Wrap(err, "failed to create reservation staff")
		}
		// Stage B: inverse facade with empty excluded ⇒ full active universe capable.
		if err := s.repo.UpdateExcludedReservationTypes(txCtx, clinicID, staff.ID, []uint64{}); err != nil {
			slog.ErrorContext(txCtx, "failed to set excluded courses", "error", err)
			return apperrors.Wrap(err, "failed to set excluded courses")
		}
		var err error
		excluded, err = s.repo.FindAllExcludedReservationTypes(txCtx, clinicID, staff.ID)
		if err != nil {
			slog.ErrorContext(
				txCtx,
				"failed to get excluded reservation types",
				"error", err,
				"clinic_id", clinicID,
			)
			return apperrors.Wrap(err, "failed to get excluded reservation types")
		}
		return nil
	}); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create reservation staff")
	}
	slog.InfoContext(ctx, "reservation staff created",
		slog.Uint64("staff_id", staff.ID),
		slog.Uint64("clinic_id", clinicID))
	return staff, excluded, nil
}

func (s *reservationStaffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	// TASK-021 UNIT-021-A: request excluded_type_ids removed. Capability replacement is only via
	// capable-reservation-types API. Update is staff-fields only + exclusion readback.
	// BE-refactor.md X-8: staff 本体更新を WithTx で括り原子化する。	// mutation-specific UPDATE lock を最初に取得し、所有権確認から更新までを直列化する。
	// FindByID の SHARE lock から UPDATE への lock upgrade は、Update と PatchStatus が
	// 同時実行された際に相互待ちを作り得るため mutation の先頭では使用しない。
	var updated *model.Staff
	var excluded []model.StaffReservationExclusion
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockForMutation(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to verify reservation staff ownership", "error", err)
			return apperrors.Wrap(err, "failed to verify reservation staff ownership")
		}
		fields := buildReservationStaffUpdate(input)
		if len(fields) > 0 {
			if err := s.repo.Update(txCtx, clinicID, id, fields); err != nil {
				slog.ErrorContext(txCtx, "failed to update reservation staff", "error", err, "id", id, "clinic_id", clinicID)
				return apperrors.Wrap(err, "failed to update reservation staff")
			}
		}
		var err error
		updated, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(
				txCtx,
				"failed to get reservation staff after update",
				"error", err,
				"id", id,
				"clinic_id", clinicID,
			)
			return apperrors.Wrap(err, "failed to get reservation staff after update")
		}
		excluded, err = s.repo.FindAllExcludedReservationTypes(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(
				txCtx,
				"failed to get excluded reservation types",
				"error", err,
				"id", id,
				"clinic_id", clinicID,
			)
			return apperrors.Wrap(err, "failed to get excluded reservation types")
		}
		return nil
	}); err != nil {
		return nil, nil, err //nolint:wrapcheck // tx 閉包内の 3 分岐とも文脈付き wrap 済み（同義二重ラップ回避）
	}
	slog.InfoContext(ctx, "reservation staff updated",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, excluded, nil
}

func (s *reservationStaffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if s.staffDeleter == nil {
		return apperrors.WrapInternalServerError("reservation staff deleter is not configured")
	}
	if err := s.staffDeleter.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation staff")
	}
	slog.InfoContext(ctx, "reservation staff deleted",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationStaffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error) {
	var result *model.Staff
	var excluded []model.StaffReservationExclusion
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockForMutation(txCtx, clinicID, id); err != nil {
			slog.ErrorContext(txCtx, "failed to verify reservation staff ownership", "error", err)
			return apperrors.Wrap(err, "failed to verify reservation staff ownership")
		}
		if err := s.repo.Update(txCtx, clinicID, id, map[string]any{"is_active": isActive}); err != nil {
			slog.ErrorContext(txCtx, "failed to patch staff status", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to patch staff status")
		}
		var err error
		result, err = s.repo.FindByID(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get reservation staff after patch status", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to get reservation staff after patch status")
		}
		excluded, err = s.repo.FindAllExcludedReservationTypes(txCtx, clinicID, id)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get excluded reservation types", "error", err, "id", id, "clinic_id", clinicID)
			return apperrors.Wrap(err, "failed to get excluded reservation types")
		}
		return nil
	}); err != nil {
		return nil, nil, err //nolint:wrapcheck // tx 閉包内の全分岐を文脈付き wrap 済み
	}
	slog.InfoContext(ctx, "reservation staff status patched",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("is_active", isActive))
	return result, excluded, nil
}

func (s *reservationStaffService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if direction != "up" && direction != "down" {
		return apperrors.WrapInvalidInput("direction must be 'up' or 'down'")
	}
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to verify reservation staff ownership", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify reservation staff ownership")
	}
	if err := s.repo.UpdateSortOrder(ctx, clinicID, id, direction); err != nil {
		slog.ErrorContext(ctx, "failed to reorder reservation staff", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder reservation staff")
	}
	return nil
}

// ListExcludedByStaffIDs は複数スタッフの除外コースをバルク取得してスタッフID→除外コース一覧のマップを返す
// Stage B: derived from capabilities (exclusion table is not the SoT).
func (s *reservationStaffService) ListExcludedByStaffIDs(
	ctx context.Context,
	clinicID uint64,
	staffIDs []uint64,
) (map[uint64][]model.StaffReservationExclusion, error) {
	items, err := s.repo.FindAllExcludedReservationTypesByStaffIDs(ctx, clinicID, staffIDs)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list excluded service types", "error", err)
		return nil, apperrors.Wrap(err, "failed to list excluded service types")
	}
	m := make(map[uint64][]model.StaffReservationExclusion, len(staffIDs))
	for i := range items {
		sid := items[i].StaffID
		m[sid] = append(m[sid], items[i])
	}
	return m, nil
}

// ListCapableByStaffIDs は複数スタッフの対応可能コースをバルク取得してスタッフID→一覧のマップを返す
func (s *reservationStaffService) ListCapableByStaffIDs(
	ctx context.Context,
	clinicID uint64,
	staffIDs []uint64,
) (map[uint64][]model.StaffReservationCapability, error) {
	items, err := s.repo.FindAllReservationCapabilitiesByStaffIDs(ctx, clinicID, staffIDs)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list capable reservation types", "error", err)
		return nil, apperrors.Wrap(err, "failed to list capable reservation types")
	}
	m := make(map[uint64][]model.StaffReservationCapability, len(staffIDs))
	for i := range items {
		sid := items[i].StaffID
		m[sid] = append(m[sid], items[i])
	}
	return m, nil
}
