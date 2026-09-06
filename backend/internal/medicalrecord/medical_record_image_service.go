package medicalrecord

import (
	"context"
	"fmt"
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
	repo             MedicalRecordImageRepository
	medRec           medicalRecordLocker
	pets             petFinder
	examinations     medicalRecordImageExaminationFinder
	staffs           clinicalStaffLocker
	staffAssignments clinicalStaffAssignmentLocker
	transactor       Transactor
}

type medicalRecordImageExaminationFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
}

type clinicalStaffLocker interface {
	LockActiveByIDForShare(ctx context.Context, id uint64) (*model.Staff, error)
}

type clinicalStaffAssignmentLocker interface {
	LockActiveByStaffAndClinic(ctx context.Context, staffID, clinicID uint64) (*model.StaffClinicAssignment, error)
}

// NewMedicalRecordImageService は MedicalRecordImageService を初期化して返す
func NewMedicalRecordImageService(repo MedicalRecordImageRepository, medRec medicalRecordLocker, transactor Transactor) MedicalRecordImageService {
	return &medicalRecordImageService{repo: repo, medRec: medRec, transactor: transactor}
}

// NewMedicalRecordImageServiceWithRelationValidation は、request 由来の exam_id / staff_id と
// 親カルテ紐付けペットの死亡状態を保存トランザクション内で検証する依存を明示的に注入する。
func NewMedicalRecordImageServiceWithRelationValidation(
	repo MedicalRecordImageRepository,
	medRec medicalRecordLocker,
	pets petFinder,
	examinations medicalRecordImageExaminationFinder,
	staffs clinicalStaffLocker,
	staffAssignments clinicalStaffAssignmentLocker,
	transactor Transactor,
) MedicalRecordImageService {
	return &medicalRecordImageService{
		repo:             repo,
		medRec:           medRec,
		pets:             pets,
		examinations:     examinations,
		staffs:           staffs,
		staffAssignments: staffAssignments,
		transactor:       transactor,
	}
}

func (s *medicalRecordImageService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.MedicalRecordImage, error) {
	result, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list record images")
	}
	return result, nil
}

func (s *medicalRecordImageService) Create(ctx context.Context, clinicID, medicalRecordID uint64, input *CreateMedicalRecordImageInput) (*model.MedicalRecordImage, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if err := validateMedicalImageType(string(input.ImageType)); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate medical image type")
	}
	if clinicID == 0 || medicalRecordID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id and medical_record_id are required")
	}
	if s.repo == nil || s.medRec == nil {
		return nil, apperrors.WrapInternalServerError("medical record image persistence dependencies are required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("medical record image transaction dependency is required")
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
		// 死亡ペット検証のため親行を保持する（lockDraftMedicalRecord は親を破棄するため直接 lock）。
		parent, err := s.medRec.LockByIDForUpdate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to find medical record")
		}
		if parent == nil {
			return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", medicalRecordID))
		}
		if parent.Status == model.MedicalRecordStatusFinalized {
			return apperrors.WrapConflict("確定済みカルテに画像を追加できません")
		}
		// SEC-CS-F14: 死亡ペットのカルテへの画像作成を fail-closed で拒否（storage 作成前の service 境界）。
		if err := s.validateLinkedPetNotDeceased(txCtx, clinicID, parent.PetID); err != nil {
			return err
		}
		if err := s.validateExaminationReference(txCtx, clinicID, medicalRecordID, input.ExamID); err != nil {
			return err
		}
		if err := validateClinicalStaffReference(
			txCtx,
			clinicID,
			input.StaffID,
			s.staffs,
			s.staffAssignments,
		); err != nil {
			return err
		}

		if err := s.repo.Create(txCtx, image); err != nil {
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

// validateLinkedPetNotDeceased は親カルテに紐づくペットが生存していることを検証する（SEC-CS-F14）。
// PetID 欠落・pets 依存欠落は fail-closed。死亡時は Create も storage 永続化も行わない。
func (s *medicalRecordImageService) validateLinkedPetNotDeceased(
	ctx context.Context,
	clinicID uint64,
	petID *uint64,
) error {
	if petID == nil || *petID == 0 {
		return apperrors.WrapInvalidInput("medical record has no linked pet")
	}
	if s.pets == nil {
		return apperrors.WrapInternalServerError("medical record image pet validation dependency is required")
	}
	return validatePetNotDeceased(
		ctx,
		s.pets,
		clinicID,
		*petID,
		"死亡したペットのカルテに画像を追加できません",
	)
}

func (s *medicalRecordImageService) validateExaminationReference(
	ctx context.Context,
	clinicID, medicalRecordID uint64,
	examinationID *uint64,
) error {
	if examinationID == nil {
		return nil
	}
	if *examinationID == 0 {
		return apperrors.WrapInvalidInput("exam_id must be greater than zero")
	}
	if s.examinations == nil {
		return apperrors.WrapInternalServerError("medical record image examination validation dependency is required")
	}
	examination, err := s.examinations.FindByID(ctx, clinicID, *examinationID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify image examination ownership")
	}
	if examination == nil ||
		examination.ID != *examinationID ||
		examination.ClinicID != clinicID ||
		examination.MedicalRecordID == nil ||
		*examination.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("examination", "relation")
	}
	return nil
}

func validateClinicalStaffReference(
	ctx context.Context,
	clinicID uint64,
	staffID *uint64,
	staffs clinicalStaffLocker,
	assignments clinicalStaffAssignmentLocker,
) error {
	if staffID == nil {
		return nil
	}
	if *staffID == 0 {
		return apperrors.WrapInvalidInput("staff_id must be greater than zero")
	}
	if staffs == nil || assignments == nil {
		return apperrors.WrapInternalServerError("clinical staff validation dependencies are required")
	}

	staff, err := staffs.LockActiveByIDForShare(ctx, *staffID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify clinical staff")
	}
	if staff == nil || staff.ID != *staffID || !staff.IsActive {
		return apperrors.WrapNotFound("staff", "active assignment")
	}

	assignment, err := assignments.LockActiveByStaffAndClinic(ctx, *staffID, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify clinical staff assignment")
	}
	if assignment == nil || assignment.StaffID != *staffID || assignment.ClinicID != clinicID {
		return apperrors.WrapNotFound("staff", "active assignment")
	}
	return nil
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
			return apperrors.Wrap(err, "failed to delete record image")
		}
		return nil
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "record image deleted", slog.Uint64("image_id", imageID))
	return nil
}
