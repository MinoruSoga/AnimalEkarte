package trimming

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// RegisterRoutes registers trimming's 23 authenticated legacy-compatible routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	trimmings := rg.Group("/trimmings")
	trimmings.GET("", h.requirePermission(string(model.ResourceTrimming), "view"), h.ListTrimmings)
	trimmings.GET("/:id", h.requirePermission(string(model.ResourceTrimming), "view"), h.GetTrimming)
	trimmings.POST("", h.requirePermission(string(model.ResourceTrimming), "create"), h.CreateTrimming)
	trimmings.PATCH("/:id", h.requirePermission(string(model.ResourceTrimming), "edit"), h.UpdateTrimming)
	trimmings.DELETE("/:id", h.requirePermission(string(model.ResourceTrimming), "delete"), h.DeleteTrimming)

	masters := rg.Group("/masters")
	masters.GET("/trimming-courses", h.requirePermission(string(model.ResourceMasterTrimming), "view"), h.ListTrimmingCourses)
	masters.POST("/trimming-courses", h.requirePermission(string(model.ResourceMasterTrimming), "create"), h.CreateTrimmingCourse)
	masters.PATCH("/trimming-courses/reorder", h.requirePermission(string(model.ResourceMasterTrimming), "edit"), h.ReorderTrimmingCourses)
	masters.GET("/trimming-courses/:id", h.requirePermission(string(model.ResourceMasterTrimming), "view"), h.GetTrimmingCourse)
	masters.PATCH("/trimming-courses/:id", h.requirePermission(string(model.ResourceMasterTrimming), "edit"), h.UpdateTrimmingCourse)
	masters.DELETE("/trimming-courses/:id", h.requirePermission(string(model.ResourceMasterTrimming), "delete"), h.DeleteTrimmingCourse)

	masters.GET("/trimming-options", h.requirePermission(string(model.ResourceMasterTrimming), "view"), h.ListTrimmingOptions)
	masters.POST("/trimming-options", h.requirePermission(string(model.ResourceMasterTrimming), "create"), h.CreateTrimmingOption)
	masters.PATCH("/trimming-options/reorder", h.requirePermission(string(model.ResourceMasterTrimming), "edit"), h.ReorderTrimmingOptions)
	masters.GET("/trimming-options/:id", h.requirePermission(string(model.ResourceMasterTrimming), "view"), h.GetTrimmingOption)
	masters.PATCH("/trimming-options/:id", h.requirePermission(string(model.ResourceMasterTrimming), "edit"), h.UpdateTrimmingOption)
	masters.DELETE("/trimming-options/:id", h.requirePermission(string(model.ResourceMasterTrimming), "delete"), h.DeleteTrimmingOption)

	masters.GET("/trimming-course-types", h.requirePermission(string(model.ResourceMasterTrimming), "view"), h.ListTrimmingCourseTypes)
	masters.POST("/trimming-course-types", h.requirePermission(string(model.ResourceMasterTrimming), "create"), h.CreateTrimmingCourseType)
	masters.PATCH("/trimming-course-types/reorder", h.requirePermission(string(model.ResourceMasterTrimming), "edit"), h.ReorderTrimmingCourseTypes)
	masters.GET("/trimming-course-types/:id", h.requirePermission(string(model.ResourceMasterTrimming), "view"), h.GetTrimmingCourseType)
	masters.PATCH("/trimming-course-types/:id", h.requirePermission(string(model.ResourceMasterTrimming), "edit"), h.UpdateTrimmingCourseType)
	masters.DELETE("/trimming-course-types/:id", h.requirePermission(string(model.ResourceMasterTrimming), "delete"), h.DeleteTrimmingCourseType)
}
