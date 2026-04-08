package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	company, err := h.svc.Company.Update(c.Request.Context(), &service.UpdateCompanyInput{
		Name:                      req.Name,
		PostalCode:                req.PostalCode,
		Address:                   req.Address,
		PhoneNumber:               req.PhoneNumber,
		FaxNumber:                 req.FaxNumber,
		Email:                     req.Email,
		Website:                   req.Website,
		DirectorName:              req.DirectorName,
		RegistrationNumber:        req.RegistrationNumber,
		InvoiceRegistrationNumber: req.InvoiceRegistrationNumber,
		LogoURL:                   req.LogoURL,
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
	rg.PATCH("/company", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateCompany)
}
