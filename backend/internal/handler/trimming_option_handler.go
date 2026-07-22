package handler

import "github.com/gin-gonic/gin"

// BE9-2E compatibility delegates; remove when central route registration moves in BE9-2F.
func (h *Handler) ListTrimmingOptions(c *gin.Context) {
	h.trimmingDomainHandler().ListTrimmingOptions(c)
}
func (h *Handler) GetTrimmingOption(c *gin.Context) { h.trimmingDomainHandler().GetTrimmingOption(c) }
func (h *Handler) CreateTrimmingOption(c *gin.Context) {
	h.trimmingDomainHandler().CreateTrimmingOption(c)
}
func (h *Handler) UpdateTrimmingOption(c *gin.Context) {
	h.trimmingDomainHandler().UpdateTrimmingOption(c)
}
func (h *Handler) DeleteTrimmingOption(c *gin.Context) {
	h.trimmingDomainHandler().DeleteTrimmingOption(c)
}
func (h *Handler) ReorderTrimmingOptions(c *gin.Context) {
	h.trimmingDomainHandler().ReorderTrimmingOptions(c)
}
