package lstep

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// uploadSharedFileRequest はPOST /shared-files のフォームパラメータ
// Purpose: model 定数 (inspection_result|vaccine_cert|other); empty → other (LSB-06).
type uploadSharedFileRequest struct {
	Purpose string  `form:"purpose" binding:"omitempty,max=50,oneof=inspection_result vaccine_cert other"`
	OwnerID *uint64 `form:"owner_id"`
}

type sharedFileUploadMeta struct {
	fileName    string
	contentType string
	fileType    string
	fileSize    int64
}

func newSharedFileUploadMeta(header *multipart.FileHeader) (sharedFileUploadMeta, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType, allowed := model.AllowedFileExtensions[ext]
	if !allowed {
		return sharedFileUploadMeta{}, apperrors.WrapInvalidInput(fmt.Sprintf("file type %q is not allowed", ext))
	}

	return sharedFileUploadMeta{
		fileName:    header.Filename,
		contentType: contentType,
		fileType:    model.AllowedFileTypes[ext],
		fileSize:    header.Size,
	}, nil
}

func (r uploadSharedFileRequest) toServiceInput(content io.Reader, meta sharedFileUploadMeta) (*UploadSharedFileInput, error) {
	purpose, err := validateSharedFilePurpose(r.Purpose)
	if err != nil {
		return nil, err
	}

	return &UploadSharedFileInput{
		Content:     content,
		FileName:    meta.fileName,
		ContentType: meta.contentType,
		FileType:    meta.fileType,
		FileSize:    meta.fileSize,
		Purpose:     purpose,
		OwnerID:     r.OwnerID,
	}, nil
}

// validateSharedFilePurpose enforces purpose enum + length (LSB-06). Empty → other.
// Length cap matches DDL shared_files.purpose varchar(50).
func validateSharedFilePurpose(purpose string) (string, error) {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return model.SharedFilePurposeOther, nil
	}
	if len(purpose) > 50 {
		return "", apperrors.WrapInvalidInput("purpose length must be <= 50")
	}
	switch purpose {
	case model.SharedFilePurposeInspectionResult, model.SharedFilePurposeVaccineCert, model.SharedFilePurposeOther:
		return purpose, nil
	default:
		return "", apperrors.WrapInvalidInput(fmt.Sprintf("invalid purpose: %s", purpose))
	}
}
