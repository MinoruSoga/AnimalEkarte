package medicalrecord

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestResolveMedicalRecordImageStorageKey(t *testing.T) {
	tests := []struct {
		name       string
		stored     string
		wantKey    string
		wantObject bool
		wantErr    bool
	}{
		{
			name:       "scheme-less medical-records key",
			stored:     "medical-records/5/uuid.png",
			wantKey:    "medical-records/5/uuid.png",
			wantObject: true,
		},
		{
			name:       "local /uploads prefix uses rest as key",
			stored:     "/uploads/medical-records/5/uuid.png",
			wantKey:    "medical-records/5/uuid.png",
			wantObject: true,
		},
		{
			name:       "https path containing medical-records/ extracts suffix",
			stored:     "https://cdn.example/medical-records/5/uuid.png",
			wantKey:    "medical-records/5/uuid.png",
			wantObject: true,
		},
		{
			name:       "http nested path extracts medical-records suffix",
			stored:     "http://cdn.example/foo/medical-records/5/uuid.png",
			wantKey:    "medical-records/5/uuid.png",
			wantObject: true,
		},
		{
			name:       "https query is stripped from extracted key",
			stored:     "https://cdn.example/medical-records/5/uuid.png?X-Amz-Signature=abc",
			wantKey:    "medical-records/5/uuid.png",
			wantObject: true,
		},
		{
			name:       "empty is not a storage object",
			stored:     "",
			wantKey:    "",
			wantObject: false,
		},
		{
			name:       "JSON-create https without medical-records suffix is not extracted",
			stored:     "https://example.test/a.png",
			wantKey:    "",
			wantObject: false,
		},
		{
			name:       "https path with medical-records as filename prefix is not a storage object",
			stored:     "https://example.test/medical-records-policy.png",
			wantKey:    "",
			wantObject: false,
		},
		{
			name:       "local /uploads without medical-records key is not a storage object",
			stored:     "/uploads/image.png",
			wantKey:    "",
			wantObject: false,
		},
		{
			name:       "ambiguous https medical-records path without object key is not extracted",
			stored:     "https://cdn.example/medical-records",
			wantKey:    "",
			wantObject: false,
		},
		{
			name:       "https medical-records/ with empty remainder is fail-closed",
			stored:     "https://cdn.example/medical-records/",
			wantKey:    "",
			wantObject: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, isObject, err := resolveMedicalRecordImageStorageKey(tt.stored)
			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, isObject)
				assert.Empty(t, gotKey)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantObject, isObject)
			assert.Equal(t, tt.wantKey, gotKey)
		})
	}
}

func TestMedicalRecordImageKeyBelongsToRecord(t *testing.T) {
	assert.True(t, medicalRecordImageKeyBelongsToRecord("medical-records/5/uuid.png", 5))
	assert.False(t, medicalRecordImageKeyBelongsToRecord("medical-records/99/uuid.png", 5))
	assert.False(t, medicalRecordImageKeyBelongsToRecord("medical-records/50/uuid.png", 5))
	assert.False(t, medicalRecordImageKeyBelongsToRecord("medical-records/5/", 5))
	assert.False(t, medicalRecordImageKeyBelongsToRecord("medical-records/5/../other.png", 5))
	assert.False(t, medicalRecordImageKeyBelongsToRecord("other/5/uuid.png", 5))
}

func TestToMedicalRecordImageResponse(t *testing.T) {
	takenAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	timestamps := time.Date(2026, 5, 2, 11, 30, 0, 0, time.UTC)
	examID := uint64(7)
	staffID := uint64(3)

	tests := []struct {
		name  string
		image *model.MedicalRecordImage
		check func(t *testing.T, got medicalRecordImageResponse)
	}{
		{
			name: "full fields with staff",
			image: &model.MedicalRecordImage{
				ID:              1,
				MedicalRecordID: 10,
				ImageURL:        "https://example.test/a.png",
				ThumbnailURL:    "https://example.test/a-thumb.png",
				FileName:        "a.png",
				FileSize:        1024,
				MimeType:        "image/png",
				ImageType:       model.MedicalImageTypeXray,
				Description:     "desc",
				TakenAt:         &takenAt,
				ExamID:          &examID,
				StaffID:         &staffID,
				SortOrder:       3,
				CreatedAt:       timestamps,
				UpdatedAt:       timestamps,
				Staff:           &model.Staff{ID: staffID, Name: "田中"},
			},
			check: func(t *testing.T, got medicalRecordImageResponse) {
				assert.Equal(t, uint64(1), got.ID)
				assert.Equal(t, uint64(10), got.MedicalRecordID)
				assert.Equal(t, "https://example.test/a.png", got.ImageURL)
				assert.Equal(t, "https://example.test/a-thumb.png", got.ThumbnailURL)
				assert.Equal(t, "a.png", got.FileName)
				assert.Equal(t, int64(1024), got.FileSize)
				assert.Equal(t, "image/png", got.MimeType)
				assert.Equal(t, string(model.MedicalImageTypeXray), got.ImageType)
				assert.Equal(t, "desc", got.Description)
				assert.Equal(t, 3, got.SortOrder)
				if assert.NotNil(t, got.TakenAt) {
					assert.True(t, got.TakenAt.Equal(takenAt))
				}
				assert.Equal(t, &examID, got.ExamID)
				assert.Equal(t, &staffID, got.StaffID)
				assert.True(t, got.CreatedAt.Equal(timestamps))
				assert.True(t, got.UpdatedAt.Equal(timestamps))
				if assert.NotNil(t, got.Staff) {
					assert.Equal(t, staffID, got.Staff.ID)
					assert.Equal(t, "田中", got.Staff.Name)
				}
			},
		},
		{
			name: "nil optional fields without staff",
			image: &model.MedicalRecordImage{
				ID:              2,
				MedicalRecordID: 20,
				ImageType:       model.MedicalImageTypeOther,
				CreatedAt:       timestamps,
				UpdatedAt:       timestamps,
			},
			check: func(t *testing.T, got medicalRecordImageResponse) {
				assert.Nil(t, got.TakenAt)
				assert.Nil(t, got.ExamID)
				assert.Nil(t, got.StaffID)
				assert.Nil(t, got.Staff)
				assert.Equal(t, string(model.MedicalImageTypeOther), got.ImageType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toMedicalRecordImageResponse(tt.image)
			tt.check(t, got)
		})
	}
}
