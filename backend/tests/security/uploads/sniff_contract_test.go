// Package uploads holds memory-only validation helpers for EMR-332 C10
// (upload validation chain). Fixtures are tiny inert magic-byte buffers only:
// no executable payloads, no oversized/decompression-bomb samples, no network,
// no DB, no filesystem persistence.
package uploads

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/imageupload"
)

// memFile is a memory-only multipart.File for the sniff contract.
type memFile struct{ *bytes.Reader }

func (memFile) Close() error { return nil }

func newMemFile(b []byte) multipart.File { return memFile{Reader: bytes.NewReader(b)} }

// Inert fixtures: magic bytes only — never a valid runnable payload.
var (
	pngHead  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
	htmlHead = []byte("<!DOCTYPE html><html><body>static inert fixture</body></html>")
)

var testConstraints = imageupload.FileConstraints{
	MaxSize:          10 * 1024 * 1024,
	AllowedMIMETypes: map[string]bool{"image/jpeg": true, "image/png": true},
	MIMEByExt:        map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png"},
	ExtByMIME:        map[string]string{"image/jpeg": ".jpg", "image/png": ".png"},
	AllowedLabel:     "jpeg, png",
}

// C10: content-sniff matches declared MIME for a real PNG header.
func TestVerifySniffedContentAcceptsPNG(t *testing.T) {
	if err := imageupload.VerifySniffedContent(newMemFile(pngHead), "image/png"); err != nil {
		t.Fatalf("png fixture rejected: %v", err)
	}
}

// C10: an HTML text payload declared as image/png is rejected by sniffing.
func TestVerifySniffedContentRejectsHTMLDisguisedAsPNG(t *testing.T) {
	err := imageupload.VerifySniffedContent(newMemFile(htmlHead), "image/png")
	if err == nil {
		t.Fatal("html fixture accepted as image/png")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Fatalf("unexpected error shape: %v", err)
	}
}

// C10: the file cursor is rewound after sniffing so the upload stream is intact.
func TestVerifySniffedContentRewinds(t *testing.T) {
	f := newMemFile(pngHead)
	if err := imageupload.VerifySniffedContent(f, "image/png"); err != nil {
		t.Fatalf("png fixture rejected: %v", err)
	}
	pos, err := f.Seek(0, 1)
	if err != nil || pos != 0 {
		t.Fatalf("cursor not rewound: pos=%d err=%v", pos, err)
	}
}

// C10: header validation rejects oversize, unknown ext, and ext/MIME mismatch.
func TestValidateFileHeaderRejections(t *testing.T) {
	mkHeader := func(name, mime string, size int64) *multipart.FileHeader {
		h := textproto.MIMEHeader{}
		if mime != "" {
			h.Set("Content-Type", mime)
		}
		return &multipart.FileHeader{Filename: name, Size: size, Header: h}
	}
	cases := []struct {
		name   string
		header *multipart.FileHeader
	}{
		{"oversize", mkHeader("a.png", "image/png", 11*1024*1024)},
		{"html-ext", mkHeader("a.html", "", 4)},
		{"ext-mime-mismatch", mkHeader("a.png", "image/jpeg", 4)},
		{"exe-ext", mkHeader("a.exe", "", 4)},
	}
	for _, tc := range cases {
		if _, err := testConstraints.ValidateFileHeader(tc.header); err == nil {
			t.Fatalf("%s: expected rejection", tc.name)
		}
	}
}

// C10: accepted header yields meta, and stored name derives ext from verified
// MIME — never from the client filename.
func TestStoredNameDerivesExtFromMIME(t *testing.T) {
	name, err := testConstraints.NewStoredName("image/png", time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("NewStoredName: %v", err)
	}
	if !strings.HasSuffix(name, ".png") {
		t.Fatalf("stored name lacks verified ext: %s", name)
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
		t.Fatalf("stored name contains traversal characters: %s", name)
	}
}
