package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/validation"
)

// MedicalRecordService カルテサービスインターフェース
type MedicalRecordService interface {
	GetAllMedicalRecords(ctx context.Context) ([]model.MedicalRecord, error)
	GetMedicalRecordByID(ctx context.Context, id string) (*model.MedicalRecord, error)
	GetMedicalRecordsByPetID(ctx context.Context, petID string) ([]model.MedicalRecord, error)
	GetMedicalRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.MedicalRecord, error)
	CreateMedicalRecord(ctx context.Context, req *model.CreateMedicalRecordRequest) (*model.MedicalRecord, error)
	UpdateMedicalRecord(ctx context.Context, id string, req *model.UpdateMedicalRecordRequest) (*model.MedicalRecord, error)
	DeleteMedicalRecord(ctx context.Context, id string) error
}

// GetAllMedicalRecords 全てのカルテを取得
func (s *Service) GetAllMedicalRecords(ctx context.Context) ([]model.MedicalRecord, error) {
	return s.medicalRecordRepo.GetAllMedicalRecords(ctx)
}

// GetMedicalRecordByID IDでカルテを取得
func (s *Service) GetMedicalRecordByID(ctx context.Context, id string) (*model.MedicalRecord, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid medical record ID format")
	}

	record, err := s.medicalRecordRepo.GetMedicalRecordByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if record == nil {
		return nil, apperrors.WrapNotFound("medical record", id)
	}

	return record, nil
}

// GetMedicalRecordsByPetID ペットIDでカルテを取得
func (s *Service) GetMedicalRecordsByPetID(ctx context.Context, petID string) ([]model.MedicalRecord, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.medicalRecordRepo.GetMedicalRecordsByPetID(ctx, uid.String())
}

// GetMedicalRecordsByOwnerID 飼い主IDでカルテを取得
func (s *Service) GetMedicalRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.MedicalRecord, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.medicalRecordRepo.GetMedicalRecordsByOwnerID(ctx, uid.String())
}

// CreateMedicalRecord カルテを作成
func (s *Service) CreateMedicalRecord(ctx context.Context, req *model.CreateMedicalRecordRequest) (*model.MedicalRecord, error) {
	// バリデーション
	if err := validation.ValidateCreateMedicalRecord(req); err != nil {
		return nil, err
	}

	// Dateをパース
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
	}

	// RecordNoを生成
	recordNo := req.RecordNo
	if recordNo == "" {
		recordNo = generateRecordNo()
	}

	record := &model.MedicalRecord{
		RecordNo:        recordNo,
		Date:            date,
		OwnerName:       req.OwnerName,
		PetName:         req.PetName,
		Species:         req.Species,
		ChiefComplaint:  req.ChiefComplaint,
		TreatmentPolicy: req.TreatmentPolicy,
		PhysicalExam:    req.PhysicalExam,
		DiagnosisDetails: req.DiagnosisDetails,
		Doctor:          req.Doctor,
		Status:          req.Status,
	}

	if record.Status == "" {
		record.Status = "作成中"
	}

	// OwnerIDをUUIDに変換（オプショナル）
	if req.OwnerID != nil && *req.OwnerID != "" {
		ownerUID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid owner ID format")
		}
		record.OwnerID = &ownerUID
	}

	// PetIDをUUIDに変換（オプショナル）
	if req.PetID != nil && *req.PetID != "" {
		petUID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid pet ID format")
		}
		record.PetID = &petUID
	}

	// 診断カテゴリID
	if req.Diagnosis1CategoryID != nil && *req.Diagnosis1CategoryID != "" {
		diagUID, err := uuid.Parse(*req.Diagnosis1CategoryID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid diagnosis1_category_id format")
		}
		record.Diagnosis1CategoryID = &diagUID
	}
	if req.Diagnosis1NameID != nil && *req.Diagnosis1NameID != "" {
		diagUID, err := uuid.Parse(*req.Diagnosis1NameID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid diagnosis1_name_id format")
		}
		record.Diagnosis1NameID = &diagUID
	}
	if req.Diagnosis2CategoryID != nil && *req.Diagnosis2CategoryID != "" {
		diagUID, err := uuid.Parse(*req.Diagnosis2CategoryID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid diagnosis2_category_id format")
		}
		record.Diagnosis2CategoryID = &diagUID
	}
	if req.Diagnosis2NameID != nil && *req.Diagnosis2NameID != "" {
		diagUID, err := uuid.Parse(*req.Diagnosis2NameID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid diagnosis2_name_id format")
		}
		record.Diagnosis2NameID = &diagUID
	}

	if err := s.medicalRecordRepo.CreateMedicalRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// UpdateMedicalRecord カルテを更新
func (s *Service) UpdateMedicalRecord(ctx context.Context, id string, req *model.UpdateMedicalRecordRequest) (*model.MedicalRecord, error) {
	// バリデーション
	if err := validation.ValidateUpdateMedicalRecord(req); err != nil {
		return nil, err
	}

	record, err := s.GetMedicalRecordByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 各フィールドを更新
	if req.Date != nil {
		t, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid date format, expected YYYY-MM-DD")
		}
		record.Date = t
	}

	if req.OwnerID != nil {
		if *req.OwnerID == "" {
			record.OwnerID = nil
		} else {
			ownerUID, err := uuid.Parse(*req.OwnerID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid owner ID format")
			}
			record.OwnerID = &ownerUID
		}
	}

	if req.OwnerName != nil {
		record.OwnerName = *req.OwnerName
	}

	if req.PetID != nil {
		if *req.PetID == "" {
			record.PetID = nil
		} else {
			petUID, err := uuid.Parse(*req.PetID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid pet ID format")
			}
			record.PetID = &petUID
		}
	}

	if req.PetName != nil {
		record.PetName = *req.PetName
	}

	if req.Species != nil {
		record.Species = *req.Species
	}

	if req.ChiefComplaint != nil {
		record.ChiefComplaint = *req.ChiefComplaint
	}

	if req.TreatmentPolicy != nil {
		record.TreatmentPolicy = *req.TreatmentPolicy
	}

	if req.PhysicalExam != nil {
		record.PhysicalExam = *req.PhysicalExam
	}

	if req.DiagnosisDetails != nil {
		record.DiagnosisDetails = *req.DiagnosisDetails
	}

	if req.Doctor != nil {
		record.Doctor = *req.Doctor
	}

	if req.Status != nil {
		record.Status = *req.Status
	}

	if req.Diagnosis1CategoryID != nil {
		if *req.Diagnosis1CategoryID == "" {
			record.Diagnosis1CategoryID = nil
		} else {
			diagUID, err := uuid.Parse(*req.Diagnosis1CategoryID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid diagnosis1_category_id format")
			}
			record.Diagnosis1CategoryID = &diagUID
		}
	}
	if req.Diagnosis1NameID != nil {
		if *req.Diagnosis1NameID == "" {
			record.Diagnosis1NameID = nil
		} else {
			diagUID, err := uuid.Parse(*req.Diagnosis1NameID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid diagnosis1_name_id format")
			}
			record.Diagnosis1NameID = &diagUID
		}
	}
	if req.Diagnosis2CategoryID != nil {
		if *req.Diagnosis2CategoryID == "" {
			record.Diagnosis2CategoryID = nil
		} else {
			diagUID, err := uuid.Parse(*req.Diagnosis2CategoryID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid diagnosis2_category_id format")
			}
			record.Diagnosis2CategoryID = &diagUID
		}
	}
	if req.Diagnosis2NameID != nil {
		if *req.Diagnosis2NameID == "" {
			record.Diagnosis2NameID = nil
		} else {
			diagUID, err := uuid.Parse(*req.Diagnosis2NameID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid diagnosis2_name_id format")
			}
			record.Diagnosis2NameID = &diagUID
		}
	}

	if err := s.medicalRecordRepo.UpdateMedicalRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteMedicalRecord カルテを削除
func (s *Service) DeleteMedicalRecord(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid medical record ID format")
	}

	// 存在確認
	_, err = s.GetMedicalRecordByID(ctx, uid.String())
	if err != nil {
		return err
	}

	return s.medicalRecordRepo.DeleteMedicalRecord(ctx, uid.String())
}

// generateRecordNo カルテ番号を生成するヘルパー関数
func generateRecordNo() string {
	return fmt.Sprintf("MR%d", time.Now().Unix())
}
