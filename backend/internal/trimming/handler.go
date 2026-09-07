package trimming

import "github.com/gin-gonic/gin"

// PermissionMiddleware is trimming's consumer-side RBAC middleware view.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// handlerServices is the domain-local, typed handler dependency set.
type handlerServices struct {
	Trimming           Service
	TrimmingCourse     TrimmingCourseService
	TrimmingCourseType TrimmingCourseTypeService
	TrimmingOption     TrimmingOptionService
}

// Handler owns the trimming HTTP behavior and route registration.
type Handler struct {
	svc               *handlerServices
	requirePermission PermissionMiddleware
}

// NewHandler constructs the compatibility HTTP handler. Direct route composition uses
// NewHandlerWithPermission.
func NewHandler(
	trimmingService Service,
	courseService TrimmingCourseService,
	courseTypeService TrimmingCourseTypeService,
	optionService TrimmingOptionService,
) *Handler {
	return NewHandlerWithPermission(
		trimmingService,
		courseService,
		courseTypeService,
		optionService,
		nil,
	)
}

// NewHandlerWithPermission constructs the domain-owned HTTP boundary.
func NewHandlerWithPermission(
	trimmingService Service,
	courseService TrimmingCourseService,
	courseTypeService TrimmingCourseTypeService,
	optionService TrimmingOptionService,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		svc: &handlerServices{
			Trimming:           trimmingService,
			TrimmingCourse:     courseService,
			TrimmingCourseType: courseTypeService,
			TrimmingOption:     optionService,
		},
		requirePermission: requirePermission,
	}
}
