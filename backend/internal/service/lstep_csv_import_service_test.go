package service

import (
	"bytes"
	"context"
	"strings"
	"testing"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// TestImportFriendAttributesCSV_FileTooLarge: 50MB+1 バイトの reader で WrapInvalidInput を返すことを確認。
// csvImportRepo.Create 到達前 (line 69-71) に return するため nil repo で safe。
func TestImportFriendAttributesCSV_FileTooLarge(t *testing.T) {
	big := bytes.Repeat([]byte("x"), 50*1024*1024+1)
	svc := &lstepCsvImportService{
		db:            nil,
		csvImportRepo: nil,
		snapshotRepo:  nil,
		ownerRepo:     nil,
	}
	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "large.csv", bytes.NewReader(big), nil)
	if err == nil {
		t.Fatal("expected error for 50MB+ file, got nil")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got: %v", err)
	}
}

// TestImportFriendAttributesCSV_EmptyFile: 0 バイトの reader で WrapInvalidInput を返すことを確認。
// CSV パース後の空チェック (line 86-88) で return するため nil repo で safe。
func TestImportFriendAttributesCSV_EmptyFile(t *testing.T) {
	svc := &lstepCsvImportService{
		db:            nil,
		csvImportRepo: nil,
		snapshotRepo:  nil,
		ownerRepo:     nil,
	}
	_, err := svc.ImportFriendAttributesCSV(context.Background(), 100, "empty.csv", strings.NewReader(""), nil)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
	if !apperrors.IsInvalidInput(err) {
		t.Errorf("expected InvalidInput error, got: %v", err)
	}
}
