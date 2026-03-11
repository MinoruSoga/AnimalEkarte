package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllOwners retrieves all owners.
func (s *Service) GetAllOwners(ctx context.Context) ([]model.Owner, error) {
	return s.ownerRepo.GetAllOwners(ctx)
}

// GetOwnerByID retrieves an owner by ID.
func (s *Service) GetOwnerByID(ctx context.Context, id string) (*model.Owner, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	owner, err := s.ownerRepo.GetOwnerByID(ctx, uid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return owner, nil
}

// CreateOwner creates a new owner.
func (s *Service) CreateOwner(ctx context.Context, req *model.CreateOwnerRequest) (*model.Owner, error) {
	owner := &model.Owner{
		OwnerName:      req.OwnerName,
		OwnerNameKana:  req.OwnerNameKana,
		Company:        req.Company,
		PostalCode:     req.PostalCode,
		Address1:       req.Address1,
		Address2:       req.Address2,
		HomePostalCode: req.HomePostalCode,
		HomeAddress1:   req.HomeAddress1,
		HomeAddress2:   req.HomeAddress2,
		Phone:          req.Phone,
		CompanyPhone:   req.CompanyPhone,
		Email:          req.Email,
		Remarks:        req.Remarks,
		IsDangerous:    req.IsDangerous,
		DiscountRate:   req.DiscountRate,
		MembershipType: req.MembershipType,
	}

	if err := s.ownerRepo.CreateOwner(ctx, owner); err != nil {
		return nil, fmt.Errorf("failed to create owner: %w", err)
	}

	return owner, nil
}

// UpdateOwner updates an existing owner.
func (s *Service) UpdateOwner(ctx context.Context, id string, req *model.UpdateOwnerRequest) (*model.Owner, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}

	owner, err := s.ownerRepo.GetOwnerByID(ctx, uid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find owner for update: %w", err)
	}

	if req.OwnerName != nil {
		owner.OwnerName = *req.OwnerName
	}
	if req.OwnerNameKana != nil {
		owner.OwnerNameKana = req.OwnerNameKana
	}
	if req.Company != nil {
		owner.Company = *req.Company
	}
	if req.PostalCode != nil {
		owner.PostalCode = *req.PostalCode
	}
	if req.Address1 != nil {
		owner.Address1 = *req.Address1
	}
	if req.Address2 != nil {
		owner.Address2 = *req.Address2
	}
	if req.HomePostalCode != nil {
		owner.HomePostalCode = *req.HomePostalCode
	}
	if req.HomeAddress1 != nil {
		owner.HomeAddress1 = *req.HomeAddress1
	}
	if req.HomeAddress2 != nil {
		owner.HomeAddress2 = *req.HomeAddress2
	}
	if req.Phone != nil {
		owner.Phone = *req.Phone
	}
	if req.CompanyPhone != nil {
		owner.CompanyPhone = *req.CompanyPhone
	}
	if req.Email != nil {
		owner.Email = *req.Email
	}
	if req.Remarks != nil {
		owner.Remarks = *req.Remarks
	}
	if req.IsDangerous != nil {
		owner.IsDangerous = *req.IsDangerous
	}
	if req.DiscountRate != nil {
		owner.DiscountRate = *req.DiscountRate
	}
	if req.MembershipType != nil {
		owner.MembershipType = *req.MembershipType
	}

	if err := s.ownerRepo.UpdateOwner(ctx, owner); err != nil {
		return nil, fmt.Errorf("failed to update owner: %w", err)
	}

	return owner, nil
}

// DeleteOwner deletes an owner.
func (s *Service) DeleteOwner(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return apperrors.ErrInvalidInput
	}

	// Check existence first
	_, err = s.ownerRepo.GetOwnerByID(ctx, uid)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return apperrors.ErrNotFound
		}
		return err
	}

	if err := s.ownerRepo.DeleteOwner(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete owner: %w", err)
	}
	return nil
}
