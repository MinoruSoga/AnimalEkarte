package reservation

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetSettings はLIFF公開設定を返す（機密フィールドは除外）。
func (s *liffService) GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	setting, err := s.settingRepo.FindByClinicID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get reservation setting", "error", err)
		return nil, apperrors.Wrap(err, "failed to get reservation setting")
	}
	return setting, nil
}

// GetProfile は顧客プロフィールを返す。
func (s *liffService) GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
	c, err := s.customerRepo.FindByID(ctx, clinicID, customerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get customer profile", "error", err)
		return nil, apperrors.Wrap(err, "failed to get customer profile")
	}
	return c, nil
}

// GetTrimmingCourses はLIFF向けトリミングコース一覧を返す（is_active=true のみ）。
func (s *liffService) GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	all, err := s.trimmingCourseRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimming courses", "error", err)
		return nil, apperrors.Wrap(err, "failed to get trimming courses")
	}
	result := make([]model.TrimmingCourse, 0, len(all))
	for i := range all {
		if all[i].IsActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

// GetTrimmingOptions はLIFF向けトリミングオプション一覧を返す（is_active=true のみ）。
func (s *liffService) GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	all, err := s.trimmingOptionRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get trimming options", "error", err)
		return nil, apperrors.Wrap(err, "failed to get trimming options")
	}
	result := make([]model.TrimmingOption, 0, len(all))
	for i := range all {
		if all[i].IsActive {
			result = append(result, all[i])
		}
	}
	return result, nil
}

// GetCourses はLIFF向け公開コース一覧を返す（有効かつ公開中の葉ノードのみ）。
func (s *liffService) GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	all, err := s.typeLiffRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get courses", "error", err)
		return nil, apperrors.Wrap(err, "failed to get courses")
	}
	hasChildren := make(map[uint64]struct{}, len(all))
	for i := range all {
		if all[i].ParentID != nil {
			hasChildren[*all[i].ParentID] = struct{}{}
		}
	}
	result := make([]model.ReservationType, 0, len(all))
	for i := range all {
		if !all[i].IsActive || all[i].IsInternal || !all[i].ReservationVisible {
			continue
		}
		if _, isParent := hasChildren[all[i].ID]; isParent {
			continue
		}
		result = append(result, all[i])
	}
	return result, nil
}

func (s *liffService) findLiffCourse(ctx context.Context, clinicID, typeID uint64) (*model.ReservationType, error) {
	course, err := s.typeLiffRepo.FindByID(ctx, clinicID, typeID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get course", "error", err)
		return nil, apperrors.Wrap(err, "failed to get course")
	}
	return course, nil
}

func (s *liffService) findActiveLiffCourse(ctx context.Context, clinicID, typeID uint64) (*model.ReservationType, error) {
	course, err := s.findLiffCourse(ctx, clinicID, typeID)
	if err != nil {
		return nil, err
	}
	// Same LINE-create gate as reservation_validators.validateReservationMasterOwnership.
	// Inactive, internal, and reservation-invisible types share one invalid-input so
	// existence of hidden types is not confirmed by a distinct error.
	if !course.IsActive || !course.ReservationVisible || course.IsInternal {
		return nil, apperrors.WrapInvalidInput("reservation type is not available for LINE reservation")
	}
	return course, nil
}

// GetStaffs は有効な予約区分の対応スタッフ一覧を返す（reservation_visible=true && typeIDに対応可能）。
func (s *liffService) GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error) {
	if _, err := s.findActiveLiffCourse(ctx, clinicID, typeID); err != nil {
		return nil, err
	}
	all, err := s.staffRepo.FindAll(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get staffs", "error", err)
		return nil, apperrors.Wrap(err, "failed to get staffs")
	}
	return s.filterVisibleStaffsByTypeID(ctx, clinicID, typeID, all)
}
