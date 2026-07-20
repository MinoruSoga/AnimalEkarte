package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error)
	Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, imageID uint64) error
}

type medicalRecordImageService struct {
	repo       MedicalRecordImageRepository
	medRec     medicalRecordLocker
	transactor Transactor
}

// NewMedicalRecordImageService は MedicalRecordImageService を初期化して返す
func NewMedicalRecordImageService(repo MedicalRecordImageRepository, medRec medicalRecordLocker, transactor Transactor) MedicalRecordImageService {
	return &medicalRecordImageService{repo: repo, medRec: medRec, transactor: transactor}
}

func (s *medicalRecordImageService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	result, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list record images", "error", err)
		return nil, apperrors.Wrap(err, "failed to list record images")
	}
	return result, nil
}

func (s *medicalRecordImageService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
	if err := validateMedicalImageType(string(input.ImageType)); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate medical image type")
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

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// SD-2 + BE-refactor.md X-11: 親カルテが確定済みの場合は画像追加を拒否。LockByIDForUpdate の
		// 行ロックで finalize（medical_record_repository.Update の draft-only WHERE）と直列化し、
		// 確定と同時の画像追加が確定済みカルテに混入する競合を防ぐ（examination_service.go Create 同型）。
		if err := lockDraftMedicalRecord(txCtx, s.medRec, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテに画像を追加できません"); err != nil {
			return err
		}

		if err := s.repo.Create(txCtx, image); err != nil {
			slog.ErrorContext(txCtx, "failed to create record image", "error", err)
			return apperrors.Wrap(err, "failed to create record image")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "record image created",
		slog.Uint64("image_id", image.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return image, nil
}

func (s *medicalRecordImageService) Delete(ctx context.Context, clinicID, medicalRecordID, imageID uint64) error {
	image, err := s.repo.FindByID(ctx, clinicID, imageID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get record image")
	}
	if image.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("medical_record_image", "not found in this medical record")
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// SD-2 + BE-refactor.md H-8d: 親カルテが確定済みの場合は画像削除を拒否。LockByIDForUpdate の
		// 行ロックで finalize と直列化し、確定と同時の画像削除が確定済みカルテに混入する競合を防ぐ
		// （examination_service.go Delete 同型）。
		if err := lockDraftMedicalRecord(txCtx, s.medRec, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテの画像は削除できません"); err != nil {
			return err
		}

		if err := s.repo.Delete(txCtx, clinicID, imageID); err != nil {
			slog.ErrorContext(txCtx, "failed to delete record image", "error", err, "clinic_id", clinicID, "image_id", imageID)
			return apperrors.Wrap(err, "failed to delete record image")
		}
		return nil
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "record image deleted", slog.Uint64("image_id", imageID))
	return nil
}
