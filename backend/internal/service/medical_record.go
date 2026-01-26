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
		return nil, apperrors.WrapNotFound("medical record with id %s not found", id)
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

	// PetIDをUUIDに変換
	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	// OwnerIDをUUIDに変換
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	// VisitDateをパース
	var visitDate time.Time
	if req.VisitDate != "" {
		visitDate, err = parseVisitDate(req.VisitDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid visit date format")
		}
	} else {
		visitDate = time.Now()
	}

	// DoctorIDをUUIDに変換（オプショナル）
	var doctorID *uuid.UUID
	if req.DoctorID != "" {
		doctorUUID, err := uuid.Parse(req.DoctorID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid doctor ID format")
		}
		doctorID = &doctorUUID
	}

	// RecordNoを生成（現在のタイムスタンプを使用）
	recordNo := generateRecordNo()

	record := &model.MedicalRecord{
		RecordNo:       recordNo,
		PetID:          petID,
		OwnerID:        ownerID,
		DoctorID:       doctorID,
		VisitDate:      visitDate,
		VisitType:      req.VisitType,
		ChiefComplaint: req.ChiefComplaint,
		Subjective:     req.Subjective,
		Objective:      req.Objective,
		Assessment:     req.Assessment,
		Plan:           req.Plan,
		SurgeryNotes:   req.SurgeryNotes,
		Diagnosis:      req.Diagnosis,
		Treatment:      req.Treatment,
		Prescription:   req.Prescription,
		Notes:          req.Notes,
		Status:         req.Status,
	}

	// デフォルト値設定
	if record.VisitType == "" {
		record.VisitType = "初診"
	}
	if record.Status == "" {
		record.Status = "作成中"
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
	if req.PetID != nil {
		petID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid pet ID format")
		}
		record.PetID = petID
	}

	if req.OwnerID != nil {
		ownerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid owner ID format")
		}
		record.OwnerID = ownerID
	}

	if req.DoctorID != nil {
		if *req.DoctorID == "" {
			record.DoctorID = nil
		} else {
			doctorID, err := uuid.Parse(*req.DoctorID)
			if err != nil {
				return nil, apperrors.WrapInvalidInput("invalid doctor ID format")
			}
			record.DoctorID = &doctorID
		}
	}

	if req.VisitDate != nil {
		visitDate, err := parseVisitDate(*req.VisitDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid visit date format")
		}
		record.VisitDate = visitDate
	}

	if req.VisitType != nil {
		record.VisitType = *req.VisitType
	}

	if req.ChiefComplaint != nil {
		record.ChiefComplaint = *req.ChiefComplaint
	}

	if req.Subjective != nil {
		record.Subjective = *req.Subjective
	}

	if req.Objective != nil {
		record.Objective = *req.Objective
	}

	if req.Assessment != nil {
		record.Assessment = *req.Assessment
	}

	if req.Plan != nil {
		record.Plan = *req.Plan
	}

	if req.SurgeryNotes != nil {
		record.SurgeryNotes = *req.SurgeryNotes
	}

	if req.Diagnosis != nil {
		record.Diagnosis = *req.Diagnosis
	}

	if req.Treatment != nil {
		record.Treatment = *req.Treatment
	}

	if req.Prescription != nil {
		record.Prescription = *req.Prescription
	}

	if req.Notes != nil {
		record.Notes = *req.Notes
	}

	if req.Status != nil {
		record.Status = *req.Status
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

// parseVisitDate 診察日時をパースするヘルパー関数
func parseVisitDate(dateStr string) (time.Time, error) {
	// RFC3339形式
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}

	// YYYY-MM-DD HH:MM:SS形式
	if t, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
		return t, nil
	}

	// YYYY-MM-DD形式
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}

	return time.Time{}, apperrors.WrapInvalidInput("invalid date format")
}

// generateRecordNo カルテ番号を生成するヘルパー関数
func generateRecordNo() string {
	return fmt.Sprintf("MR%d", time.Now().Unix())
}
