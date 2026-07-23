package handler

import "github.com/gin-gonic/gin"

func (h *Handler) ListClinicHolidays(c *gin.Context) {
	h.clinicDomainHandler().ListClinicHolidays(c)
}

func (h *Handler) SetClinicHoliday(c *gin.Context) {
	h.clinicDomainHandler().SetClinicHoliday(c)
}

func (h *Handler) DeleteClinicHoliday(c *gin.Context) {
	h.clinicDomainHandler().DeleteClinicHoliday(c)
}

func (h *Handler) RegisterClinicHolidayRoutes(rg *gin.RouterGroup) {
	h.clinicDomainHandler().RegisterClinicHolidayRoutes(rg)
}
