package trimming

// handlerServices is the domain-local, typed handler dependency set.
type handlerServices struct {
	Trimming           TrimmingService
	TrimmingCourse     TrimmingCourseService
	TrimmingCourseType TrimmingCourseTypeService
	TrimmingOption     TrimmingOptionService
}

// Handler owns the trimming HTTP behavior. Route registration remains in the legacy handler
// aggregator until BE9-2F so the current canonical integration HOLD does not touch central files.
type Handler struct {
	svc *handlerServices
}

// NewHandler constructs the trimming HTTP handler used by the BE9-2F compatibility facade.
func NewHandler(
	trimmingService TrimmingService,
	courseService TrimmingCourseService,
	courseTypeService TrimmingCourseTypeService,
	optionService TrimmingOptionService,
) *Handler {
	return &Handler{svc: &handlerServices{
		Trimming:           trimmingService,
		TrimmingCourse:     courseService,
		TrimmingCourseType: courseTypeService,
		TrimmingOption:     optionService,
	}}
}
