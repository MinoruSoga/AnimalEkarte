package medicalrecord

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	medicalRecordImageMaxUploadSize  = 10 * 1024 * 1024 // 10MB
	medicalRecordImageMaxRequestSize = medicalRecordImageMaxUploadSize + 1024*1024
)

// GIF を許可するのは意図的: カルテ画像は院内保存専用で LINE 配信を経由しないため、
// 歩様など動きの記録に有用な GIF を受け付ける。LINE 配信経路の共有ファイル
// （model/shared_file.go AllowedFileExtensions）は LINE API 制約で GIF 不可 — 同期してはならない。
var allowedMedicalRecordImageMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"application/pdf": true,
}

// createMedicalRecordImageRequest は診療画像 JSON 作成のバインド struct。
// MRC-09: upload 経路と同程度の MIME/size/URL 境界検証を要求する。
type createMedicalRecordImageRequest struct {
	ImageURL     string     `json:"image_url"     binding:"required,url,max=2048"`
	ThumbnailURL string     `json:"thumbnail_url" binding:"omitempty,url,max=2048"`
	FileName     string     `json:"file_name"     binding:"omitempty,max=255"`
	FileSize     int64      `json:"file_size"     binding:"omitempty,min=0"`
	MimeType     string     `json:"mime_type"     binding:"omitempty,max=127"`
	ImageType    string     `json:"image_type"`
	Description  string     `json:"description"   binding:"omitempty,max=2000"`
	TakenAt      *time.Time `json:"taken_at"`
	ExamID       *uint64    `json:"exam_id"`
	StaffID      *uint64    `json:"staff_id"`
	SortOrder    int        `json:"sort_order"`
}

// validateJSONCreate applies upload-parity guards for JSON create (MIME allowlist, size cap).
func (r *createMedicalRecordImageRequest) validateJSONCreate() error {
	if r.FileSize > medicalRecordImageMaxUploadSize {
		return apperrors.WrapInvalidInput(fmt.Sprintf(
			"file size exceeds limit of %dMB",
			medicalRecordImageMaxUploadSize/1024/1024,
		))
	}
	if r.FileSize < 0 {
		return apperrors.WrapInvalidInput("file_size must be non-negative")
	}
	mimeType := strings.TrimSpace(r.MimeType)
	if mimeType == "" {
		// MIME 未指定時は拡張子から推定（upload 経路と同型）。
		ext := strings.ToLower(filepath.Ext(r.FileName))
		if ext == "" {
			ext = strings.ToLower(filepath.Ext(r.ImageURL))
		}
		detected, ok := medicalRecordImageMIMETypeFromExt(ext)
		if !ok {
			return apperrors.WrapInvalidInput("unsupported file type; allowed: jpeg, png, gif, pdf")
		}
		mimeType = detected
	}
	if !allowedMedicalRecordImageMIMETypes[mimeType] {
		return apperrors.WrapInvalidInput("unsupported MIME type: " + mimeType)
	}
	r.MimeType = mimeType
	return nil
}

func (r *createMedicalRecordImageRequest) toServiceInput() *CreateMedicalRecordImageInput {
	imageType := model.MedicalImageType(r.ImageType)
	if imageType == "" {
		imageType = model.MedicalImageTypeOther
	}

	return &CreateMedicalRecordImageInput{
		ImageURL:     r.ImageURL,
		ThumbnailURL: r.ThumbnailURL,
		FileName:     r.FileName,
		FileSize:     r.FileSize,
		MimeType:     r.MimeType,
		ImageType:    imageType,
		Description:  r.Description,
		TakenAt:      r.TakenAt,
		ExamID:       r.ExamID,
		StaffID:      r.StaffID,
		SortOrder:    r.SortOrder,
	}
}

type uploadedMedicalRecordImageInput struct {
	ImageURL string
	FileName string
	FileSize int64
	MimeType string
	TakenAt  *time.Time
}

func (i uploadedMedicalRecordImageInput) toServiceInput() *CreateMedicalRecordImageInput {
	return &CreateMedicalRecordImageInput{
		ImageURL:  i.ImageURL,
		FileName:  i.FileName,
		FileSize:  i.FileSize,
		MimeType:  i.MimeType,
		ImageType: model.MedicalImageTypeOther,
		TakenAt:   i.TakenAt,
	}
}

type medicalRecordImageUploadRequest struct {
	fileHeader *multipart.FileHeader
}

func newMedicalRecordImageUploadRequest(fileHeader *multipart.FileHeader) medicalRecordImageUploadRequest {
	return medicalRecordImageUploadRequest{fileHeader: fileHeader}
}

type medicalRecordImageUploadMeta struct {
	fileName string
	fileSize int64
	fileExt  string
	mimeType string
}

func (r medicalRecordImageUploadRequest) validate() (medicalRecordImageUploadMeta, error) {
	if r.fileHeader.Size > medicalRecordImageMaxUploadSize {
		return medicalRecordImageUploadMeta{}, apperrors.WrapInvalidInput(fmt.Sprintf(
			"file size exceeds limit of %dMB",
			medicalRecordImageMaxUploadSize/1024/1024,
		))
	}

	fileExt := strings.ToLower(filepath.Ext(r.fileHeader.Filename))
	mimeType := r.fileHeader.Header.Get("Content-Type")
	if mimeType == "" || !allowedMedicalRecordImageMIMETypes[mimeType] {
		detectedType, ok := medicalRecordImageMIMETypeFromExt(fileExt)
		if !ok {
			return medicalRecordImageUploadMeta{}, apperrors.WrapInvalidInput("unsupported file type; allowed: jpeg, png, gif, pdf")
		}
		mimeType = detectedType
	}
	if !allowedMedicalRecordImageMIMETypes[mimeType] {
		return medicalRecordImageUploadMeta{}, apperrors.WrapInvalidInput("unsupported MIME type: " + mimeType)
	}

	return medicalRecordImageUploadMeta{
		fileName: r.fileHeader.Filename,
		fileSize: r.fileHeader.Size,
		fileExt:  fileExt,
		mimeType: mimeType,
	}, nil
}

func medicalRecordImageMIMETypeFromExt(fileExt string) (string, bool) {
	switch fileExt {
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".png":
		return "image/png", true
	case ".gif":
		return "image/gif", true
	case ".pdf":
		return "application/pdf", true
	default:
		return "", false
	}
}

func (m medicalRecordImageUploadMeta) newStoredName(now time.Time) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", apperrors.Wrap(err, "failed to generate unique filename")
	}
	return fmt.Sprintf("%d_%s%s", now.UnixNano(), hex.EncodeToString(randomBytes), m.fileExt), nil
}

func (m medicalRecordImageUploadMeta) uploadKey(medicalRecordID uint64, storedName string) string {
	return fmt.Sprintf("medical-records/%d/%s", medicalRecordID, storedName)
}

func (m medicalRecordImageUploadMeta) toUploadedInput(imageURL string, takenAt time.Time) uploadedMedicalRecordImageInput {
	return uploadedMedicalRecordImageInput{
		ImageURL: imageURL,
		FileName: m.fileName,
		FileSize: m.fileSize,
		MimeType: m.mimeType,
		TakenAt:  &takenAt,
	}
}
