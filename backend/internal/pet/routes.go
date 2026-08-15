package pet

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// RegisterRoutes registers the 20 pet-owned authenticated routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	h.registerPetRoutes(rg)
	h.registerAnimalSpeciesRoutes(rg)
}

func (h *Handler) registerPetRoutes(rg *gin.RouterGroup) {
	h.registerOwnerSharedPetRoutes(rg)
	rg.GET("/owners/:id/report/pets", h.requirePermission(string(model.ResourceOwners), "view"), h.ListOwnerReportPets)

	pets := rg.Group("/pets")
	pets.GET("", h.requirePermission(string(model.ResourceOwners), "view"), h.ListPets)
	pets.GET("/:id", h.requirePermission(string(model.ResourceOwners), "view"), h.GetPet)
	pets.POST("", h.requirePermission(string(model.ResourceOwners), "create"), h.CreatePet)
	pets.PATCH("/:id", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdatePet)
	pets.DELETE("/:id", h.requirePermission(string(model.ResourceOwners), "delete"), h.DeletePet)
	pets.GET("/:id/first-visit", h.requirePermission(string(model.ResourceMedicalRecords), "view"), h.GetPetFirstVisit)

	h.registerPetOwnerRoutes(pets)
	h.registerChronicConditionRoutes(pets)
}

func (h *Handler) registerOwnerSharedPetRoutes(rg *gin.RouterGroup) {
	rg.GET(
		"/owners/:id/shared-pets",
		h.requirePermission(string(model.ResourceOwners), "view"),
		h.ListOwnerSharedPets,
	)
}

func (h *Handler) registerPetOwnerRoutes(pets *gin.RouterGroup) {
	petOwners := pets.Group("/:id/sub-owners")
	petOwners.GET("", h.requirePermission(string(model.ResourceOwners), "view"), h.ListPetOwners)
	petOwners.PUT("", h.requirePermission(string(model.ResourceOwners), "edit"), h.ReplacePetOwners)
}

func (h *Handler) registerChronicConditionRoutes(pets *gin.RouterGroup) {
	chronicConditions := pets.Group("/:id/chronic-conditions")
	chronicConditions.GET("", h.requirePermission(string(model.ResourceOwners), "view"), h.ListChronicConditions)
	chronicConditions.POST("", h.requirePermission(string(model.ResourceOwners), "create"), h.CreateChronicCondition)
	chronicConditions.PATCH("/:cc_id", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateChronicCondition)
	chronicConditions.DELETE("/:cc_id", h.requirePermission(string(model.ResourceOwners), "delete"), h.DeleteChronicCondition)
}

func (h *Handler) registerAnimalSpeciesRoutes(rg *gin.RouterGroup) {
	masters := rg.Group("/masters")
	masters.GET("/animal-species", h.requirePermission(string(model.ResourceMasterAnimalSpecies), "view"), h.ListAnimalSpecies)
	masters.POST("/animal-species", h.requirePermission(string(model.ResourceMasterAnimalSpecies), "create"), h.CreateAnimalSpecies)
	masters.PATCH("/animal-species/reorder", h.requirePermission(string(model.ResourceMasterAnimalSpecies), "edit"), h.ReorderAnimalSpecies)
	masters.GET("/animal-species/:id", h.requirePermission(string(model.ResourceMasterAnimalSpecies), "view"), h.GetAnimalSpecies)
	masters.PATCH("/animal-species/:id", h.requirePermission(string(model.ResourceMasterAnimalSpecies), "edit"), h.UpdateAnimalSpecies)
	masters.DELETE("/animal-species/:id", h.requirePermission(string(model.ResourceMasterAnimalSpecies), "delete"), h.DeleteAnimalSpecies)
}
