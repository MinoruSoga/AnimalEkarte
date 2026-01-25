package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllOwners retrieves all owners from the database.
func (r *Repository) GetAllOwners(ctx context.Context) ([]model.Owner, error) {
	var owners []model.Owner
	if err := r.db.WithContext(ctx).Find(&owners).Error; err != nil {
		return nil, fmt.Errorf("failed to get all owners: %w", err)
	}
	return owners, nil
}

// GetOwnerByID retrieves a single owner by ID from the database.
func (r *Repository) GetOwnerByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
	var owner model.Owner
	if err := r.db.WithContext(ctx).Preload("Pets").First(&owner, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("owner not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get owner by id: %w", err)
	}
	return &owner, nil
}

// CreateOwner creates a new owner record in the database.
func (r *Repository) CreateOwner(ctx context.Context, owner *model.Owner) error {
	if err := r.db.WithContext(ctx).Create(owner).Error; err != nil {
		return fmt.Errorf("failed to create owner: %w", err)
	}
	return nil
}

// UpdateOwner updates an existing owner record in the database.
func (r *Repository) UpdateOwner(ctx context.Context, owner *model.Owner) error {
	if err := r.db.WithContext(ctx).Save(owner).Error; err != nil {
		return fmt.Errorf("failed to update owner: %w", err)
	}
	return nil
}

// DeleteOwner deletes an owner record from the database.
func (r *Repository) DeleteOwner(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Owner{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete owner: %w", err)
	}
	return nil
}
