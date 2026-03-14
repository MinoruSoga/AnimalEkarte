package handler

import "time"

// createRecordImageRequest は診療画像作成のバインド struct
type createRecordImageRequest struct {
	ImageURL     string     `json:"image_url"     binding:"required"`
	ThumbnailURL string     `json:"thumbnail_url"`
	FileName     string     `json:"file_name"`
	FileSize     int64      `json:"file_size"`
	MimeType     string     `json:"mime_type"`
	ImageType    string     `json:"image_type"`
	Description  string     `json:"description"`
	TakenAt      *time.Time `json:"taken_at"`
	ExamID       *uint64    `json:"exam_id"`
	StaffID      *uint64    `json:"staff_id"`
	SortOrder    int        `json:"sort_order"`
}
