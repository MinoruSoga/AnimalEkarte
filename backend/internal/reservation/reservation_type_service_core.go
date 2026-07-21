package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

func (s *reservationTypeService) List(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	items, err := s.repo.FindAllWithChildren(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list reservation types", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to list reservation types")
	}
	return items, nil
}

func (s *reservationTypeService) GetByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	result, err := s.repo.FindByIDWithChildren(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	return result, nil
}

func (s *reservationTypeService) Create(ctx context.Context, clinicID uint64, input *CreateReservationTypeInput) (*model.ReservationType, error) {
	if err := sharedkernel.ValidateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if err := validateMaxConcurrent(input.MaxConcurrent); err != nil {
		return nil, err
	}
	if err := s.validateReservationTypeParent(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := s.validateReservationTypeGroup(ctx, clinicID, input.GroupID); err != nil {
		return nil, err
	}
	reservationDayOption := model.ReservationDayOption(input.ReservationDayOption)
	if reservationDayOption == "" {
		reservationDayOption = model.DayOptionNone
	}
	durationMinutes := 15
	if input.DurationMinutes != nil {
		durationMinutes = *input.DurationMinutes
	}
	reservationVisible := true
	if input.ReservationVisible != nil {
		reservationVisible = *input.ReservationVisible
	}
	category := model.ReservationTypeCategory(input.Category)
	if category == "" {
		category = model.ReservationTypeCategoryGeneral
	}

	st := &model.ReservationType{
		ClinicID:               clinicID,
		Name:                   input.Name,
		Color:                  input.Color,
		IsActive:               input.IsActive,
		Description:            input.Description,
		SortOrder:              input.SortOrder,
		Category:               category,
		ParentID:               input.ParentID,
		ReservationDisplayName: input.ReservationDisplayName,
		DurationMinutes:        durationMinutes,
		MaxConcurrent:          input.MaxConcurrent,
		ShortName:              input.ShortName,
		ShowShortName:          input.ShowShortName,
		ReservationVisible:     reservationVisible,
		ReservationComment:     input.ReservationComment,
		ReservationImageURL:    input.ReservationImageURL,
		ReservationDayOption:   reservationDayOption,
		IsInternal:             input.IsInternal,
		GroupID:                input.GroupID,
	}
	if err := s.repo.Create(ctx, st); err != nil {
		slog.ErrorContext(ctx, "failed to create reservation type", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to create reservation type")
	}
	slog.InfoContext(ctx, "reservation type created", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", st.ID))
	created, err := s.repo.FindByIDWithChildren(ctx, clinicID, st.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type after create", "error", err, "id", st.ID, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation type after create")
	}
	return created, nil
}

func (s *reservationTypeService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeInput) (*model.ReservationType, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgInputNotNil)
	}
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation type")
	}
	if err := sharedkernel.ValidateOptionalName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate optional name")
	}
	if err := validateMaxConcurrent(input.MaxConcurrent); err != nil {
		return nil, err
	}
	if input.ClearParentID && input.ParentID != nil {
		return nil, apperrors.WrapInvalidInput("clear_parent_id と parent_id は同時に指定できません")
	}
	if input.ParentID != nil && *input.ParentID == id {
		return nil, apperrors.WrapInvalidInput("自分自身を親に設定できません")
	}
	if input.ParentID != nil {
		childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
		if err != nil {
			slog.ErrorContext(ctx, "failed to count reservation type children", "error", err, "id", id, "clinic_id", clinicID)
			return nil, apperrors.Wrap(err, "failed to count reservation type children")
		}
		if childCount > 0 {
			return nil, apperrors.WrapInvalidInput("子予約区分を持つ予約区分は子階層に移動できません")
		}
	}
	if err := s.validateReservationTypeParent(ctx, clinicID, input.ParentID); err != nil {
		return nil, err
	}
	if err := s.validateReservationTypeGroup(ctx, clinicID, input.GroupID); err != nil {
		return nil, err
	}
	fields := buildReservationTypeUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	if _, err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to update reservation type")
	}
	result, err := s.repo.FindByIDWithChildren(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation type after update", "error", err, "id", id, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get reservation type after update")
	}
	slog.InfoContext(ctx, "reservation type updated", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", id))
	return result, nil
}

func (s *reservationTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
	if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to find reservation type")
	}
	childCount, err := s.repo.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check reservation type children", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check reservation type children")
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この予約区分には子予約区分が登録されているため削除できません")
	}
	count, err := s.repo.CountUsageByReservationTypeID(ctx, clinicID, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check reservation type usage", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to check reservation type usage")
	}
	if count > 0 {
		return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		slog.ErrorContext(ctx, "failed to delete reservation type", "error", err, "id", id, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to delete reservation type")
	}
	slog.InfoContext(ctx, "reservation type deleted", slog.Uint64("clinic_id", clinicID), slog.Uint64("reservation_type_id", id))
	return nil
}

func (s *reservationTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(sharedkernel.ErrMsgIDsNotEmpty)
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		slog.ErrorContext(ctx, "failed to reorder reservation types", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to reorder reservation types")
	}
	slog.InfoContext(ctx, "reservation type reordered", slog.Uint64("clinic_id", clinicID), slog.Int("count", len(ids)))
	return nil
}

func validateMaxConcurrent(maxConcurrent *int) error {
	if maxConcurrent != nil && *maxConcurrent <= 0 {
		return apperrors.WrapInvalidInput("max_concurrent は1以上で指定してください")
	}
	return nil
}

func (s *reservationTypeService) validateReservationTypeParent(ctx context.Context, clinicID uint64, parentID *uint64) error {
	if parentID == nil {
		return nil
	}
	parent, err := s.repo.FindByID(ctx, clinicID, *parentID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find parent reservation type", "error", err, "parent_id", *parentID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to find parent reservation type")
	}
	if parent.ParentID != nil {
		return apperrors.WrapInvalidInput("親予約区分にはルートノードのみ指定できます")
	}
	return nil
}

// validateReservationTypeGroup は BE-refactor.md X-14/U6b のクロステナント write 防止:
// GroupID が caller の clinic に属することを INSERT/UPDATE 前に検証する
// (validateReservationTypeParent / reservation_validators.go の typeRepo と同型の
// nil 許容パターン。groupRepo 未設定時はチェックをスキップする — production は常に非 nil)。
func (s *reservationTypeService) validateReservationTypeGroup(ctx context.Context, clinicID uint64, groupID *uint64) error {
	if s.groupRepo == nil || groupID == nil {
		return nil
	}
	if _, err := s.groupRepo.FindByID(ctx, clinicID, *groupID); err != nil {
		slog.ErrorContext(ctx, "reservation type group not found or belongs to different clinic", "error", err, "group_id", *groupID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify reservation type group ownership")
	}
	return nil
}
