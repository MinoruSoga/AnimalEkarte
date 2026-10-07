package trimming

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
)

const (
	trimmingImageMaxUploadSize  = 10 * 1024 * 1024 // 10MB
	trimmingImageMaxRequestSize = trimmingImageMaxUploadSize + 1024*1024
)

// allowedTrimmingImageMIMETypes はトリミング画像（仕上がり・スタイル写真）として
// 受理する画像型。PDF は画像フィールドではないため含めない（FE は accept="image/*"）。
var allowedTrimmingImageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

type trimmingImageUploadRequest struct {
	fileHeader *multipart.FileHeader
}

func newTrimmingImageUploadRequest(fileHeader *multipart.FileHeader) trimmingImageUploadRequest {
	return trimmingImageUploadRequest{fileHeader: fileHeader}
}

type trimmingImageUploadMeta struct {
	fileSize int64
	mimeType string
}

func (r trimmingImageUploadRequest) validate() (trimmingImageUploadMeta, error) {
	if r.fileHeader.Size > trimmingImageMaxUploadSize {
		return trimmingImageUploadMeta{}, apperrors.WrapInvalidInput(fmt.Sprintf(
			"file size exceeds limit of %dMB",
			trimmingImageMaxUploadSize/1024/1024,
		))
	}

	fileExt := strings.ToLower(filepath.Ext(r.fileHeader.Filename))
	mimeType := r.fileHeader.Header.Get("Content-Type")
	if mimeType == "" || !allowedTrimmingImageMIMETypes[mimeType] {
		detectedType, ok := trimmingImageMIMETypeFromExt(fileExt)
		if !ok {
			return trimmingImageUploadMeta{}, apperrors.WrapInvalidInput("unsupported file type; allowed: jpeg, png, gif")
		}
		mimeType = detectedType
	}
	if !allowedTrimmingImageMIMETypes[mimeType] {
		return trimmingImageUploadMeta{}, apperrors.WrapInvalidInput("unsupported MIME type: " + mimeType)
	}
	// SEC-CS3-O6 相当: 拡張子と宣言 MIME の不一致（.html を image/png と偽装等）を拒否する。
	// 保存名の拡張子は sniff 検証後の mimeType から導出するため、ここで騙されても
	// active content の拡張子は保存されない。多層防御として両者の一致を要求する。
	extMIME, ok := trimmingImageMIMETypeFromExt(fileExt)
	if !ok || extMIME != mimeType {
		return trimmingImageUploadMeta{}, apperrors.WrapInvalidInput("file extension does not match declared file type; allowed: jpeg, png, gif")
	}

	return trimmingImageUploadMeta{
		fileSize: r.fileHeader.Size,
		mimeType: mimeType,
	}, nil
}

func trimmingImageMIMETypeFromExt(fileExt string) (string, bool) {
	switch fileExt {
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".png":
		return "image/png", true
	case ".gif":
		return "image/gif", true
	default:
		return "", false
	}
}

// trimmingImageExtForMIME は検証済み mimeType から保存拡張子を導出する。
// 元ファイル名の拡張子は保存名に使わない（active content の拡張子温存を防ぐ）。
func trimmingImageExtForMIME(mimeType string) (string, bool) {
	switch mimeType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}

// verifySniffedContent は先頭バイトの content-sniff が検証済み mimeType と一致するか確認する。
// 宣言 Content-Type 偽装を実バイトで検出する第二層。
func (m trimmingImageUploadMeta) verifySniffedContent(file multipart.File) error {
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

func (m trimmingImageUploadMeta) newStoredName(now time.Time) (string, error) {
	ext, ok := trimmingImageExtForMIME(m.mimeType)
	if !ok {
		return "", apperrors.WrapInvalidInput("unsupported MIME type: " + m.mimeType)
	}
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", apperrors.Wrap(err, "failed to generate unique filename")
	}
	return fmt.Sprintf("%d_%s%s", now.UnixNano(), hex.EncodeToString(randomBytes), ext), nil
}

// uploadKey はクリニックスコープのオブジェクトキーを返す。
// クライアント提供のファイル名は使わない（サーバ生成名のみ）。
func (m trimmingImageUploadMeta) uploadKey(clinicID uint64, storedName string) string {
	return fmt.Sprintf("%s%d/%s", trimmingImageStorageKeyPrefix, clinicID, storedName)
}
