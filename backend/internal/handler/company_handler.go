package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/service"
)

// GetCompany は法人情報を返す（シングルトン）
func (h *Handler) GetCompany(c *gin.Context) {
	company, err := h.svc.Company.Get(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCompanyResponse(company))
}

// UpdateCompany は法人情報を部分更新する
func (h *Handler) UpdateCompany(c *gin.Context) {
	var req updateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	company, err := h.svc.Company.Update(c.Request.Context(), &service.UpdateCompanyInput{
		Name:               req.Name,
		PostalCode:         req.PostalCode,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		FaxNumber:          req.FaxNumber,
		Email:              req.Email,
		Website:            req.Website,
		DirectorName:       req.DirectorName,
		RegistrationNumber: req.RegistrationNumber,
		LogoURL:            req.LogoURL,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCompanyResponse(company))
}

// RegisterCompanyRoutes は法人情報関連のルートを登録する
func (h *Handler) RegisterCompanyRoutes(rg *gin.RouterGroup) {
	rg.GET("/company", h.GetCompany)
	rg.PATCH("/company", h.UpdateCompany)
}
