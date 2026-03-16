package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- RecordImage モック ----

type mockRecordImageRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error)
	findByIDFn              func(ctx context.Context, imageID uint64) (*model.RecordImage, error)
	createFn                func(ctx context.Context, image *model.RecordImage) error
	deleteFn                func(ctx context.Context, imageID uint64) error
}

func (m *mockRecordImageRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error) {
	return m.listByMedicalRecordIDFn(ctx, medicalRecordID)
}

func (m *mockRecordImageRepository) FindByID(ctx context.Context, imageID uint64) (*model.RecordImage, error) {
	return m.findByIDFn(ctx, imageID)
}

func (m *mockRecordImageRepository) Create(ctx context.Context, image *model.RecordImage) error {
	return m.createFn(ctx, image)
}

func (m *mockRecordImageRepository) Delete(ctx context.Context, imageID uint64) error {
	return m.deleteFn(ctx, imageID)
}

// ---- Tests ----

func TestRecordImageService_List(t *testing.T) {
	tests := []struct {
		name              string
		medicalRecordID   uint64
		repoImages        []model.RecordImage
		repoErr           error
		wantLen           int
		wantErr           bool
	}{
		{
			name:            "returns record images for medical record",
			medicalRecordID: 1,
			repoImages: []model.RecordImage{
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
			repoImages:      []model.RecordImage{},
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
			repo := &mockRecordImageRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _ uint64) ([]model.RecordImage, error) {
					return tt.repoImages, tt.repoErr
				},
			}
			svc := NewRecordImageService(repo)

			images, err := svc.List(context.Background(), tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, images, tt.wantLen)
			}
		})
	}
}

func TestRecordImageService_Create(t *testing.T) {
	takenAt := time.Now()
	examID := uint64(1)
	staffID := uint64(5)

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateRecordImageInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates record image with explicit image type",
			medicalRecordID: 1,
			input: &CreateRecordImageInput{
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
			input: &CreateRecordImageInput{
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
			input: &CreateRecordImageInput{
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
			input: &CreateRecordImageInput{
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
			repo := &mockRecordImageRepository{
				createFn: func(_ context.Context, _ *model.RecordImage) error {
					return tt.repoErr
				},
			}
			svc := NewRecordImageService(repo)

			image, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

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

func TestRecordImageService_Delete(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		imageID         uint64
		repoImage       *model.RecordImage
		findByIDErr     error
		deleteErr       error
		wantErr         bool
	}{
		{
			name:            "deletes record image successfully",
			medicalRecordID: 1,
			imageID:         1,
			repoImage: &model.RecordImage{
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
			repoImage: &model.RecordImage{
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
			repoImage: &model.RecordImage{
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
			repo := &mockRecordImageRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.RecordImage, error) {
					return tt.repoImage, tt.findByIDErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewRecordImageService(repo)

			err := svc.Delete(context.Background(), tt.medicalRecordID, tt.imageID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
