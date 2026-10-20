package trimming

import (
	"context"
	"errors"
	"regexp"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"gorm.io/gorm"
)

// CourseImageDrafts owns committed, irreversible upload/submission provenance.
// HTTP composition must enforce the current selected-clinic master grant.
type CourseImageDrafts interface {
	ReserveUpload(context.Context, uint64, uint64, string) error
	CompleteUpload(context.Context, uint64, uint64, string) error
	ProtectSubmission(context.Context, uint64, uint64, string) error
	ReserveDiscard(context.Context, uint64, uint64, string) (bool, error)
}
type courseImageDraftRepository interface {
	Find(context.Context, ImageKey) (*model.TrimmingCourseImageUpload, error)
	Reserve(context.Context, ImageKey, uint64) error
	MarkUploaded(context.Context, ImageKey, uint64) error
	MarkSubmitted(context.Context, ImageKey) error
}
type courseImageDrafts struct {
	tx        Transactor
	repo      courseImageDraftRepository
	lifecycle ImageLifecycleRepository
}

func (s *courseImageDrafts) locked(ctx context.Context, clinic, actor uint64, stored string, fn func(context.Context, ImageKey) error) error {
	key, err := ownedImageKey(clinic, stored)
	if err != nil || key.value != stored || actor == 0 || clinic == 0 {
		return apperrors.WrapInvalidInput("canonical course image key and actor required")
	}
	if s.tx == nil || s.repo == nil || s.lifecycle == nil {
		return apperrors.WrapInternalServerError("course image provenance unavailable")
	}
	// Never join the business save transaction: rejected/unknown saves retain proof.
	return s.tx.WithTx(persistence.DetachTx(ctx), func(txCtx context.Context) error {
		if err := s.lifecycle.LockImageKeys(txCtx, []ImageKey{key}); err != nil {
			return err
		}
		return fn(txCtx, key)
	})
}

var generatedCourseImageKey = regexp.MustCompile(`^trimming-images/[1-9][0-9]*/[0-9]+_[0-9a-f]{32}\.(jpg|jpeg|png|gif)$`)

func (s *courseImageDrafts) ReserveUpload(ctx context.Context, c, a uint64, k string) error {
	if !generatedCourseImageKey.MatchString(k) {
		return apperrors.WrapInvalidInput("invalid generated course image key")
	}
	return s.locked(ctx, c, a, k, func(ctx context.Context, key ImageKey) error {
		gone, err := s.lifecycle.ImageKeyDeleted(ctx, key)
		if err != nil {
			return err
		}
		if gone {
			return apperrors.WrapConflict("course image key retired")
		}
		used, err := s.lifecycle.ImageKeyInUse(ctx, c, k, 0)
		if err != nil {
			return err
		}
		if used {
			return apperrors.WrapConflict("course image original is protected")
		}
		return s.repo.Reserve(ctx, key, a)
	})
}
func (s *courseImageDrafts) CompleteUpload(ctx context.Context, c, a uint64, k string) error {
	return s.locked(ctx, c, a, k, func(ctx context.Context, key ImageKey) error { return s.repo.MarkUploaded(ctx, key, a) })
}
func (s *courseImageDrafts) ProtectSubmission(ctx context.Context, c, a uint64, k string) error {
	unknown, otherActor := false, false
	err := s.locked(ctx, c, a, k, func(ctx context.Context, key ImageKey) error {
		gone, err := s.lifecycle.ImageKeyDeleted(ctx, key)
		if err != nil {
			return err
		}
		if gone {
			return apperrors.WrapConflict("course image key retired; upload again")
		}
		row, err := s.repo.Find(ctx, key)
		if err != nil {
			return err
		}
		if row == nil {
			return nil
		}
		unknown = row.UploadedAt == nil
		otherActor = row.UploaderStaffID != a
		// The mark is allowed even during upload. A late upload completion cannot
		// turn a submitted/unknown image back into an unsubmitted draft.
		return s.repo.MarkSubmitted(ctx, key)
	})
	if err != nil {
		return err
	}
	if unknown {
		return apperrors.WrapConflict("course image upload result is unknown")
	}
	if otherActor {
		used := false
		if err := s.locked(ctx, c, a, k, func(ctx context.Context, key ImageKey) error {
			var err error
			used, err = s.lifecycle.ImageKeyInUse(ctx, c, key.value, 0)
			return err
		}); err != nil {
			return err
		}
		if !used {
			return apperrors.WrapConflict("another staff's draft is protected")
		}
	}
	return nil
}
func (s *courseImageDrafts) ReserveDiscard(ctx context.Context, c, a uint64, k string) (bool, error) {
	reserved := false
	err := s.locked(ctx, c, a, k, func(ctx context.Context, key ImageKey) error {
		row, err := s.repo.Find(ctx, key)
		if err != nil {
			return err
		}
		if row == nil || row.UploaderStaffID != a || row.UploadedAt == nil || row.SubmittedAt != nil {
			return nil
		}
		used, err := s.lifecycle.ImageKeyInUse(ctx, c, k, 0)
		if err != nil {
			return err
		}
		if used {
			return nil
		}
		if err := s.lifecycle.MarkImageKeyDeleted(ctx, key, a); err != nil {
			return err
		}
		reserved = true
		return nil
	})
	return reserved, err
}

type courseImageUploadRepository struct{ db *gorm.DB }

func (r *courseImageUploadRepository) Find(ctx context.Context, key ImageKey) (*model.TrimmingCourseImageUpload, error) {
	var row model.TrimmingCourseImageUpload
	err := persistence.DBOrTx(ctx, r.db).Where("clinic_id = ? AND object_key = ?", key.clinicID, key.value).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}
func (r *courseImageUploadRepository) Reserve(ctx context.Context, key ImageKey, a uint64) error {
	return persistence.DBOrTx(ctx, r.db).Create(&model.TrimmingCourseImageUpload{ClinicID: key.clinicID, ObjectKey: key.value, UploaderStaffID: a}).Error
}
func (r *courseImageUploadRepository) MarkUploaded(ctx context.Context, key ImageKey, a uint64) error {
	result := persistence.DBOrTx(ctx, r.db).Exec(`UPDATE trimming_course_image_uploads SET uploaded_at = COALESCE(uploaded_at, NOW()) WHERE clinic_id = ? AND object_key = ? AND uploader_staff_id = ?`, key.clinicID, key.value, a)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapConflict("course upload provenance missing")
	}
	return nil
}
func (r *courseImageUploadRepository) MarkSubmitted(ctx context.Context, key ImageKey) error {
	result := persistence.DBOrTx(ctx, r.db).Exec(`UPDATE trimming_course_image_uploads SET submitted_at = COALESCE(submitted_at, NOW()) WHERE clinic_id = ? AND object_key = ?`, key.clinicID, key.value)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapConflict("course upload provenance missing")
	}
	return nil
}
