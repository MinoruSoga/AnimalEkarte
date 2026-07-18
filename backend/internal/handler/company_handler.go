package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	company, err := h.svc.Company.Update(c.Request.Context(), req.toServiceInput())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toCompanyResponse(company))
}

// RegisterCompanyRoutes は法人情報関連のルートを登録する
func (h *Handler) RegisterCompanyRoutes(rg *gin.RouterGroup) {
	rg.GET("/company", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetCompany)
	rg.PATCH("/company", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateCompany)
}
