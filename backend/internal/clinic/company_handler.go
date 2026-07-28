package clinic

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// GetCompany は法人情報を返す（シングルトン）
func (h *Handler) GetCompany(c *gin.Context) {
	company, err := h.companySvc.Get(c.Request.Context())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToCompanyResponse(company))
}

// UpdateCompany は法人情報を部分更新する
func (h *Handler) UpdateCompany(c *gin.Context) {
	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	company, err := h.companySvc.Update(c.Request.Context(), req.ToServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToCompanyResponse(company))
}

// RegisterCompanyRoutes は法人情報関連のルートを登録する
func (h *Handler) RegisterCompanyRoutes(rg *gin.RouterGroup) {
	rg.GET("/company", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.GetCompany)
	rg.PATCH("/company", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateCompany)
}
