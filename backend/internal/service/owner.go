package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/errors"
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
		return nil, errors.ErrInvalidInput
	}
	owner, err := s.ownerRepo.GetOwnerByID(ctx, uid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return owner, nil
}

// CreateOwner creates a new owner.
func (s *Service) CreateOwner(ctx context.Context, req *model.CreateOwnerRequest) (*model.Owner, error) {
	owner := &model.Owner{
		Name:     req.Name,
		NameKana: req.NameKana,
		Phone:    req.Phone,
		Email:    req.Email,
		Address:  req.Address,
		Notes:    req.Notes,
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
		return nil, errors.ErrInvalidInput
	}

	owner, err := s.ownerRepo.GetOwnerByID(ctx, uid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find owner for update: %w", err)
	}

	if req.Name != "" {
		owner.Name = req.Name
	}
	if req.NameKana != "" {
		owner.NameKana = req.NameKana
	}
	if req.Phone != "" {
		owner.Phone = req.Phone
	}
	if req.Email != "" {
		owner.Email = req.Email
	}
	if req.Address != "" {
		owner.Address = req.Address
	}
	if req.Notes != "" {
		owner.Notes = req.Notes
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
		return errors.ErrInvalidInput
	}

	// Check existence first
	_, err = s.ownerRepo.GetOwnerByID(ctx, uid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.ErrNotFound
		}
		return err
	}

	if err := s.ownerRepo.DeleteOwner(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete owner: %w", err)
	}
	return nil
}
