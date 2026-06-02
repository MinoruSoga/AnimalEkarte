package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- MedicalRecordImage モック ----

type mockMedicalRecordImageRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	findByIDFn              func(ctx context.Context, clinicID, imageID uint64) (*model.MedicalRecordImage, error)
	createFn                func(ctx context.Context, image *model.MedicalRecordImage) error
	deleteFn                func(ctx context.Context, clinicID, imageID uint64) error
}

func (m *mockMedicalRecordImageRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockMedicalRecordImageRepository) FindByID(ctx context.Context, clinicID, imageID uint64) (*model.MedicalRecordImage, error) {
	return m.findByIDFn(ctx, clinicID, imageID)
}

func (m *mockMedicalRecordImageRepository) Create(ctx context.Context, image *model.MedicalRecordImage) error {
	return m.createFn(ctx, image)
}

func (m *mockMedicalRecordImageRepository) Delete(ctx context.Context, clinicID, imageID uint64) error {
	return m.deleteFn(ctx, clinicID, imageID)
}

// ---- Tests ----

func TestMedicalRecordImageService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoImages      []model.MedicalRecordImage
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns record images for medical record",
			medicalRecordID: 1,
			repoImages: []model.MedicalRecordImage{
				{ID: 1, MedicalRecordID: 1, ImageType: model.MedicalImageTypeXray, FileName: "xray1.jpg"},
				{ID: 2, MedicalRecordID: 1, ImageType: model.MedicalImageTypeEcho, FileName: "ultrasound1.jpg"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no images exist",
			medicalRecordID: 999,
			repoImages:      []model.MedicalRecordImage{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoImages:      nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordImageRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.MedicalRecordImage, error) {
					return tt.repoImages, tt.repoErr
				},
			}
			svc := NewMedicalRecordImageService(repo)

			images, err := svc.List(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, images, tt.wantLen)
			}
		})
	}
}

func TestMedicalRecordImageService_Create(t *testing.T) {
	takenAt := time.Now()
	examID := uint64(1)
	staffID := uint64(5)

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateMedicalRecordImageInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates record image with explicit image type",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:     "https://example.com/image1.jpg",
				ThumbnailURL: "https://example.com/thumb1.jpg",
				FileName:     "image1.jpg",
				FileSize:     1024000,
				MimeType:     "image/jpeg",
				ImageType:    model.MedicalImageTypeXray,
				Description:  "X-ray of chest",
				TakenAt:      &takenAt,
				ExamID:       &examID,
				StaffID:      &staffID,
				SortOrder:    1,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates record image with default image type when empty",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:     "https://example.com/image2.jpg",
				ThumbnailURL: "https://example.com/thumb2.jpg",
				FileName:     "image2.jpg",
				FileSize:     2048000,
				MimeType:     "image/jpeg",
				ImageType:    "", // Will default to Other
				Description:  "Generic image",
				SortOrder:    2,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error on invalid image type",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:  "https://example.com/image3.jpg",
				FileName:  "image3.jpg",
				ImageType: model.MedicalImageType("invalid_type"),
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateMedicalRecordImageInput{
				ImageURL:  "https://example.com/image4.jpg",
				FileName:  "image4.jpg",
				ImageType: model.MedicalImageTypeEcho,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordImageRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecordImage) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordImageService(repo)

			image, err := svc.Create(context.Background(), 1, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, image)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, image)
				assert.Equal(t, tt.medicalRecordID, image.MedicalRecordID)
			}
		})
	}
}

func TestMedicalRecordImageService_Delete(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		imageID         uint64
		repoImage       *model.MedicalRecordImage
		findByIDErr     error
		deleteErr       error
		wantErr         bool
	}{
		{
			name:            "deletes record image successfully",
			medicalRecordID: 1,
			imageID:         1,
			repoImage: &model.MedicalRecordImage{
				ID:              1,
				MedicalRecordID: 1,
				FileName:        "image1.jpg",
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     false,
		},
		{
			name:            "returns not found when image does not belong to medical record",
			medicalRecordID: 1,
			imageID:         999,
			repoImage: &model.MedicalRecordImage{
				ID:              999,
				MedicalRecordID: 2, // Different medical record
				FileName:        "image999.jpg",
			},
			findByIDErr: nil,
			deleteErr:   nil,
			wantErr:     true,
		},
		{
			name:            "returns error when image not found",
			medicalRecordID: 1,
			imageID:         999,
			repoImage:       nil,
			findByIDErr:     errors.New("not found"),
			deleteErr:       nil,
			wantErr:         true,
		},
		{
			name:            "returns error when delete fails",
			medicalRecordID: 1,
			imageID:         1,
			repoImage: &model.MedicalRecordImage{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByIDErr: nil,
			deleteErr:   errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordImageRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecordImage, error) {
					return tt.repoImage, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewMedicalRecordImageService(repo)

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID, tt.imageID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
