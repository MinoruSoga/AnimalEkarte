package service

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
			slog.ErrorContext(txCtx, "failed to create reservation staff", "error", err)
			return apperrors.Wrap(err, "failed to create reservation staff")
		}
		if len(input.ExcludedTypeIDs) > 0 {
			if err := s.repo.UpdateExcludedReservationTypes(txCtx, clinicID, staff.ID, input.ExcludedTypeIDs); err != nil {
				slog.ErrorContext(txCtx, "failed to set excluded courses", "error", err)
				return apperrors.Wrap(err, "failed to set excluded courses")
			}
		}
		return nil
	}); err != nil {
		return nil, nil, apperrors.Wrap(err, "failed to create reservation staff")
	}
	slog.InfoContext(ctx, "reservation staff created",
		slog.Uint64("staff_id", staff.ID),
		slog.Uint64("clinic_id", clinicID))
	excluded, err := s.repo.FindAllExcludedReservationTypes(ctx, staff.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get excluded reservation types", "error", err, "clinic_id", clinicID)
		return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
	}
	return staff, excluded, nil
}

func (s *reservationStaffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationStaffInput) (*model.Staff, []model.StaffReservationExclusion, error) {
	// BE-refactor.md X-8: staff 本体更新 + 除外コース置換を WithTx で括り原子化する（Create と対称）。
	// 括らないと、fields 更新が成功し UpdateExcludedReservationTypes が失敗した場合に
	// staff 側の変更（名前/種別/表示可否等）だけがコミットされ、除外コースは古いまま残る
	// 非原子な部分更新になる。
	// BE-refactor.md H-7: clinicID 確認（所有権確認）を WithTx 閉包の外から先頭へ移動し、
	// dbOrTx 化済みの repo.FindByID を txCtx で呼ぶことで ambient tx に参加させる。
	// これにより確認〜更新の間にスタッフ削除/所属変更が起きる TOCTOU 窓を閉じる
	// （repo.Update は R13 で updateScopedByID 化済み・RowsAffected==0 は WrapNotFound）。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.FindByID(txCtx, clinicID, id); err != nil {
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
		if input.ExcludedTypeIDs != nil {
			if err := s.repo.UpdateExcludedReservationTypes(txCtx, clinicID, id, *input.ExcludedTypeIDs); err != nil {
				slog.ErrorContext(txCtx, "failed to update excluded courses", "error", err, "id", id, "clinic_id", clinicID)
				return apperrors.Wrap(err, "failed to update excluded courses")
			}
		}
		return nil
	}); err != nil {
		return nil, nil, err //nolint:wrapcheck // tx 閉包内の 3 分岐とも文脈付き wrap 済み（同義二重ラップ回避）
	}
	updated, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation staff after update", "error", err, "id", id, "clinic_id", clinicID)
		return nil, nil, apperrors.Wrap(err, "failed to get reservation staff after update")
	}
	slog.InfoContext(ctx, "reservation staff updated",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	excluded, err := s.repo.FindAllExcludedReservationTypes(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get excluded reservation types", "error", err, "id", id, "clinic_id", clinicID)
		return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
	}
	return updated, excluded, nil
}

func (s *reservationStaffService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to verify reservation staff ownership")
	}
	count, err := s.repo.CountUsageByStaffID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count reservation staff usage", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to count reservation staff usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("このスタッフは予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete reservation staff", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete reservation staff")
	}
	slog.InfoContext(ctx, "reservation staff deleted",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}

func (s *reservationStaffService) PatchStatus(ctx context.Context, clinicID, id uint64, isActive bool) (*model.Staff, []model.StaffReservationExclusion, error) {
	if _, err := s.GetByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to verify reservation staff ownership", "error", err)
		return nil, nil, apperrors.Wrap(err, "failed to verify reservation staff ownership")
	}
	if err := s.repo.Update(ctx, clinicID, id, map[string]any{"is_active": isActive}); err != nil {
		slog.ErrorContext(ctx, "failed to patch staff status", "error", err, "id", id, "clinic_id", clinicID)
		return nil, nil, apperrors.Wrap(err, "failed to patch staff status")
	}
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation staff after patch status", "error", err, "id", id, "clinic_id", clinicID)
		return nil, nil, apperrors.Wrap(err, "failed to get reservation staff after patch status")
	}
	slog.InfoContext(ctx, "reservation staff status patched",
		slog.Uint64("staff_id", id),
		slog.Uint64("clinic_id", clinicID),
		slog.Bool("is_active", isActive))
	excluded, err := s.repo.FindAllExcludedReservationTypes(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get excluded reservation types", "error", err, "id", id, "clinic_id", clinicID)
		return nil, nil, apperrors.Wrap(err, "failed to get excluded reservation types")
	}
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
func (s *reservationStaffService) ListExcludedByStaffIDs(ctx context.Context, staffIDs []uint64) (map[uint64][]model.StaffReservationExclusion, error) {
	items, err := s.repo.FindAllExcludedReservationTypesByStaffIDs(ctx, staffIDs)
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
