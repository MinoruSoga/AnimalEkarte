package medicalrecord

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestMemoryMedicalRecordImageUploadQuotaStore_StaffConcurrency(t *testing.T) {
	store := newMemoryMedicalRecordImageUploadQuotaStore()
	ctx := context.Background()
	const clinicID, staffID uint64 = 1, 10

	var releases []func(context.Context)
	for i := 0; i < medicalRecordImageUploadStaffMaxConcurrent; i++ {
		release, err := store.Acquire(ctx, clinicID, staffID, 1024)
		require.NoError(t, err, "slot %d should acquire", i)
		releases = append(releases, release)
	}

	_, err := store.Acquire(ctx, clinicID, staffID, 1024)
	require.Error(t, err)
	assert.ErrorIs(t, err, errMedicalRecordImageUploadConcurrency)

	// Different staff under clinic concurrency budget still succeeds.
	releaseOther, err := store.Acquire(ctx, clinicID, staffID+1, 1024)
	require.NoError(t, err)
	releaseOther(ctx)

	releases[0](ctx)
	releaseAfter, err := store.Acquire(ctx, clinicID, staffID, 1024)
	require.NoError(t, err, "release frees a staff concurrency slot")
	releaseAfter(ctx)

	for _, release := range releases[1:] {
		release(ctx)
	}
}

func TestMemoryMedicalRecordImageUploadQuotaStore_StaffRateLimit(t *testing.T) {
	store := newMemoryMedicalRecordImageUploadQuotaStore()
	ctx := context.Background()
	const clinicID, staffID uint64 = 2, 20

	for i := 0; i < medicalRecordImageUploadStaffRatePerMinute; i++ {
		release, err := store.Acquire(ctx, clinicID, staffID, 1)
		require.NoError(t, err, "rate slot %d", i)
		release(ctx) // released still counts toward rate window
	}

	_, err := store.Acquire(ctx, clinicID, staffID, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, errMedicalRecordImageUploadRateLimit)
}

func TestMemoryMedicalRecordImageUploadQuotaStore_StaffByteBudget(t *testing.T) {
	store := newMemoryMedicalRecordImageUploadQuotaStore()
	ctx := context.Background()
	const clinicID, staffID uint64 = 3, 30

	// Fill budget with two acquires that sum exactly to the staff byte budget.
	half := int64(medicalRecordImageUploadStaffByteBudget / 2)
	release1, err := store.Acquire(ctx, clinicID, staffID, half)
	require.NoError(t, err)
	release1(ctx)
	release2, err := store.Acquire(ctx, clinicID, staffID, half)
	require.NoError(t, err)
	release2(ctx)

	_, err = store.Acquire(ctx, clinicID, staffID, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, errMedicalRecordImageUploadByteBudget)
}

func TestMemoryMedicalRecordImageUploadQuotaStore_SharedAuthoritativeState(t *testing.T) {
	// Two "handlers" share one store — simulates multi-process authoritative state.
	store := newMemoryMedicalRecordImageUploadQuotaStore()
	ctx := context.Background()
	const clinicID, staffID uint64 = 4, 40

	var held []func(context.Context)
	for i := 0; i < medicalRecordImageUploadStaffMaxConcurrent; i++ {
		release, err := store.Acquire(ctx, clinicID, staffID, 10)
		require.NoError(t, err)
		held = append(held, release)
	}

	_, errA := store.Acquire(ctx, clinicID, staffID, 10)
	_, errB := store.Acquire(ctx, clinicID, staffID, 10)
	require.ErrorIs(t, errA, errMedicalRecordImageUploadConcurrency)
	require.ErrorIs(t, errB, errMedicalRecordImageUploadConcurrency)

	for _, release := range held {
		release(ctx)
	}
}

func TestMemoryMedicalRecordImageUploadQuotaStore_ClinicConcurrency(t *testing.T) {
	store := newMemoryMedicalRecordImageUploadQuotaStore()
	ctx := context.Background()
	const clinicID uint64 = 5

	var held []func(context.Context)
	for i := 0; i < medicalRecordImageUploadClinicMaxConcurrent; i++ {
		// Spread across staff so staff concurrency does not trip first.
		staffID := uint64(100 + i)
		release, err := store.Acquire(ctx, clinicID, staffID, 1)
		require.NoError(t, err, "clinic slot %d staff %d", i, staffID)
		held = append(held, release)
	}

	_, err := store.Acquire(ctx, clinicID, 999, 1)
	require.ErrorIs(t, err, errMedicalRecordImageUploadConcurrency)

	for _, release := range held {
		release(ctx)
	}
}

func TestUploadMedicalRecordImage_QuotaRejectsExcessConcurrentBeforeBodyRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newMemoryMedicalRecordImageUploadQuotaStore()

	// Hold staff concurrency slots open so the next upload is rejected.
	ctx := context.Background()
	var held []func(context.Context)
	for i := 0; i < medicalRecordImageUploadStaffMaxConcurrent; i++ {
		release, err := store.Acquire(ctx, 1, 1, 1)
		require.NoError(t, err)
		held = append(held, release)
	}
	t.Cleanup(func() {
		for _, release := range held {
			release(ctx)
		}
	})

	h := NewMedicalRecordImageHandler(
		&mockMedicalRecordImageService{
			createFn: func(context.Context, uint64, uint64, *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				t.Fatal("service must not run when quota rejects")
				return nil, nil
			},
		},
		&mockMedicalRecordService{
			getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
			},
		},
		&mockMedicalRecordImageUploader{
			uploadFn: func(context.Context, string, io.Reader, string) (string, error) {
				t.Fatal("uploader must not run when quota rejects")
				return "", nil
			},
		},
		store,
	)

	body, contentType := buildImageMultipart(t, "photo.png", "image/png", []byte("fake-image-bytes"))
	counting := &countingUploadRequestBody{Reader: body}
	req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", counting)
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(counting.Len())

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)

	h.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "too many concurrent medical record image uploads")
	assert.Zero(t, counting.bytesRead, "quota rejection must happen before multipart body read")
}

func TestUploadMedicalRecordImage_NilQuotaFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewMedicalRecordImageHandler(
		&mockMedicalRecordImageService{},
		&mockMedicalRecordService{
			getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
			},
		},
		&mockMedicalRecordImageUploader{},
		nil,
	)

	body, contentType := buildImageMultipart(t, "photo.png", "image/png", []byte("fake"))
	req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
	req.Header.Set("Content-Type", contentType)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)

	h.UploadMedicalRecordImage(c)

	assert.True(t, recorder.Code == http.StatusInternalServerError || recorder.Code == http.StatusTooManyRequests)
	assert.Contains(t, recorder.Body.String(), "upload quota unavailable")
}

func TestUploadMedicalRecordImage_ParallelSharedStoreEnforcesConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newMemoryMedicalRecordImageUploadQuotaStore()

	block := make(chan struct{})
	var entered atomic.Int32
	var uploadCalls atomic.Int32

	newHandler := func() *MedicalRecordImageHandler {
		return NewMedicalRecordImageHandler(
			&mockMedicalRecordImageService{
				createFn: func(_ context.Context, _, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
					return &model.MedicalRecordImage{
						ID:              uint64(uploadCalls.Load()),
						MedicalRecordID: medicalRecordID,
						ImageURL:        input.ImageURL,
					}, nil
				},
			},
			&mockMedicalRecordService{
				getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
				},
			},
			&mockMedicalRecordImageUploader{
				uploadFn: func(context.Context, string, io.Reader, string) (string, error) {
					uploadCalls.Add(1)
					if entered.Add(1) <= int32(medicalRecordImageUploadStaffMaxConcurrent) {
						<-block
					}
					return "https://uploads.example.test/photo.png", nil
				},
			},
			store,
		)
	}

	const total = medicalRecordImageUploadStaffMaxConcurrent + 2
	results := make([]int, total)
	var wg sync.WaitGroup
	started := make(chan struct{}, total)

	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			h := newHandler()
			body, contentType := buildImageMultipart(t, "photo.png", "image/png", []byte("fake-image-bytes"))
			req := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
			req.Header.Set("Content-Type", contentType)
			req.ContentLength = int64(body.Len())
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: "5"}}
			setClinicID(c)
			setStaffID(c)
			started <- struct{}{}
			h.UploadMedicalRecordImage(c)
			results[idx] = recorder.Code
		}(i)
	}

	for i := 0; i < total; i++ {
		<-started
	}
	// Give acquirers a moment to race on the shared store before unblocking uploads.
	time.Sleep(50 * time.Millisecond)
	close(block)
	wg.Wait()

	var created, limited int
	for _, code := range results {
		switch code {
		case http.StatusCreated:
			created++
		case http.StatusTooManyRequests:
			limited++
		default:
			t.Errorf("unexpected status %d", code)
		}
	}
	assert.Equal(t, medicalRecordImageUploadStaffMaxConcurrent, created)
	assert.Equal(t, 2, limited)
}

type countingUploadRequestBody struct {
	Reader    *bytes.Buffer
	bytesRead int
}

func (r *countingUploadRequestBody) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.bytesRead += n
	return n, err
}

func (r *countingUploadRequestBody) Len() int {
	return r.Reader.Len()
}
