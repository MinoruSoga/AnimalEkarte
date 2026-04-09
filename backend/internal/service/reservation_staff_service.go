package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ReservationStaffService は予約スタッフのビジネスロジックインターフェース
type ReservationStaffService interface {
	List(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, error)
	PatchSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
	// GetExcludedServiceTypes は指定スタッフの除外コース一覧を返す
	GetExcludedServiceTypes(ctx context.Context, staffID uint64) ([]model.StaffExcludedServiceType, error)
	// ListExcludedByStaffIDs は複数スタッフの除外コースをバルク取得してスタッフID→除外コース一覧のマップを返す（N+1回避）
	ListExcludedByStaffIDs(ctx context.Context, staffIDs []uint64) (map[uint64][]model.StaffExcludedServiceType, error)
}

// CreateReservationStaffInput は予約スタッフ作成の入力データ
type CreateReservationStaffInput struct {
	Name               string
	StaffType          string
	ReservationVisible bool
	ReservationComment string
	SortOrder          int
	ExcludedCourseIDs  []uint64
}

// UpdateReservationStaffInput は予約スタッフ更新の入力データ（ポインタ型でゼロ値を区別）
type UpdateReservationStaffInput struct {
	Name               *string
	StaffType          *string
	ReservationVisible *bool
	ReservationComment *string
	SortOrder          *int
	ExcludedCourseIDs  *[]uint64
}

type reservationStaffService struct {
	repo repository.ReservationStaffRepository
}

func NewReservationStaffService(repo repository.ReservationStaffRepository) ReservationStaffService {
	return &reservationStaffService{repo: repo}
}

func (s *reservationStaffService) List(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	staffs, err := s.repo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list reservation staffs")
	}
	return staffs, nil
}

func (s *reservationStaffService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	// clinicID はセキュリティ確認用：FindAllByClinicID でフィルタ済みのものと突き合わせる代わりに
	// 一覧に存在するか検証する
	staffs, err := s.repo.FindAllByClinicID(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation staff")
	}
	for i := range staffs {
		if staffs[i].ID == id {
			return &staffs[i], nil
		}
	}
	return nil, apperrors.WrapNotFound("reservation_staff", "id")
}

func (s *reservationStaffService) Create(ctx context.Context, clinicID uint64, input *CreateReservationStaffInput) (*model.Staff, error) {
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
	if err := s.repo.Create(ctx, staff, clinicID); err != nil {
		return nil, apperrors.Wrap(err, "failed to create reservation staff")
	}
	if len(input.ExcludedCourseIDs) > 0 {
		if err := s.repo.ReplaceExcludedServiceTypes(ctx, staff.ID, input.ExcludedCourseIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to set excluded courses")
		}
	}
	slog.InfoContext(ctx, "reservation staff created",
		slog.Uint64("staff_id", staff.ID),
		slog.Uint64("clinic_id", clinicID))
	return staff, nil
}

func (s *reservationStaffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, error) {
	// clinicID 確認
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return nil, err
	}

	fields := buildReservationStaffUpdateFields(input)
	if len(fields) > 0 {
		if err := s.repo.Update(ctx, id, fields); err != nil {
			return nil, apperrors.Wrap(err, "failed to update reservation staff")
		}
	}
	if input.ExcludedCourseIDs != nil {
		if err := s.repo.ReplaceExcludedServiceTypes(ctx, id, *input.ExcludedCourseIDs); err != nil {
			return nil, apperrors.Wrap(err, "failed to update excluded courses")
		}
	}
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation staff after update")
	}
	slog.InfoContext(ctx, "reservation staff updated",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	return updated, nil
}

func (s *reservationStaffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return err
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation staff")
	}
	slog.InfoContext(ctx, "reservation staff deleted",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationStaffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, error) {
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, id, map[string]any{"is_active": isActive}); err != nil {
		return nil, apperrors.Wrap(err, "failed to patch staff status")
	}
	result, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation staff after patch status")
	}
	return result, nil
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

// GetExcludedServiceTypes は指定スタッフの除外コース一覧を返す
func (s *reservationStaffService) GetExcludedServiceTypes(ctx context.Context, staffID uint64) ([]model.StaffExcludedServiceType, error) {
	items, err := s.repo.FindExcludedServiceTypes(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get excluded service types")
	}
	return items, nil
}

// ListExcludedByStaffIDs は複数スタッフの除外コースをバルク取得してスタッフID→除外コース一覧のマップを返す
func (s *reservationStaffService) ListExcludedByStaffIDs(ctx context.Context, staffIDs []uint64) (map[uint64][]model.StaffExcludedServiceType, error) {
	items, err := s.repo.FindExcludedServiceTypesByStaffIDs(ctx, staffIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list excluded service types")
	}
	m := make(map[uint64][]model.StaffExcludedServiceType, len(staffIDs))
	for i := range items {
		sid := items[i].StaffID
		m[sid] = append(m[sid], items[i])
	}
	return m, nil
}

func buildReservationStaffUpdateFields(input *UpdateReservationStaffInput) map[string]any {
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.StaffType != nil {
		fields["staff_type"] = *input.StaffType
	}
	if input.ReservationVisible != nil {
		fields["reservation_visible"] = *input.ReservationVisible
	}
	if input.ReservationComment != nil {
		fields["reservation_comment"] = *input.ReservationComment
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	return fields
}
