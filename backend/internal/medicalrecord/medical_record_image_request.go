package medicalrecord

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	// SEC-CS3-O6: 拡張子と宣言 MIME の不一致（例: .html を image/png と偽装）を拒否する。
	// 保存名の拡張子は sniff 検証後の mimeType から導出するため、ここで騙されても
	// active content の拡張子は保存されない。多層防御として両者の一致を要求する。
	extMIME, ok := medicalRecordImageMIMETypeFromExt(fileExt)
	if !ok || extMIME != mimeType {
		return medicalRecordImageUploadMeta{}, apperrors.WrapInvalidInput("file extension does not match declared file type; allowed: jpeg, png, gif, pdf")
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

// medicalRecordImageExtForMIME は検証済み mimeType から保存拡張子を導出する。
// 元ファイル名の拡張子は保存名に使わない（SEC-CS3-O6: active content の拡張子温存を防ぐ）。
func medicalRecordImageExtForMIME(mimeType string) (string, bool) {
	switch mimeType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "application/pdf":
		return ".pdf", true
	default:
		return "", false
	}
}

// verifySniffedContent は先頭バイトの content-sniff が検証済み mimeType と一致するか確認する。
// 宣言 Content-Type 偽装（.html を image/png 申告）を実バイトで検出する SEC-CS3-O6 の第二層。
func (m medicalRecordImageUploadMeta) verifySniffedContent(file multipart.File) error {
	const sniffLen = 512
	head := make([]byte, sniffLen)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return apperrors.Wrap(err, "failed to read uploaded file")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return apperrors.Wrap(err, "failed to rewind uploaded file")
	}
	sniffed, _, _ := strings.Cut(http.DetectContentType(head[:n]), ";")
	if strings.TrimSpace(sniffed) != m.mimeType {
		return apperrors.WrapInvalidInput("file content does not match the declared file type")
	}
	return nil
}

func (m medicalRecordImageUploadMeta) newStoredName(now time.Time) (string, error) {
	ext, ok := medicalRecordImageExtForMIME(m.mimeType)
	if !ok {
		return "", apperrors.WrapInvalidInput("unsupported MIME type: " + m.mimeType)
	}
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", apperrors.Wrap(err, "failed to generate unique filename")
	}
	return fmt.Sprintf("%d_%s%s", now.UnixNano(), hex.EncodeToString(randomBytes), ext), nil
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
