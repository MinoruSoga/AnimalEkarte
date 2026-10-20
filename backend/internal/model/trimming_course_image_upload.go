package model

import "time"

// TrimmingCourseImageUpload is retained provenance, never a TTL deletion queue.
type TrimmingCourseImageUpload struct {
	ClinicID        uint64 `gorm:"primaryKey"`
	ObjectKey       string `gorm:"primaryKey"`
	UploaderStaffID uint64
	CreatedAt       time.Time
	UploadedAt      *time.Time
	SubmittedAt     *time.Time
}
