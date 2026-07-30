package identitylink

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts identity-link routes under /identity-links.
// GET = view, link/unlink = edit (DEC-15 manage maps to edit action string).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	viewRes, viewAct := identityLinksView()
	editRes, editAct := identityLinksEdit()

	g := rg.Group("/identity-links")
	g.GET("/owners/search", h.requirePermission(viewRes, viewAct), h.SearchOwners)
	g.GET("/pets/search", h.requirePermission(viewRes, viewAct), h.SearchPets)

	g.GET("/owner-groups/:id", h.requirePermission(viewRes, viewAct), h.GetOwnerGroup)
	g.GET("/owners/:clinic_id/:owner_id/group", h.requirePermission(viewRes, viewAct), h.FindOwnerGroupByMember)
	g.POST("/owner-groups", h.requirePermission(editRes, editAct), h.CreateOwnerGroup)
	g.POST("/owner-groups/:id/members", h.requirePermission(editRes, editAct), h.AddOwnerMembers)
	g.DELETE("/owner-groups/:id/members", h.requirePermission(editRes, editAct), h.UnlinkOwnerMember)

	g.GET("/pet-groups/:id", h.requirePermission(viewRes, viewAct), h.GetPetGroup)
	g.GET("/pets/:clinic_id/:pet_id/group", h.requirePermission(viewRes, viewAct), h.FindPetGroupByMember)
	g.POST("/pet-groups", h.requirePermission(editRes, editAct), h.CreatePetGroup)
	g.POST("/pet-groups/:id/members", h.requirePermission(editRes, editAct), h.AddPetMembers)
	g.DELETE("/pet-groups/:id/members", h.requirePermission(editRes, editAct), h.UnlinkPetMember)

	g.GET(
		"/pets/:clinic_id/:pet_id/treatment-history",
		h.requirePermission(viewRes, viewAct),
		h.ListLinkedTreatmentHistory,
	)
}
