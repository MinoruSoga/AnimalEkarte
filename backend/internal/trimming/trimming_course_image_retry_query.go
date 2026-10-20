package trimming

import "gorm.io/gorm"

// Filter proven candidates before LIMIT so old unproven tombstones cannot
// monopolize the retry batch. The worker rechecks proof and references under lock.
func pendingCourseImageDeletionQuery(db *gorm.DB, clinicID uint64, limit int) *gorm.DB {
	return db.Table("trimming_image_deletions AS d").Select("d.object_key").
		Joins("JOIN trimming_course_image_uploads u ON u.clinic_id = d.clinic_id AND u.object_key = d.object_key AND u.uploader_staff_id = d.actor_staff_id").
		Where("d.clinic_id = ? AND d.storage_deleted_at IS NULL AND u.uploaded_at IS NOT NULL AND u.submitted_at IS NULL", clinicID).
		Order("d.last_attempt_at ASC NULLS FIRST, d.object_key ASC").Limit(limit)
}
