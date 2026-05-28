package handler

import (
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCreateMedicalRecordImageRequest_ToServiceInput(t *testing.T) {
	takenAt := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)
	examID := uint64(10)
	staffID := uint64(20)

	input := createMedicalRecordImageRequest{
		ImageURL:     "https://example.test/image.png",
		ThumbnailURL: "https://example.test/thumb.png",
		FileName:     "image.png",
		FileSize:     1234,
		MimeType:     "image/png",
		ImageType:    string(model.MedicalImageTypeXray),
		Description:  "xray image",
		TakenAt:      &takenAt,
		ExamID:       &examID,
		StaffID:      &staffID,
		SortOrder:    2,
	}.toServiceInput()

	if input.ImageURL != "https://example.test/image.png" {
		t.Fatalf("ImageURL = %q", input.ImageURL)
	}
	if input.ImageType != model.MedicalImageTypeXray {
		t.Fatalf("ImageType = %s, want %s", input.ImageType, model.MedicalImageTypeXray)
	}
	if input.TakenAt != &takenAt {
		t.Fatalf("TakenAt pointer was not preserved")
	}
	if input.ExamID != &examID {
		t.Fatalf("ExamID pointer was not preserved")
	}
	if input.StaffID != &staffID {
		t.Fatalf("StaffID pointer was not preserved")
	}
}

func TestCreateMedicalRecordImageRequest_ToServiceInput_DefaultImageType(t *testing.T) {
	input := createMedicalRecordImageRequest{ImageURL: "https://example.test/image.png"}.toServiceInput()

	if input.ImageType != model.MedicalImageTypeOther {
		t.Fatalf("ImageType = %s, want %s", input.ImageType, model.MedicalImageTypeOther)
	}
}

func TestUploadedMedicalRecordImageInput_ToServiceInput(t *testing.T) {
	takenAt := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)

	input := uploadedMedicalRecordImageInput{
		ImageURL: "/uploads/image.png",
		FileName: "image.png",
		FileSize: 2048,
		MimeType: "image/png",
		TakenAt:  &takenAt,
	}.toServiceInput()

	if input.ImageURL != "/uploads/image.png" {
		t.Fatalf("ImageURL = %q", input.ImageURL)
	}
	if input.ImageType != model.MedicalImageTypeOther {
		t.Fatalf("ImageType = %s, want %s", input.ImageType, model.MedicalImageTypeOther)
	}
	if input.TakenAt != &takenAt {
		t.Fatalf("TakenAt pointer was not preserved")
	}
}

func TestMedicalRecordImageUploadRequest_Validate(t *testing.T) {
	header := multipart.FileHeader{
		Filename: "image.PNG",
		Header:   textproto.MIMEHeader{"Content-Type": []string{"image/png"}},
		Size:     2048,
	}

	meta, err := newMedicalRecordImageUploadRequest(&header).validate()
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}

	if meta.fileName != "image.PNG" {
		t.Fatalf("fileName = %q, want image.PNG", meta.fileName)
	}
	if meta.fileExt != ".png" {
		t.Fatalf("fileExt = %q, want .png", meta.fileExt)
	}
	if meta.mimeType != "image/png" {
		t.Fatalf("mimeType = %q, want image/png", meta.mimeType)
	}
	if meta.fileSize != 2048 {
		t.Fatalf("fileSize = %d, want 2048", meta.fileSize)
	}
}

func TestMedicalRecordImageUploadRequest_Validate_DetectsMIMETypeFromExt(t *testing.T) {
	header := multipart.FileHeader{
		Filename: "report.pdf",
		Header:   textproto.MIMEHeader{},
		Size:     1024,
	}

	meta, err := newMedicalRecordImageUploadRequest(&header).validate()
	if err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if meta.mimeType != "application/pdf" {
		t.Fatalf("mimeType = %q, want application/pdf", meta.mimeType)
	}
}

func TestMedicalRecordImageUploadRequest_Validate_RejectsUnsupportedExt(t *testing.T) {
	header := multipart.FileHeader{
		Filename: "image.bmp",
		Header:   textproto.MIMEHeader{},
		Size:     1024,
	}

	_, err := newMedicalRecordImageUploadRequest(&header).validate()
	if err == nil {
		t.Fatal("validate returned nil error")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("error = %q, want unsupported file type", err.Error())
	}
}

func TestMedicalRecordImageUploadMeta_NewStoredName(t *testing.T) {
	now := time.Date(2026, 5, 28, 9, 0, 0, 123, time.UTC)

	storedName, err := (medicalRecordImageUploadMeta{fileExt: ".png"}).newStoredName(now)
	if err != nil {
		t.Fatalf("newStoredName returned error: %v", err)
	}

	if !strings.HasPrefix(storedName, "1779958800000000123_") {
		t.Fatalf("storedName = %q, want timestamp prefix", storedName)
	}
	if !strings.HasSuffix(storedName, ".png") {
		t.Fatalf("storedName = %q, want .png suffix", storedName)
	}
}
