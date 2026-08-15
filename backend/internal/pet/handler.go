package pet

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware is pet's consumer-side RBAC middleware view.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// PetOwnerDetailsFinder supplies clinic-scoped owner names for the public
// secondary-owner response.
type PetOwnerDetailsFinder interface {
	FindByIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) ([]*model.Owner, error)
}

// Handler owns pet, animal-species, and chronic-condition HTTP behavior.
type Handler struct {
	pets              Service
	animalSpecies     AnimalSpeciesService
	chronicConditions ChronicConditionService
	petOwners         PetOwnerService
	petOwnerDetails   PetOwnerDetailsFinder
	requirePermission PermissionMiddleware
}

// NewHandler constructs the aggregate pet HTTP boundary.
func NewHandler(
	pets Service,
	animalSpecies AnimalSpeciesService,
	chronicConditions ChronicConditionService,
	requirePermission PermissionMiddleware,
) *Handler {
	return NewHandlerWithPetOwners(
		pets,
		animalSpecies,
		chronicConditions,
		nil,
		nil,
		requirePermission,
	)
}

// NewHandlerWithPetOwners constructs the aggregate pet HTTP boundary with
// secondary-owner dependencies.
func NewHandlerWithPetOwners(
	pets Service,
	animalSpecies AnimalSpeciesService,
	chronicConditions ChronicConditionService,
	petOwners PetOwnerService,
	petOwnerDetails PetOwnerDetailsFinder,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		pets:              pets,
		animalSpecies:     animalSpecies,
		chronicConditions: chronicConditions,
		petOwners:         petOwners,
		petOwnerDetails:   petOwnerDetails,
		requirePermission: requirePermission,
	}
}
