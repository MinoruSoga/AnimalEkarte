// Package service provides business logic implementations for Trimming entity.
package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *trimmingService) validateTrimmingReservationType(ctx context.Context, clinicID, reservationTypeID uint64, requireActive bool) error {
	if s.reservationType == nil {
		return apperrors.WrapInternalServerError("reservation type repository is required")
	}

	reservationType, err := s.reservationType.FindByID(ctx, clinicID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get trimming reservation type")
	}
	if reservationType.Category != model.ReservationTypeCategoryTrimming {
		return apperrors.WrapInvalidInput("reservation_type_id must be a trimming reservation type")
	}
	if requireActive && !reservationType.IsActive {
		return apperrors.WrapInvalidInput("reservation_type_id references an inactive trimming reservation type")
	}
	return nil
}

func (s *trimmingService) requireBookingConstraintDependencies() error {
	if s.unavailableTime == nil {
		return apperrors.WrapInternalServerError("reservation type unavailable-time repository is required")
	}
	if s.reservation == nil {
		return apperrors.WrapInternalServerError("reservation repository is required")
	}
	if s.reservationType == nil {
		return apperrors.WrapInternalServerError("reservation type repository is required")
	}
	return nil
}

// validateTrimmingCourseAndOptions は appointment_trimming_detail の CourseID / OptionIDs が
// caller の clinic に属し、かつ is_active であることを永続化前に検証する（X-14c: 2 repo ガード
// + #228: 無効化されたコース/オプションを新規に紐付けさせない）。request にIDがあるのに
// 対応repositoryが未DIなら、clinic ownershipを検証できないためfail-closedにする。
// existingCourseID/existingOptionIDs はカルテに既に紐づいている ID（Create 系呼び出しでは常に
// nil/空 — 新規紐付けはすべて active 必須）。既にリンク済みの ID は is_active チェックを免除し
// （データを消さない）、新規に追加される ID のみ active を要求する。
func (s *trimmingService) validateTrimmingCourseAndOptions(
	ctx context.Context, clinicID uint64, courseID *uint64, optionIDs []uint64,
	existingCourseID *uint64, existingOptionIDs []uint64,
) error {
	if courseID != nil && s.trimmingCourseRepo == nil {
		return apperrors.WrapInternalServerError("trimming course repository is required")
	}
	if len(optionIDs) > 0 && s.trimmingOptionRepo == nil {
		return apperrors.WrapInternalServerError("trimming option repository is required")
	}
	if s.trimmingCourseRepo != nil {
		if err := validateOwnedMasterFK(ctx, "trimming course", clinicID, courseID,
			func(actx context.Context, cid, mid uint64) error {
				course, err := s.trimmingCourseRepo.FindByID(actx, cid, mid)
				if err != nil {
					return err
				}
				if !course.IsActive && (existingCourseID == nil || *existingCourseID != mid) {
					return apperrors.WrapInvalidInput("course_id references an inactive trimming course")
				}
				return nil
			}); err != nil {
			return err
		}
	}
	if s.trimmingOptionRepo != nil {
		existingOptions := make(map[uint64]bool, len(existingOptionIDs))
		for _, existingID := range existingOptionIDs {
			existingOptions[existingID] = true
		}
		if err := validateOwnedMasterFKs(ctx, "trimming option", clinicID, optionIDs,
			func(actx context.Context, cid, mid uint64) error {
				option, err := s.trimmingOptionRepo.FindByID(actx, cid, mid)
				if err != nil {
					return err
				}
				if !option.IsActive && !existingOptions[mid] {
					return apperrors.WrapInvalidInput("option_ids references an inactive trimming option")
				}
				return nil
			}); err != nil {
			return err
		}
	}
	return nil
}
