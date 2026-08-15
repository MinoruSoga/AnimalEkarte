package medicalrecord

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func newMedicalRecordImageUploadContext(
	t *testing.T,
	content []byte,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	body, contentType := buildImageMultipart(t, "photo.png", "image/png", content)
	request := httptest.NewRequest(http.MethodPost, "/medical-records/5/images/upload", body)
	request.Header.Set("Content-Type", contentType)
	request.ContentLength = -1

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = request
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	setClinicID(c)
	setStaffID(c)
	return c, recorder
}

func TestUploadMedicalRecordImage_RejectsChunkedOversizedRequestBeforeUpload(t *testing.T) {
	c, recorder := newMedicalRecordImageUploadContext(
		t,
		make([]byte, medicalRecordImageMaxRequestSize),
	)
	h := newHandlerWithMedicalRecordImageSvc(
		&mockMedicalRecordService{
			getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
			},
		},
		&mockMedicalRecordImageService{
			createFn: func(context.Context, uint64, uint64, *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				t.Fatal("service create must not run for an oversized request")
				return nil, nil
			},
		},
		&mockMedicalRecordImageUploader{
			uploadFn: func(context.Context, string, io.Reader, string) (string, error) {
				t.Fatal("uploader must not run for an oversized request")
				return "", nil
			},
		},
	)

	h.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
}

func TestUploadMedicalRecordImage_AllowsExactFileLimitWithMultipartOverhead(t *testing.T) {
	c, recorder := newMedicalRecordImageUploadContext(
		t,
		make([]byte, medicalRecordImageMaxUploadSize),
	)
	h := newHandlerWithMedicalRecordImageSvc(
		&mockMedicalRecordService{
			getByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: id, ClinicID: clinicID}, nil
			},
		},
		&mockMedicalRecordImageService{
			createFn: func(_ context.Context, _, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
				return &model.MedicalRecordImage{
					ID:              9,
					MedicalRecordID: medicalRecordID,
					ImageURL:        input.ImageURL,
				}, nil
			},
		},
		&mockMedicalRecordImageUploader{
			uploadFn: func(context.Context, string, io.Reader, string) (string, error) {
				return "https://uploads.example.test/photo.png", nil
			},
		},
	)

	h.UploadMedicalRecordImage(c)

	assert.Equal(t, http.StatusCreated, recorder.Code)
}
