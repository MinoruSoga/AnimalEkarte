package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// uploadSharedFileRequest はPOST /shared-files のフォームパラメータ
type uploadSharedFileRequest struct {
	Purpose string  `form:"purpose"`
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

func (r uploadSharedFileRequest) toServiceInput(content io.Reader, meta sharedFileUploadMeta) *service.UploadSharedFileInput {
	purpose := r.Purpose
	if purpose == "" {
		purpose = model.SharedFilePurposeOther
	}

	return &service.UploadSharedFileInput{
		Content:     content,
		FileName:    meta.fileName,
		ContentType: meta.contentType,
		FileType:    meta.fileType,
		FileSize:    meta.fileSize,
		Purpose:     purpose,
		OwnerID:     r.OwnerID,
	}
}
