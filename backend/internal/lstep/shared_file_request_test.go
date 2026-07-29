package lstep

import (
	"mime/multipart"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestUploadSharedFileRequest_ToServiceInput(t *testing.T) {
	ownerID := uint64(10)
	content := strings.NewReader("file")

	input, err := (&uploadSharedFileRequest{
		Purpose: model.SharedFilePurposeVaccineCert,
		OwnerID: &ownerID,
	}).toServiceInput(content, sharedFileUploadMeta{
		fileName:    "file.pdf",
		contentType: "application/pdf",
		fileType:    "pdf",
		fileSize:    4,
	})
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Content != content {
		t.Fatalf("Content reader was not preserved")
	}
	if input.FileName != "file.pdf" {
		t.Fatalf("FileName = %q, want file.pdf", input.FileName)
	}
	if input.ContentType != "application/pdf" {
		t.Fatalf("ContentType = %q, want application/pdf", input.ContentType)
	}
	if input.FileType != "pdf" {
		t.Fatalf("FileType = %q, want pdf", input.FileType)
	}
	if input.FileSize != 4 {
		t.Fatalf("FileSize = %d, want 4", input.FileSize)
	}
	if input.Purpose != model.SharedFilePurposeVaccineCert {
		t.Fatalf("Purpose = %q, want %q", input.Purpose, model.SharedFilePurposeVaccineCert)
	}
	if input.OwnerID != &ownerID {
		t.Fatalf("OwnerID pointer was not preserved")
	}
}

func TestUploadSharedFileRequest_ToServiceInput_DefaultPurpose(t *testing.T) {
	input, err := (&uploadSharedFileRequest{}).toServiceInput(strings.NewReader("file"), sharedFileUploadMeta{
		fileName:    "file.pdf",
		contentType: "application/pdf",
		fileType:    "pdf",
		fileSize:    4,
	})
	if err != nil {
		t.Fatalf("toServiceInput returned error: %v", err)
	}

	if input.Purpose != model.SharedFilePurposeOther {
		t.Fatalf("Purpose = %q, want %q", input.Purpose, model.SharedFilePurposeOther)
	}
}

func TestUploadSharedFileRequest_ToServiceInput_RejectsInvalidPurpose(t *testing.T) {
	_, err := (&uploadSharedFileRequest{Purpose: "prescription"}).toServiceInput(
		strings.NewReader("file"),
		sharedFileUploadMeta{fileName: "file.pdf", contentType: "application/pdf", fileType: "pdf", fileSize: 4},
	)
	if err == nil {
		t.Fatal("expected invalid purpose error")
	}
	if !strings.Contains(err.Error(), "invalid purpose") {
		t.Fatalf("error = %q, want invalid purpose", err.Error())
	}
}

func TestUploadSharedFileRequest_ToServiceInput_RejectsPurposeLongerThan50(t *testing.T) {
	long := strings.Repeat("a", 51)
	_, err := (&uploadSharedFileRequest{Purpose: long}).toServiceInput(
		strings.NewReader("file"),
		sharedFileUploadMeta{fileName: "file.pdf", contentType: "application/pdf", fileType: "pdf", fileSize: 4},
	)
	if err == nil {
		t.Fatal("expected purpose length error")
	}
	if !strings.Contains(err.Error(), "purpose length") {
		t.Fatalf("error = %q, want purpose length", err.Error())
	}
}

func TestValidateSharedFilePurpose_MatchesModelConstants(t *testing.T) {
	for _, purpose := range []string{
		model.SharedFilePurposeInspectionResult,
		model.SharedFilePurposeVaccineCert,
		model.SharedFilePurposeOther,
	} {
		got, err := validateSharedFilePurpose(purpose)
		if err != nil {
			t.Fatalf("purpose %q: %v", purpose, err)
		}
		if got != purpose {
			t.Fatalf("purpose %q: got %q", purpose, got)
		}
	}
}

func TestNewSharedFileUploadMeta(t *testing.T) {
	meta, err := newSharedFileUploadMeta(&multipart.FileHeader{
		Filename: "file.PDF",
		Size:     4,
	})
	if err != nil {
		t.Fatalf("newSharedFileUploadMeta returned error: %v", err)
	}

	if meta.fileName != "file.PDF" {
		t.Fatalf("fileName = %q, want file.PDF", meta.fileName)
	}
	if meta.contentType != "application/pdf" {
		t.Fatalf("contentType = %q, want application/pdf", meta.contentType)
	}
	if meta.fileType != model.SharedFileTypePDF {
		t.Fatalf("fileType = %q, want %q", meta.fileType, model.SharedFileTypePDF)
	}
	if meta.fileSize != 4 {
		t.Fatalf("fileSize = %d, want 4", meta.fileSize)
	}
}

func TestNewSharedFileUploadMeta_RejectsUnsupportedExt(t *testing.T) {
	_, err := newSharedFileUploadMeta(&multipart.FileHeader{Filename: "file.exe"})
	if err == nil {
		t.Fatal("newSharedFileUploadMeta returned nil error")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("error = %q, want not allowed", err.Error())
	}
}
