package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountingService 会計サービスインターフェース
type AccountingService interface {
	GetAllAccounting(ctx context.Context) ([]model.Accounting, error)
	GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error)
	GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error)
	GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error)
	GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error)
	CreateAccounting(ctx context.Context, req *model.CreateAccountingRequest) (*model.Accounting, error)
	UpdateAccounting(ctx context.Context, id string, req *model.UpdateAccountingRequest) (*model.Accounting, error)
	DeleteAccounting(ctx context.Context, id string) error
}

// GetAllAccounting 全ての会計を取得
func (s *Service) GetAllAccounting(ctx context.Context) ([]model.Accounting, error) {
	return s.accountingRepo.GetAllAccounting(ctx)
}

// GetAccountingByID IDで会計を取得
func (s *Service) GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid accounting ID format")
	}

	acc, err := s.accountingRepo.GetAccountingByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, apperrors.WrapNotFound("accounting with id %s not found", id)
	}

	return acc, nil
}

// GetAccountingByPetID ペットIDで会計を取得
func (s *Service) GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error) {
	uid, err := uuid.Parse(petID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	return s.accountingRepo.GetAccountingByPetID(ctx, uid.String())
}

// GetAccountingByOwnerID 飼い主IDで会計を取得
func (s *Service) GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error) {
	uid, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	return s.accountingRepo.GetAccountingByOwnerID(ctx, uid.String())
}

// GetAccountingByStatus ステータスで会計を取得
func (s *Service) GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error) {
	return s.accountingRepo.GetAccountingByStatus(ctx, status)
}

// CreateAccounting 会計を作成
func (s *Service) CreateAccounting(ctx context.Context, req *model.CreateAccountingRequest) (*model.Accounting, error) {
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid owner ID format")
	}

	petID, err := uuid.Parse(req.PetID)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid pet ID format")
	}

	scheduledDate, err := time.Parse("2006-01-02", req.ScheduledDate)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid scheduled_date format, expected YYYY-MM-DD")
	}

	acc := &model.Accounting{
		ID:            uuid.New(),
		OwnerID:       ownerID,
		OwnerName:     req.OwnerName,
		PetID:         petID,
		PetName:       req.PetName,
		PetSpecies:    req.PetSpecies,
		Status:        req.Status,
		ScheduledDate: scheduledDate,
		Memo:          req.Memo,
	}

	if acc.Status == "" {
		acc.Status = "waiting"
	}

	if req.MedicalRecordID != nil {
		mrUID, err := uuid.Parse(*req.MedicalRecordID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid medical record ID format")
		}
		acc.MedicalRecordID = &mrUID
	}

	if req.HospitalizationID != nil {
		hospUID, err := uuid.Parse(*req.HospitalizationID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid hospitalization ID format")
		}
		acc.HospitalizationID = &hospUID
	}

	if err := s.accountingRepo.CreateAccounting(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// UpdateAccounting 会計を更新
func (s *Service) UpdateAccounting(ctx context.Context, id string, req *model.UpdateAccountingRequest) (*model.Accounting, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("invalid accounting ID format")
	}

	acc, err := s.accountingRepo.GetAccountingByID(ctx, uid.String())
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, apperrors.WrapNotFound("accounting with id %s not found", id)
	}

	// Update fields
	if req.OwnerName != nil {
		acc.OwnerName = *req.OwnerName
	}
	if req.PetName != nil {
		acc.PetName = *req.PetName
	}
	if req.PetSpecies != nil {
		acc.PetSpecies = req.PetSpecies
	}
	if req.Status != nil {
		acc.Status = *req.Status
	}
	if req.ScheduledDate != nil {
		t, err := time.Parse("2006-01-02", *req.ScheduledDate)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid scheduled_date format, expected YYYY-MM-DD")
		}
		acc.ScheduledDate = t
	}
	if req.Memo != nil {
		acc.Memo = *req.Memo
	}
	if req.MedicalRecordID != nil {
		mrUID, err := uuid.Parse(*req.MedicalRecordID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid medical record ID format")
		}
		acc.MedicalRecordID = &mrUID
	}
	if req.HospitalizationID != nil {
		hospUID, err := uuid.Parse(*req.HospitalizationID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid hospitalization ID format")
		}
		acc.HospitalizationID = &hospUID
	}
	if req.OwnerID != nil {
		ownerUID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid owner ID format")
		}
		acc.OwnerID = ownerUID
	}
	if req.PetID != nil {
		petUID, err := uuid.Parse(*req.PetID)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("invalid pet ID format")
		}
		acc.PetID = petUID
	}

	if err := s.accountingRepo.UpdateAccounting(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// DeleteAccounting 会計を削除
func (s *Service) DeleteAccounting(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid accounting ID format")
	}

	return s.accountingRepo.DeleteAccounting(ctx, uid.String())
}
