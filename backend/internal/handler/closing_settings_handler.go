package handler

import "github.com/gin-gonic/gin"

func (h *Handler) GetClosingSettings(c *gin.Context) {
	h.clinicDomainHandler().GetClosingSettings(c)
}

func (h *Handler) UpdateClosingSettings(c *gin.Context) {
	h.clinicDomainHandler().UpdateClosingSettings(c)
}

func (h *Handler) ListSpecialPeriods(c *gin.Context) {
	h.clinicDomainHandler().ListSpecialPeriods(c)
}

func (h *Handler) CreateSpecialPeriod(c *gin.Context) {
	h.clinicDomainHandler().CreateSpecialPeriod(c)
}

func (h *Handler) UpdateSpecialPeriod(c *gin.Context) {
	h.clinicDomainHandler().UpdateSpecialPeriod(c)
}

func (h *Handler) DeleteSpecialPeriod(c *gin.Context) {
	h.clinicDomainHandler().DeleteSpecialPeriod(c)
}

func (h *Handler) RegisterClosingSettingsRoutes(rg *gin.RouterGroup) {
	h.clinicDomainHandler().RegisterClosingSettingsRoutes(rg)
}
