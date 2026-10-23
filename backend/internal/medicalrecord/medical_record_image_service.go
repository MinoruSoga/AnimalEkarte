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
	Delete(ctx context.Context, clinicID, medicalRecordID, imageID, actorStaffID uint64, objects medicalRecordImageObjectDeleter) error
	SigningBlocked(ctx context.Context, clinicID, imageID uint64) (bool, error)
	// BlockedImageIDs は一覧表示向けの削除台帳一括突合（N+1 防止）。
	BlockedImageIDs(ctx context.Context, clinicID uint64, imageIDs []uint64) (map[uint64]bool, error)
}

type medicalRecordImageService struct {
	repo             MedicalRecordImageRepository
	medRec           medicalRecordLocker
	pets             petFinder
	examinations     medicalRecordImageExaminationFinder
	staffs           clinicalStaffLocker
	staffAssignments clinicalStaffAssignmentLocker
	transactor       Transactor
	auditTx          AuditTxLogger
	deletions        medicalRecordImageDeletionStore
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
	audits ...ImageDeletionAudit,
) MedicalRecordImageService {
	svc := &medicalRecordImageService{
		repo:             repo,
		medRec:           medRec,
		pets:             pets,
		examinations:     examinations,
		staffs:           staffs,
		staffAssignments: staffAssignments,
		transactor:       transactor,
	}
	if len(audits) > 0 {
		svc.auditTx = audits[0].audit
		svc.deletions = audits[0].store
	}
	return svc
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
		// BE-refactor.md X-11: 親カルテを行ロックして画像追加を直列化する（examination_service.go Create 同型）。
		// EMR-495: 旧 finalized 拒否は廃止した。死亡ペット検証のため親行を保持する
		// （lockDraftMedicalRecord は親を破棄するため直接 lock）。
		parent, err := s.medRec.LockByIDForUpdate(txCtx, clinicID, medicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to find medical record")
		}
		if parent == nil {
			return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", medicalRecordID))
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
