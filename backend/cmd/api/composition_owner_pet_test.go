package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/owner"
)

func TestNewOwnerPetCompositionBuildsRepositoriesServicesAndHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repositories := newOwnerPetRepositories(nil)
	composition := newOwnerPetComposition(repositories, ownerPetCompositionDependencies{})

	require.Same(t, repositories.Owner, composition.OwnerRepository)
	require.Same(t, repositories.Pet, composition.PetRepository)
	require.NotNil(t, composition.OwnerRepository)
	require.NotNil(t, composition.PetRepository)
	require.NotNil(t, composition.OwnerService)
	require.NotNil(t, composition.PetService)

	var permissionCalls int
	var requirePermission owner.PermissionMiddleware = func(_, _ string) gin.HandlerFunc {
		permissionCalls++
		return func(c *gin.Context) {
			c.Next()
		}
	}

	handlers := composition.newHandlers(nil, requirePermission, nil)
	require.NotNil(t, handlers.Owner)
	require.NotNil(t, handlers.Pet)

	router := gin.New()
	protected := router.Group("/api/v1")
	handlers.Owner.RegisterRoutes(protected)
	handlers.Pet.RegisterRoutes(protected)

	// Owner surface is 11 routes (SOLO-33 removed clinic-scoped aliases); pet surface is 20.
	require.Len(t, router.Routes(), 31)
	require.Equal(t, 31, permissionCalls)
}
