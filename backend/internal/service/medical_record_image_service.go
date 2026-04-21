package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateMedicalRecordImageInput は診療画像作成の Input DTO
type CreateMedicalRecordImageInput struct {
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

// MedicalRecordImageService は診療画像のビジネスロジック
type MedicalRecordImageService interface {
	List(ctx context.Context, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error)
	Delete(ctx context.Context, medicalRecordID, imageID uint64) error
}

type medicalRecordImageService struct {
	repo repository.MedicalRecordImageRepository
}

// NewMedicalRecordImageService は MedicalRecordImageService を初期化して返す
func NewMedicalRecordImageService(repo repository.MedicalRecordImageRepository) MedicalRecordImageService {
	return &medicalRecordImageService{repo: repo}
}

func (s *medicalRecordImageService) List(ctx context.Context, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	result, err := s.repo.FindByMedicalRecordID(ctx, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list record images")
	}
	return result, nil
}

func (s *medicalRecordImageService) Create(ctx context.Context, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
	if err := validateMedicalImageType(string(input.ImageType)); err != nil {
		return nil, err
	}
	imageType := input.ImageType
	if imageType == "" {
		imageType = model.MedicalImageTypeOther
	}

	image := &model.MedicalRecordImage{
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
		return nil, apperrors.Wrap(err, "failed to create record image")
	}
	slog.InfoContext(ctx, "record image created",
		slog.Uint64("image_id", image.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return image, nil
}

func (s *medicalRecordImageService) Delete(ctx context.Context, medicalRecordID, imageID uint64) error {
	image, err := s.repo.FindByID(ctx, imageID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get record image")
	}
	if image.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("medical_record_image", "not found in this medical record")
	}
	if err := s.repo.Delete(ctx, imageID); err != nil {
		return apperrors.Wrap(err, "failed to delete record image")
	}
	slog.InfoContext(ctx, "record image deleted", slog.Uint64("image_id", imageID))
	return nil
}
