package medicalrecord

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

	input := (&createMedicalRecordImageRequest{
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
	}).toServiceInput()

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
	input := (&createMedicalRecordImageRequest{ImageURL: "https://example.test/image.png"}).toServiceInput()

	if input.ImageType != model.MedicalImageTypeOther {
		t.Fatalf("ImageType = %s, want %s", input.ImageType, model.MedicalImageTypeOther)
	}
}

func TestCreateMedicalRecordImageRequest_ValidateJSONCreate(t *testing.T) {
	// MRC-09
	t.Run("rejects disallowed mime", func(t *testing.T) {
		req := &createMedicalRecordImageRequest{
			ImageURL: "https://example.test/x.html",
			MimeType: "text/html",
			FileSize: 10,
		}
		if err := req.validateJSONCreate(); err == nil {
			t.Fatal("expected error for text/html")
		}
	})
	t.Run("rejects oversized file", func(t *testing.T) {
		req := &createMedicalRecordImageRequest{
			ImageURL: "https://example.test/x.png",
			MimeType: "image/png",
			FileSize: medicalRecordImageMaxUploadSize + 1,
		}
		if err := req.validateJSONCreate(); err == nil {
			t.Fatal("expected error for oversized file")
		}
	})
	t.Run("accepts allowlisted mime", func(t *testing.T) {
		req := &createMedicalRecordImageRequest{
			ImageURL: "https://example.test/x.png",
			MimeType: "image/png",
			FileSize: 100,
		}
		if err := req.validateJSONCreate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestUploadedMedicalRecordImageInput_ToServiceInput(t *testing.T) {
	takenAt := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)

	input := (&uploadedMedicalRecordImageInput{
		ImageURL: "/uploads/image.png",
		FileName: "image.png",
		FileSize: 2048,
		MimeType: "image/png",
		TakenAt:  &takenAt,
	}).toServiceInput()

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

func TestMedicalRecordImageUploadMeta_NewStoredName_DifferentExtAndTime(t *testing.T) {
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	storedName, err := (medicalRecordImageUploadMeta{fileExt: ".pdf"}).newStoredName(now)
	if err != nil {
		t.Fatalf("newStoredName returned error: %v", err)
	}
	if !strings.HasSuffix(storedName, ".pdf") {
		t.Fatalf("storedName = %q, want .pdf suffix", storedName)
	}
	// unix nano for 2000-01-01T00:00:00Z prefix
	if !strings.HasPrefix(storedName, "946684800000000000_") {
		t.Fatalf("storedName = %q, want timestamp prefix", storedName)
	}
}

func TestMedicalRecordImageUploadRequest_Validate_RejectsOversizedFile(t *testing.T) {
	header := multipart.FileHeader{
		Filename: "image.png",
		Header:   textproto.MIMEHeader{"Content-Type": []string{"image/png"}},
		Size:     medicalRecordImageMaxUploadSize + 1,
	}

	_, err := newMedicalRecordImageUploadRequest(&header).validate()
	if err == nil {
		t.Fatal("validate returned nil error")
	}
	if !strings.Contains(err.Error(), "file size exceeds limit") {
		t.Fatalf("error = %q, want file size exceeds limit", err.Error())
	}
}

func TestMedicalRecordImageUploadRequest_Validate_RejectsUnknownContentTypeWithUnsupportedExt(t *testing.T) {
	header := multipart.FileHeader{
		Filename: "notes.txt",
		Header:   textproto.MIMEHeader{"Content-Type": []string{"text/plain"}},
		Size:     512,
	}

	_, err := newMedicalRecordImageUploadRequest(&header).validate()
	if err == nil {
		t.Fatal("validate returned nil error")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("error = %q, want unsupported file type", err.Error())
	}
}

func TestMedicalRecordImageMIMETypeFromExt(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		wantMIME string
		wantOK   bool
	}{
		{"jpg", ".jpg", "image/jpeg", true},
		{"jpeg", ".jpeg", "image/jpeg", true},
		{"png", ".png", "image/png", true},
		{"gif", ".gif", "image/gif", true},
		{"pdf", ".pdf", "application/pdf", true},
		{"unsupported extension", ".bmp", "", false},
		{"empty extension", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime, ok := medicalRecordImageMIMETypeFromExt(tt.ext)
			if mime != tt.wantMIME || ok != tt.wantOK {
				t.Errorf("medicalRecordImageMIMETypeFromExt(%q) = (%q, %v), want (%q, %v)", tt.ext, mime, ok, tt.wantMIME, tt.wantOK)
			}
		})
	}
}

func TestMedicalRecordImageUploadMeta_UploadKey(t *testing.T) {
	m := medicalRecordImageUploadMeta{fileExt: ".png"}

	got := m.uploadKey(42, "storedname.png")
	want := "medical-records/42/storedname.png"
	if got != want {
		t.Fatalf("uploadKey = %q, want %q", got, want)
	}
}

func TestMedicalRecordImageUploadMeta_ToUploadedInput(t *testing.T) {
	meta := medicalRecordImageUploadMeta{
		fileName: "image.png",
		fileSize: 2048,
		fileExt:  ".png",
		mimeType: "image/png",
	}
	takenAt := time.Date(2026, 5, 28, 9, 0, 0, 0, time.UTC)

	got := meta.toUploadedInput("/uploads/image.png", takenAt)

	if got.ImageURL != "/uploads/image.png" {
		t.Errorf("ImageURL = %q, want /uploads/image.png", got.ImageURL)
	}
	if got.FileName != meta.fileName {
		t.Errorf("FileName = %q, want %q", got.FileName, meta.fileName)
	}
	if got.FileSize != meta.fileSize {
		t.Errorf("FileSize = %d, want %d", got.FileSize, meta.fileSize)
	}
	if got.MimeType != meta.mimeType {
		t.Errorf("MimeType = %q, want %q", got.MimeType, meta.mimeType)
	}
	if got.TakenAt == nil || !got.TakenAt.Equal(takenAt) {
		t.Errorf("TakenAt = %v, want %v", got.TakenAt, takenAt)
	}
}
