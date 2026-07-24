package pet

import "github.com/gin-gonic/gin"

// PermissionMiddleware is pet's consumer-side RBAC middleware view.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler owns pet, animal-species, and chronic-condition HTTP behavior.
type Handler struct {
	pets              Service
	animalSpecies     AnimalSpeciesService
	chronicConditions ChronicConditionService
	requirePermission PermissionMiddleware
}

// NewHandler constructs the aggregate pet HTTP boundary.
func NewHandler(
	pets Service,
	animalSpecies AnimalSpeciesService,
	chronicConditions ChronicConditionService,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		pets:              pets,
		animalSpecies:     animalSpecies,
		chronicConditions: chronicConditions,
		requirePermission: requirePermission,
	}
}
