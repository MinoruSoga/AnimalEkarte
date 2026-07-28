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

	input := (&uploadSharedFileRequest{
		Purpose: "prescription",
		OwnerID: &ownerID,
	}).toServiceInput(content, sharedFileUploadMeta{
		fileName:    "file.pdf",
		contentType: "application/pdf",
		fileType:    "pdf",
		fileSize:    4,
	})

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
	if input.Purpose != "prescription" {
		t.Fatalf("Purpose = %q, want prescription", input.Purpose)
	}
	if input.OwnerID != &ownerID {
		t.Fatalf("OwnerID pointer was not preserved")
	}
}

func TestUploadSharedFileRequest_ToServiceInput_DefaultPurpose(t *testing.T) {
	input := (&uploadSharedFileRequest{}).toServiceInput(strings.NewReader("file"), sharedFileUploadMeta{
		fileName:    "file.pdf",
		contentType: "application/pdf",
		fileType:    "pdf",
		fileSize:    4,
	})

	if input.Purpose != model.SharedFilePurposeOther {
		t.Fatalf("Purpose = %q, want %q", input.Purpose, model.SharedFilePurposeOther)
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
