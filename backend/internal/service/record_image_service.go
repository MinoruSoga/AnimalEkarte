package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateRecordImageInput は診療画像作成の Input DTO
type CreateRecordImageInput struct {
	ImageURL     string
	ThumbnailURL string
	FileName     string
	FileSize     int64
	MimeType     string
	ImageType    model.MedicalImageType
	Description  string
	TakenAt      *time.Time
	ExamID       *uint64
	StaffID      *uint64
	SortOrder    int
}

// RecordImageService は診療画像のビジネスロジック
type RecordImageService interface {
	List(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateRecordImageInput) (*model.RecordImage, error)
	Delete(ctx context.Context, medicalRecordID, imageID uint64) error
}

type recordImageService struct {
	repo repository.RecordImageRepository
}

// NewRecordImageService は RecordImageService を初期化して返す
func NewRecordImageService(repo repository.RecordImageRepository) RecordImageService {
	return &recordImageService{repo: repo}
}

func (s *recordImageService) List(ctx context.Context, medicalRecordID uint64) ([]model.RecordImage, error) {
	return s.repo.ListByMedicalRecordID(ctx, medicalRecordID)
}

func (s *recordImageService) Create(ctx context.Context, medicalRecordID uint64, input *CreateRecordImageInput) (*model.RecordImage, error) {
	if err := validateMedicalImageType(string(input.ImageType)); err != nil {
		return nil, err
	}
	imageType := input.ImageType
	if imageType == "" {
		imageType = model.MedicalImageTypeOther
	}

	image := &model.RecordImage{
		MedicalRecordID: medicalRecordID,
		ImageURL:        input.ImageURL,
		ThumbnailURL:    input.ThumbnailURL,
		FileName:        input.FileName,
		FileSize:        input.FileSize,
		MimeType:        input.MimeType,
		ImageType:       imageType,
		Description:     input.Description,
		TakenAt:         input.TakenAt,
		ExamID:          input.ExamID,
		StaffID:         input.StaffID,
		SortOrder:       input.SortOrder,
	}
	if err := s.repo.Create(ctx, image); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "record image created",
		slog.Uint64("image_id", image.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return image, nil
}

func (s *recordImageService) Delete(ctx context.Context, medicalRecordID, imageID uint64) error {
	image, err := s.repo.FindByID(ctx, imageID)
	if err != nil {
		return err
	}
	if image.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("record_image", "not found in this medical record")
	}
	if err := s.repo.Delete(ctx, imageID); err != nil {
		return err
	}
	slog.InfoContext(ctx, "record image deleted", slog.Uint64("image_id", imageID))
	return nil
}
