package handler

import (
	"github.com/gin-gonic/gin"

	trimmingdomain "github.com/animal-ekarte/backend/internal/trimming"
)

// trimmingDomainHandler is the BE9-2E compatibility bridge for central route registration.
// Remove these delegates when the handler aggregator is retired in BE9-2F.
func (h *Handler) trimmingDomainHandler() *trimmingdomain.Handler {
	return trimmingdomain.NewHandler(
		h.svc.Trimming,
		h.svc.TrimmingCourse,
		h.svc.TrimmingCourseType,
		h.svc.TrimmingOption,
	)
}

func (h *Handler) ListTrimmings(c *gin.Context)  { h.trimmingDomainHandler().ListTrimmings(c) }
func (h *Handler) GetTrimming(c *gin.Context)    { h.trimmingDomainHandler().GetTrimming(c) }
func (h *Handler) CreateTrimming(c *gin.Context) { h.trimmingDomainHandler().CreateTrimming(c) }
func (h *Handler) UpdateTrimming(c *gin.Context) { h.trimmingDomainHandler().UpdateTrimming(c) }
func (h *Handler) DeleteTrimming(c *gin.Context) { h.trimmingDomainHandler().DeleteTrimming(c) }
