package trimming

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestUploadTrimmingImage_SlowUploadGetsFreshQuotaReleaseDeadline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{uploadFn: func(context.Context) {
		// Explicitly complete after the release budget; unlike timing races,
		// the only oracle is whether the deferred release gets a new deadline.
		timer := time.NewTimer(supersededImageCleanupTimeout + 20*time.Millisecond)
		defer timer.Stop()
		<-timer.C
	}}
	quota := &fakeImageUploadQuotaStore{releaseFn: func(ctx context.Context) {
		require.NoError(t, ctx.Err(), "slow successful uploads must release their lease")
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assert.Greater(t, time.Until(deadline), 9*time.Second)
	}}
	h := newImageHandler(&mockService{}, uploader, quota)
	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, 1, quota.releases)
}

func TestReleaseTrimmingImageQuota_DetachesCanceledRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	releaseTrimmingImageQuota(ctx, func(ctx context.Context) {
		called = true
		require.NoError(t, ctx.Err())
		_, bounded := ctx.Deadline()
		assert.True(t, bounded)
	})
	assert.True(t, called)
}

func TestUpdateTrimming_RestoredKeyInSameRecordIsRetained(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	const key = "trimming-images/1/x.jpg"
	svc := &mockService{
		updateFn: func(context.Context, uint64, uint64, *UpdateTrimmingInput) (*model.Reservation, *TrimmingReplacedImages, error) {
			// A committed X→Y, but B restored X before A's cleanup starts.
			return trimmingDetailReservation("trimming-images/1/y.jpg", ""), &TrimmingReplacedImages{Style: key}, nil
		},
		reserveImageDeletionFn: func(_ context.Context, clinicID uint64, got string, exceptID, actorID uint64) (bool, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(42), actorID)
			assert.Equal(t, key, got)
			require.Zero(t, exceptID, "current X reference must participate in cleanup")
			return false, nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trimmings/7", strings.NewReader(`{"style_image":"trimming-images/1/y.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.UpdateTrimming(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, uploader.deletes)
}

func TestDeleteTrimming_CleansLatestLockedImageKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
			t.Fatal("handler must not snapshot images before the delete transaction")
			return nil, nil
		},
		deleteWithImagesFn: func(_ context.Context, clinicID, id uint64, actorID *uint64) (*TrimmingDeletedImages, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), id)
			require.NotNil(t, actorID)
			assert.Equal(t, uint64(42), *actorID)
			return &TrimmingDeletedImages{Style: "trimming-images/1/new.jpg"}, nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, httptest.NewRequest(http.MethodDelete, "/api/v1/trimmings/7", nil))
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.DeleteTrimming(c)
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Equal(t, []string{"trimming-images/1/new.jpg"}, uploader.deletes)
}

func TestDeleteTrimmingImage_StorageFailureKeepsTombstoneAndCanRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := newImageLifecycleFake()
	s := lifecycleService(f)
	uploader := &fakeTrimmingImageUploader{deleteErr: errors.New("storage unavailable")}
	h := newImageHandler(s, uploader, &fakeImageUploadQuotaStore{})
	const key = "trimming-images/1/style.jpg"
	request := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/trimmings/images", strings.NewReader(`{"key":"`+key+`"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c := newTrimmingHandlerContext(w, req)
		h.DeleteTrimmingImage(c)
		c.Writer.WriteHeaderNow()
		return w
	}
	assert.Equal(t, http.StatusInternalServerError, request().Code)
	assert.True(t, f.deleted[ImageKey{clinicID: 1, value: key}])
	err := f.WithTx(context.Background(), func(ctx context.Context) error {
		return s.lockImageReferencesInTx(ctx, 1, key)
	})
	require.ErrorIs(t, err, apperrors.ErrConflict)
	uploader.deleteErr = nil
	assert.Equal(t, http.StatusNoContent, request().Code)
	assert.Equal(t, []string{key, key}, uploader.deletes)
}

func TestImageLifecycle_ReferenceWritersRejectTombstonesBeforeDetailWrite(t *testing.T) {
	for _, operation := range []string{"create", "create existing", "update"} {
		t.Run(operation, func(t *testing.T) {
			f := newImageLifecycleFake()
			const key = "trimming-images/1/deleted.jpg"
			f.deleted[ImageKey{clinicID: 1, value: key}] = true
			f.createFn = func(context.Context, *model.AppointmentTrimmingDetail) error {
				t.Fatal("tombstoned key must never be created")
				return nil
			}
			f.updateFn = func(context.Context, *model.AppointmentTrimmingDetail) error {
				t.Fatal("tombstoned key must never be updated")
				return nil
			}
			f.findByAppointmentIDFn = func(context.Context, uint64, uint64) (*model.AppointmentTrimmingDetail, error) {
				if operation == "create existing" {
					return nil, apperrors.WrapNotFound("trimming detail", "7")
				}
				return &model.AppointmentTrimmingDetail{ClinicID: 1, AppointmentID: 7}, nil
			}
			reservationRepo := &mockTrimmingReservationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					return newTrimmingAuditAppointment(7, 1), nil
				},
			}
			svc := newTrimmingAuditTestService(reservationRepo, f, f, noopTrimmingAuditTxLogger{})
			var err error
			if operation == "update" {
				_, _, err = svc.Update(context.Background(), 1, 7, &UpdateTrimmingInput{StyleImage: ptrString(key), ActorID: ptrUint64(42)})
			} else {
				input := &CreateTrimmingInput{ActorID: ptrUint64(42), StyleImage: key, PetID: ptrUint64(2),
					ReservationTypeID: 9, ReservationRoute: ptrString("record_shortcut"),
					StartTime: time.Date(2026, 10, 10, 1, 0, 0, 0, time.UTC), EndTime: time.Date(2026, 10, 10, 2, 0, 0, 0, time.UTC)}
				if operation == "create existing" {
					input.AppointmentID = ptrUint64(7)
				}
				_, err = svc.Create(context.Background(), 1, input)
			}
			require.ErrorIs(t, err, apperrors.ErrConflict)
		})
	}
}
