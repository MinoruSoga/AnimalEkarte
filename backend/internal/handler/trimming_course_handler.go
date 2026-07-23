package handler

import "github.com/gin-gonic/gin"

// BE9-2E compatibility delegates; remove when central route registration moves in BE9-2F.
func (h *Handler) ListTrimmingCourses(c *gin.Context) {
	h.trimmingDomainHandler().ListTrimmingCourses(c)
}
func (h *Handler) GetTrimmingCourse(c *gin.Context) { h.trimmingDomainHandler().GetTrimmingCourse(c) }
func (h *Handler) CreateTrimmingCourse(c *gin.Context) {
	h.trimmingDomainHandler().CreateTrimmingCourse(c)
}
func (h *Handler) UpdateTrimmingCourse(c *gin.Context) {
	h.trimmingDomainHandler().UpdateTrimmingCourse(c)
}
func (h *Handler) DeleteTrimmingCourse(c *gin.Context) {
	h.trimmingDomainHandler().DeleteTrimmingCourse(c)
}
func (h *Handler) ReorderTrimmingCourses(c *gin.Context) {
	h.trimmingDomainHandler().ReorderTrimmingCourses(c)
}
