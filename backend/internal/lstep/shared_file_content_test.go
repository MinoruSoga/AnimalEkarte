package lstep

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sharedFileContentCase struct {
	name        string
	fileName    string
	content     []byte
	wantMIME    string
	wantStatus  int
	wantUploads int
}

// These fixtures stay in memory; no storage, database or delivery client is used.
func TestUploadSharedFile_ContentValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var pngBytes, jpegBytes bytes.Buffer
	pixel := image.NewRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, png.Encode(&pngBytes, pixel))
	require.NoError(t, jpeg.Encode(&jpegBytes, pixel, nil))
	pdf := []byte("%PDF-1.4\n% SECCHK20261009C10\n%%EOF\n")
	marker := []byte("SECCHK20261009C10CONTENT")
	tests := []sharedFileContentCase{
		{"PNG", "marker.png", pngBytes.Bytes(), "image/png", http.StatusCreated, 1},
		{"JPEG jpg", "marker.jpg", jpegBytes.Bytes(), "image/jpeg", http.StatusCreated, 1},
		{"JPEG jpeg", "marker.jpeg", jpegBytes.Bytes(), "image/jpeg", http.StatusCreated, 1},
		{"uppercase extension", "marker.PNG", pngBytes.Bytes(), "image/png", http.StatusCreated, 1},
		{"PDF signature", "marker.pdf", pdf, "application/pdf", http.StatusCreated, 1},
		{"text as PNG", "marker.png", marker, "", http.StatusBadRequest, 0},
		{"text as JPG", "marker.jpg", marker, "", http.StatusBadRequest, 0},
		{"text as JPEG", "marker.jpeg", marker, "", http.StatusBadRequest, 0},
		{"text as PDF", "marker.pdf", marker, "", http.StatusBadRequest, 0},
		{"JPEG as PNG", "marker.png", jpegBytes.Bytes(), "", http.StatusBadRequest, 0},
		{"PNG as PDF", "marker.pdf", pngBytes.Bytes(), "", http.StatusBadRequest, 0},
		{"empty PNG", "marker.png", nil, "", http.StatusBadRequest, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runSharedFileContentCase(t, tt, marker)
		})
	}
}

func runSharedFileContentCase(t *testing.T, tt sharedFileContentCase, marker []byte) {
	t.Helper()
	uploads := 0
	svc := &mockSharedFileService{
		uploadFn: func(_ context.Context, clinicID, staffID uint64, input *UploadSharedFileInput) (*SharedFileResponse, error) {
			uploads++
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(1), staffID)
			assert.Equal(t, tt.wantMIME, input.ContentType)
			assert.Equal(t, int64(len(tt.content)), input.FileSize)
			content, err := io.ReadAll(input.Content)
			require.NoError(t, err)
			// Exact equality also verifies the sniffed prefix was rewound.
			assert.Equal(t, tt.content, content)
			return &SharedFileResponse{ID: 5, ClinicID: clinicID, FileName: tt.fileName}, nil
		},
	}
	body, contentType := buildSharedFileMultipart(t, true, tt.fileName, tt.content, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/shared-files", body)
	c.Request.Header.Set("Content-Type", contentType)
	setClinicID(c)
	setStaffID(c)
	newHandlerWithSharedFileSvc(svc).UploadSharedFile(c)
	assert.Equal(t, tt.wantStatus, w.Code)
	assert.Equal(t, tt.wantUploads, uploads)
	if tt.wantUploads == 0 {
		assert.Empty(t, w.Header().Get("Location"))
		assert.NotContains(t, w.Body.String(), string(marker))
	}
}
