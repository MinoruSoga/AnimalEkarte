package support

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seekableFile は multipart.File 相当の検証用ラッパー
type seekableFile struct {
	*bytes.Reader
}

func (f *seekableFile) Close() error { return nil }
func (f *seekableFile) ReadAt(p []byte, off int64) (int, error) {
	return f.Reader.ReadAt(p, off)
}
func newSeekableFile(b []byte) multipart.File {
	return &seekableFile{Reader: bytes.NewReader(b)}
}

var pngBytes = append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, bytes.Repeat([]byte{0}, 64)...)

func TestValidateScreenshot(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		wantErr  bool
		wantExt  string
		wantMIME string
	}{
		{name: "accepts png", content: pngBytes, wantExt: ".png", wantMIME: "image/png"},
		{name: "accepts jpeg", content: append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, bytes.Repeat([]byte{0}, 64)...), wantExt: ".jpg", wantMIME: "image/jpeg"},
		{name: "rejects text", content: []byte("this is not an image"), wantErr: true},
		{name: "rejects empty", content: []byte{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := newSeekableFile(tt.content)
			meta, err := validateScreenshot(file)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantExt, meta.extension)
			assert.Equal(t, tt.wantMIME, meta.mimeType)

			// Seek 位置が先頭に戻っていること（アップロードで全体を読める）
			all, err := io.ReadAll(file)
			require.NoError(t, err)
			assert.Equal(t, tt.content, all)
		})
	}
}

func TestUploadKey(t *testing.T) {
	meta := &screenshotUpload{mimeType: "image/png", extension: ".png"}
	key := meta.uploadKey(7, time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC))
	assert.True(t, strings.HasPrefix(key, "support-bug-reports/clinic-7/"))
	assert.True(t, strings.HasSuffix(key, ".png"))
}
