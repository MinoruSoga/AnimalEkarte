package trimming

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteTrimmingImage_PathAliasCannotDeleteReferencedObject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, alias := range []string{
		"trimming-images/1/./abc.jpg",
		"trimming-images/1//abc.jpg",
		"trimming-images/1/abc.jpg/.",
		"/uploads/trimming-images/1/./abc.jpg",
	} {
		t.Run(alias, func(t *testing.T) {
			uploader := &fakeTrimmingImageUploader{}
			reservationCalled := false
			h := newImageHandler(&mockService{reserveImageDeletionFn: func(context.Context, uint64, string, uint64, uint64) (bool, error) {
				reservationCalled = true
				return true, nil
			}}, uploader, nil)
			body, err := json.Marshal(trimmingImageDeleteRequest{Key: alias})
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/trimmings/images", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.DeleteTrimmingImage(newTrimmingHandlerContext(w, req))
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.False(t, reservationCalled, "an alias must not acquire an independent tombstone")
			assert.Empty(t, uploader.deletes, "local filepath.Clean would delete the canonical object")
		})
	}
}

func TestImageLifecycle_PathAliasesCannotReserveDeletion(t *testing.T) {
	f := newImageLifecycleFake()
	s := lifecycleService(f)
	key := ImageKey{clinicID: 1, value: "trimming-images/1/abc.jpg"}
	f.refs[key] = map[uint64]bool{7: true}
	for _, alias := range []string{"trimming-images/1/./abc.jpg", "trimming-images/1//abc.jpg"} {
		reserved, err := s.ReserveImageDeletion(context.Background(), 1, alias, 0, 42)
		require.Error(t, err)
		assert.False(t, reserved)
		assert.Empty(t, f.deleted)
	}
	assert.Empty(t, f.attempts, "aliases must be rejected before acquiring raw-key locks")
}

func TestNormalizeTrimmingImages_PathAliasRejectedInBothFields(t *testing.T) {
	for _, alias := range []string{"trimming-images/1/./abc.jpg", "/uploads/trimming-images/1//abc.jpg?v=1"} {
		for _, completed := range []bool{false, true} {
			create := &CreateTrimmingInput{}
			update := &UpdateTrimmingInput{}
			value := alias
			if completed {
				create.CompletedImage = value
				update.CompletedImage = &value
			} else {
				create.StyleImage = value
				update.StyleImage = &value
			}
			require.Error(t, normalizeCreateTrimmingImages(1, create))
			require.Error(t, normalizeUpdateTrimmingImages(1, update))
		}
	}
}
