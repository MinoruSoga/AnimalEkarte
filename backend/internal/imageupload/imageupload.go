// Package imageupload holds the shared image-upload primitives consumed by
// feature packages (medicalrecord, trimming): multipart validation, content
// sniffing, stored-name generation, object-key layout, and the clinic/staff
// upload-quota contract. Feature packages keep their own prefix, permission
// wiring, and persistence; the byte-level policy lives here once.
package imageupload

import (
	"context"
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

// Quota error sentinels. Quota-store implementations wrap these so consumers
// classify Acquire failures with errors.Is — never by matching message text.
// Unclassified errors map to fail-closed 500.
var (
	ErrQuotaConcurrency = errors.New("too many concurrent image uploads")
	ErrQuotaRateLimit   = errors.New("image upload rate limit exceeded")
	ErrQuotaByteBudget  = errors.New("image upload byte budget exceeded")
	ErrQuotaUnavailable = errors.New("upload quota unavailable")
)

// QuotaStore reserves concurrency/rate/bytes before heavy multipart work.
// release must be called when the request finishes (success or failure).
type QuotaStore interface {
	Acquire(ctx context.Context, clinicID, staffID uint64, declaredBytes int64) (release func(context.Context), err error)
}

// FileConstraints describes the accepted image set for one consumer.
// MIMEByExt maps lowercase extensions (with dot) to MIME types; ExtByMIME is
// its inverse used to derive the stored extension. AllowedLabel renders the
// human-readable accepted list in error messages (e.g. "jpeg, png, gif").
type FileConstraints struct {
	MaxSize          int64
	AllowedMIMETypes map[string]bool
	MIMEByExt        map[string]string
	ExtByMIME        map[string]string
	AllowedLabel     string
}

// FileMeta carries the validated facts about an uploaded file.
type FileMeta struct {
	FileName string
	FileSize int64
	FileExt  string
	MIMEType string
}

// ValidateFileHeader checks declared size, extension, and Content-Type against
// the constraints. SEC-CS3-O6: 拡張子と宣言 MIME の不一致（.html を image/png と
// 偽装等）を拒否する。保存名の拡張子は sniff 検証後の MIME から導出するため、
// ここで騙されても active content の拡張子は保存されない。
func (c FileConstraints) ValidateFileHeader(fileHeader *multipart.FileHeader) (FileMeta, error) {
	if fileHeader.Size > c.MaxSize {
		return FileMeta{}, apperrors.WrapInvalidInput(fmt.Sprintf(
			"file size exceeds limit of %dMB",
			c.MaxSize/1024/1024,
		))
	}

	fileExt := strings.ToLower(filepath.Ext(fileHeader.Filename))
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" || !c.AllowedMIMETypes[mimeType] {
		detectedType, ok := c.MIMEByExt[fileExt]
		if !ok {
			return FileMeta{}, apperrors.WrapInvalidInput("unsupported file type; allowed: " + c.AllowedLabel)
		}
		mimeType = detectedType
	}
	if !c.AllowedMIMETypes[mimeType] {
		return FileMeta{}, apperrors.WrapInvalidInput("unsupported MIME type: " + mimeType)
	}
	extMIME, ok := c.MIMEByExt[fileExt]
	if !ok || extMIME != mimeType {
		return FileMeta{}, apperrors.WrapInvalidInput("file extension does not match declared file type; allowed: " + c.AllowedLabel)
	}

	return FileMeta{
		FileName: fileHeader.Filename,
		FileSize: fileHeader.Size,
		FileExt:  fileExt,
		MIMEType: mimeType,
	}, nil
}

// VerifySniffedContent は先頭バイトの content-sniff が検証済み mimeType と一致するか確認する。
// 宣言 Content-Type 偽装を実バイトで検出する第二層。
func VerifySniffedContent(file multipart.File, mimeType string) error {
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
	if strings.TrimSpace(sniffed) != mimeType {
		return apperrors.WrapInvalidInput("file content does not match the declared file type")
	}
	return nil
}

// NewStoredName builds "<unixnano>_<hex16><ext>" where ext derives from the
// verified MIME type — never from the client-supplied filename (SEC-CS3-O6).
func (c FileConstraints) NewStoredName(mimeType string, now time.Time) (string, error) {
	ext, ok := c.ExtByMIME[mimeType]
	if !ok {
		return "", apperrors.WrapInvalidInput("unsupported MIME type: " + mimeType)
	}
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", apperrors.Wrap(err, "failed to generate unique filename")
	}
	return fmt.Sprintf("%d_%s%s", now.UnixNano(), hex.EncodeToString(randomBytes), ext), nil
}

// StoredKey builds the clinic/record-scoped object key "<prefix><ownerID>/<name>".
// クライアント提供のファイル名は使わない（サーバ生成名のみ）。
func StoredKey(prefix string, ownerID uint64, storedName string) string {
	return fmt.Sprintf("%s%d/%s", prefix, ownerID, storedName)
}
