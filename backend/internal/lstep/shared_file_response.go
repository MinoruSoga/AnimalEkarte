package lstep

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// sharedFileResponse はGETレスポンス
type sharedFileResponse struct {
	ID         uint64     `json:"id"`
	ClinicID   uint64     `json:"clinic_id"`
	OwnerID    *uint64    `json:"owner_id"`
	UploadedBy uint64     `json:"uploaded_by"`
	FileType   string     `json:"file_type"`
	FileName   string     `json:"file_name"`
	FileSize   int64      `json:"file_size"`
	Purpose    string     `json:"purpose"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// sharedFileSignedURLResponse は署名付きURLレスポンス
type sharedFileSignedURLResponse struct {
	SignedURL string `json:"signed_url"`
}

func toSharedFileResponse(r *SharedFileResponse) sharedFileResponse {
	return sharedFileResponse{
		ID:         r.ID,
		ClinicID:   r.ClinicID,
		OwnerID:    r.OwnerID,
		UploadedBy: r.UploadedBy,
		FileType:   r.FileType,
		FileName:   r.FileName,
		FileSize:   r.FileSize,
		Purpose:    r.Purpose,
		ExpiresAt:  httpapi.LocalTimePtr(r.ExpiresAt),
		CreatedAt:  httpapi.LocalTime(r.CreatedAt),
	}
}
