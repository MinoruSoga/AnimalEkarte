package handler

import "github.com/gin-gonic/gin"

// BE9-2E compatibility delegates; remove when central route registration moves in BE9-2F.
func (h *Handler) ListTrimmingCourseTypes(c *gin.Context) {
	h.trimmingDomainHandler().ListTrimmingCourseTypes(c)
}
func (h *Handler) GetTrimmingCourseType(c *gin.Context) {
	h.trimmingDomainHandler().GetTrimmingCourseType(c)
}
func (h *Handler) CreateTrimmingCourseType(c *gin.Context) {
	h.trimmingDomainHandler().CreateTrimmingCourseType(c)
}
func (h *Handler) UpdateTrimmingCourseType(c *gin.Context) {
	h.trimmingDomainHandler().UpdateTrimmingCourseType(c)
}
func (h *Handler) ReorderTrimmingCourseTypes(c *gin.Context) {
	h.trimmingDomainHandler().ReorderTrimmingCourseTypes(c)
}
func (h *Handler) DeleteTrimmingCourseType(c *gin.Context) {
	h.trimmingDomainHandler().DeleteTrimmingCourseType(c)
}
