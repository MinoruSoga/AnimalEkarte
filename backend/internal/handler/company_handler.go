package handler

import "github.com/gin-gonic/gin"

func (h *Handler) GetCompany(c *gin.Context) {
	h.clinicDomainHandler().GetCompany(c)
}

func (h *Handler) UpdateCompany(c *gin.Context) {
	h.clinicDomainHandler().UpdateCompany(c)
}

func (h *Handler) RegisterCompanyRoutes(rg *gin.RouterGroup) {
	h.clinicDomainHandler().RegisterCompanyRoutes(rg)
}
