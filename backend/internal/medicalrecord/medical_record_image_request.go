package medicalrecord

import (
	"mime/multipart"
	"time"

	"github.com/animal-ekarte/backend/internal/imageupload"
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

var medicalRecordImageMIMEByExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".pdf":  "application/pdf",
}

var medicalRecordImageExtByMIME = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"application/pdf": ".pdf",
}

var medicalRecordImageConstraints = imageupload.FileConstraints{
	MaxSize:          medicalRecordImageMaxUploadSize,
	AllowedMIMETypes: allowedMedicalRecordImageMIMETypes,
	MIMEByExt:        medicalRecordImageMIMEByExt,
	ExtByMIME:        medicalRecordImageExtByMIME,
	AllowedLabel:     "jpeg, png, gif, pdf",
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
	meta, err := medicalRecordImageConstraints.ValidateFileHeader(r.fileHeader)
	if err != nil {
		return medicalRecordImageUploadMeta{}, err
	}
	return medicalRecordImageUploadMeta{
		fileName: meta.FileName,
		fileSize: meta.FileSize,
		fileExt:  meta.FileExt,
		mimeType: meta.MIMEType,
	}, nil
}

func medicalRecordImageMIMETypeFromExt(fileExt string) (string, bool) {
	mimeType, ok := medicalRecordImageMIMEByExt[fileExt]
	return mimeType, ok
}

// verifySniffedContent は先頭バイトの content-sniff が検証済み mimeType と一致するか確認する。
// 宣言 Content-Type 偽装（.html を image/png 申告）を実バイトで検出する SEC-CS3-O6 の第二層。
func (m medicalRecordImageUploadMeta) verifySniffedContent(file multipart.File) error {
	return imageupload.VerifySniffedContent(file, m.mimeType)
}

func (m medicalRecordImageUploadMeta) newStoredName(now time.Time) (string, error) {
	return medicalRecordImageConstraints.NewStoredName(m.mimeType, now)
}

func (m medicalRecordImageUploadMeta) uploadKey(medicalRecordID uint64, storedName string) string {
	return imageupload.StoredKey(medicalRecordImageStorageKeyPrefix, medicalRecordID, storedName)
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
