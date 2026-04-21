package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateReservationStaffInput は予約スタッフ作成の入力データ
type CreateReservationStaffInput struct {
	Name               string
	StaffType          string
	ReservationVisible bool
	ReservationComment string
	SortOrder          int
	ExcludedTypeIDs    []uint64
}

// UpdateReservationStaffInput は予約スタッフ更新の入力データ（ポインタ型でゼロ値を区別）
type UpdateReservationStaffInput struct {
	Name               *string
	StaffType          *string
	ReservationVisible *bool
	ReservationComment *string
	SortOrder          *int
	ExcludedTypeIDs    *[]uint64
}

const (
	colReservationStaffName               = "name"
	colReservationStaffStaffType          = "staff_type"
	colReservationStaffReservationVisible = "reservation_visible"
	colReservationStaffReservationComment = "reservation_comment"
	colReservationStaffSortOrder          = "sort_order"
)

func buildReservationStaffUpdateFields(input *UpdateReservationStaffInput) map[string]any {
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
	// GetExcludedReservationTypes は指定スタッフの除外コース一覧を返す
	GetExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	// ListExcludedByStaffIDs は複数スタッフの除外コースをバルク取得してスタッフID→除外コース一覧のマップを返す（N+1回避）
	ListExcludedByStaffIDs(ctx context.Context, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error)
}

// ReservationStaffService は ReservationStaffCoreService / ReservationStaffExclusionService を統合したインターフェース。
type ReservationStaffService interface {
	ReservationStaffCoreService
	ReservationStaffExclusionService
}

type reservationStaffService struct {
	repo       repository.ReservationStaffRepository
	transactor repository.Transactor
}

func NewReservationStaffService(repo repository.ReservationStaffRepository, transactor repository.Transactor) ReservationStaffService {
	return &reservationStaffService{repo: repo, transactor: transactor}
}

func (s *reservationStaffService) List(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	staffs, err := s.repo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation staffs")
	}
	return staffs, nil
}

func (s *reservationStaffService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByIDAndClinicID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation staff")
	}
	return staff, nil
}

func (s *reservationStaffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	staffType := model.StaffType(input.StaffType)
	if staffType == "" {
		staffType = model.StaffTypeDoctor
	}
	staff := &model.Staff{
		Name:               input.Name,
		IsActive:           true,
		SortOrder:          input.SortOrder,
		StaffType:          staffType,
		ReservationVisible: input.ReservationVisible,
		ReservationComment: input.ReservationComment,
	}
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, staff, clinicID); err != nil {
			return apperrors.Wrap(err, "failed to create reservation staff")
		}
		if len(input.ExcludedTypeIDs) > 0 {
			if err := s.repo.ReplaceExcludedReservationTypes(txCtx, staff.ID, input.ExcludedTypeIDs); err != nil {
				return apperrors.Wrap(err, "failed to set excluded courses")
			}
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	slog.InfoContext(ctx, "reservation staff created",
		slog.Uint64("staff_id", staff.ID),
		slog.Uint64("clinic_id", clinicID))
	excluded, err := s.repo.FindExcludedReservationTypes(ctx, staff.ID)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
	}
	return staff, excluded, nil
}

func (s *reservationStaffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	// clinicID 確認
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to verify reservation staff ownership")
	}

	fields := buildReservationStaffUpdateFields(input)
	if len(fields) > 0 {
		if err := s.repo.UpdateFields(ctx, id, fields); err != nil {
			return nil, nil, apperrors.Wrap(err, "failed to update reservation staff")
		}
	}
	if input.ExcludedTypeIDs != nil {
		if err := s.repo.ReplaceExcludedReservationTypes(ctx, id, *input.ExcludedTypeIDs); err != nil {
			return nil, nil, apperrors.Wrap(err, "failed to update excluded courses")
		}
	}
	updated, err := s.repo.FindByIDAndClinicID(ctx, clinicID, id)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to get reservation staff after update")
	}
	slog.InfoContext(ctx, "reservation staff updated",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	excluded, err := s.repo.FindExcludedReservationTypes(ctx, id)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
	}
	return updated, excluded, nil
}

func (s *reservationStaffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to verify reservation staff ownership")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation staff")
	}
	slog.InfoContext(ctx, "reservation staff deleted",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationStaffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error) {
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to verify reservation staff ownership")
	}
	if err := s.repo.UpdateFields(ctx, id, map[string]any{"is_active": isActive}); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to patch staff status")
	}
	result, err := s.repo.FindByIDAndClinicID(ctx, clinicID, id)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to get reservation staff after patch status")
	}
	slog.InfoContext(ctx, "reservation staff status patched",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("is_active", isActive))
	excluded, err := s.repo.FindExcludedReservationTypes(ctx, id)
	if err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
	}
	return result, excluded, nil
}

func (s *reservationStaffService) PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if direction != "up" && direction != "down" {
		return apperrors.WrapInvalidInput("direction must be 'up' or 'down'")
	}
	if err := s.repo.SwapSortOrder(ctx, clinicID, id, direction); err != nil {
		return apperrors.Wrap(err, "failed to reorder reservation staff")
	}
	return nil
}

// GetExcludedReservationTypes は指定スタッフの除外コース一覧を返す
func (s *reservationStaffService) GetExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	items, err := s.repo.FindExcludedReservationTypes(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get excluded service types")
	}
	return items, nil
}

// ListExcludedByStaffIDs は複数スタッフの除外コースをバルク取得してスタッフID→除外コース一覧のマップを返す
func (s *reservationStaffService) ListExcludedByStaffIDs(ctx context.Context, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
	items, err := s.repo.FindExcludedReservationTypesByStaffIDs(ctx, staffIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list excluded service types")
	}
	m := make(map[uint64][]model.StaffReservationExclusion, len(staffIDs))
	for i := range items {
		sid := items[i].StaffID
		m[sid] = append(m[sid], items[i])
	}
	return m, nil
}
