package medicalrecord

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	medicalRecordImageSignedURLTTL     = 15 * time.Minute
	medicalRecordImageStorageKeyPrefix = "medical-records/"
	medicalRecordImageLocalURLPrefix   = "/uploads/"
)

var errMedicalRecordImageStorageKeyUnresolved = errors.New("unable to resolve medical record image storage key")

// medicalRecordImageResponse は診療画像の API レスポンス
type medicalRecordImageResponse struct {
	ID              uint64                `json:"id"`
	MedicalRecordID uint64                `json:"medical_record_id"`
	ImageURL        string                `json:"image_url"`
	ThumbnailURL    string                `json:"thumbnail_url"`
	FileName        string                `json:"file_name"`
	FileSize        int64                 `json:"file_size"`
	MimeType        string                `json:"mime_type"`
	ImageType       string                `json:"image_type"`
	Description     string                `json:"description"`
	TakenAt         *time.Time            `json:"taken_at,omitempty"`
	ExamID          *uint64               `json:"exam_id,omitempty"`
	StaffID         *uint64               `json:"staff_id,omitempty"`
	SortOrder       int                   `json:"sort_order"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Staff           *StaffSummaryResponse `json:"staff,omitempty"`
}

// toMedicalRecordImageResponse は model.MedicalRecordImage を medicalRecordImageResponse に変換する
func toMedicalRecordImageResponse(img *model.MedicalRecordImage) medicalRecordImageResponse {
	r := medicalRecordImageResponse{
		ID:              img.ID,
		MedicalRecordID: img.MedicalRecordID,
		ImageURL:        img.ImageURL,
		ThumbnailURL:    img.ThumbnailURL,
		FileName:        img.FileName,
		FileSize:        img.FileSize,
		MimeType:        img.MimeType,
		ImageType:       string(img.ImageType),
		Description:     img.Description,
		TakenAt:         httpapi.LocalTimePtr(img.TakenAt),
		ExamID:          img.ExamID,
		StaffID:         img.StaffID,
		SortOrder:       img.SortOrder,
		CreatedAt:       httpapi.LocalTime(img.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(img.UpdatedAt),
	}
	if img.Staff != nil {
		r.Staff = toStaffSummary(img.Staff)
	}
	return r
}

func medicalRecordImageHasHTTPScheme(stored string) bool {
	lower := strings.ToLower(stored)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func medicalRecordImageObjectKeyHasName(key string) bool {
	if !strings.HasPrefix(key, medicalRecordImageStorageKeyPrefix) {
		return key != ""
	}
	return strings.TrimPrefix(key, medicalRecordImageStorageKeyPrefix) != ""
}

// resolveMedicalRecordImageStorageKey extracts a storage object key from a stored image_url.
// isStorageObject=false (and err=nil) means the value is not a storage object and should be echoed.
// err!=nil is fail-closed: the stored value looks like ours but a key cannot be extracted.
func resolveMedicalRecordImageStorageKey(stored string) (key string, isStorageObject bool, err error) {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return "", false, nil
	}

	if medicalRecordImageHasHTTPScheme(stored) {
		parsed, parseErr := url.Parse(stored)
		if parseErr != nil || parsed.Host == "" {
			return "", false, errMedicalRecordImageStorageKeyUnresolved
		}
		path := parsed.Path
		extracted, ok := medicalRecordImageKeyFromURLPath(path)
		if !ok {
			return "", false, nil
		}
		if !medicalRecordImageObjectKeyHasName(extracted) {
			return "", false, errMedicalRecordImageStorageKeyUnresolved
		}
		return extracted, true, nil
	}

	if strings.HasPrefix(stored, medicalRecordImageStorageKeyPrefix) {
		if !medicalRecordImageObjectKeyHasName(stored) {
			return "", false, errMedicalRecordImageStorageKeyUnresolved
		}
		return stored, true, nil
	}

	if strings.HasPrefix(stored, medicalRecordImageLocalURLPrefix) {
		rest := strings.TrimPrefix(stored, medicalRecordImageLocalURLPrefix)
		if rest == "" {
			return "", false, errMedicalRecordImageStorageKeyUnresolved
		}
		if !strings.HasPrefix(rest, medicalRecordImageStorageKeyPrefix) {
			return "", false, nil
		}
		if !medicalRecordImageObjectKeyHasName(rest) {
			return "", false, errMedicalRecordImageStorageKeyUnresolved
		}
		return rest, true, nil
	}

	return "", false, nil
}

func medicalRecordImageKeyFromURLPath(path string) (string, bool) {
	marker := "/" + medicalRecordImageStorageKeyPrefix
	idx := strings.Index(path, marker)
	if idx < 0 {
		return "", false
	}
	return path[idx+1:], true
}

func medicalRecordImageKeyBelongsToRecord(key string, medicalRecordID uint64) bool {
	rest, ok := strings.CutPrefix(key, medicalRecordImageStorageKeyPrefix)
	if !ok {
		return false
	}
	idSeg, name, ok := strings.Cut(rest, "/")
	if !ok || name == "" || strings.Contains(name, "..") {
		return false
	}
	parsed, err := strconv.ParseUint(idSeg, 10, 64)
	return err == nil && parsed == medicalRecordID
}
