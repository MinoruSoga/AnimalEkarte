package trimming

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const courseImageAuditResource = "trimming_course"

// A course mutation takes UPDATE directly, never a SHARE lock that later
// upgrades. The locked row is the before-value source for durable history.
type courseImageUpdateLocker interface {
	LockForImageUpdate(context.Context, uint64, uint64) (*model.TrimmingCourse, error)
}

func (s *trimmingCourseService) requireCourseImageAudit(actorID uint64) error {
	if actorID == 0 {
		return apperrors.WrapInvalidInput("course image history requires a staff actor")
	}
	if s.imageAudit == nil {
		return apperrors.WrapInternalServerError("course image history is not configured")
	}
	return nil
}

func (s *trimmingCourseService) logCourseImageHistory(ctx context.Context, clinicID, actorID, courseID uint64, action string, oldKey, newKey *string) error {
	if err := s.requireCourseImageAudit(actorID); err != nil {
		return err
	}
	if err := s.imageAudit.LogEntryTx(ctx, &AuditEntry{
		ClinicID: &clinicID, ActorID: &actorID, ActorType: model.AuditActorTypeStaff,
		Action: "trimming_course.image_" + action, Resource: courseImageAuditResource, ResourceID: &courseID,
		OldValue: map[string]any{"image_object_key": trimmingAuditStringValue(oldKey)},
		NewValue: map[string]any{"image_object_key": trimmingAuditStringValue(newKey)},
		Metadata: map[string]any{"mutation_path": "course_image_reference", "original_retained": true},
	}); err != nil {
		return apperrors.Wrap(err, "failed to retain course image reference history")
	}
	return nil
}
