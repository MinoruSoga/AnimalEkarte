package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

type updateCompanyInput struct {
	Name               string `json:"name"`
	PostalCode         string `json:"postal_code"`
	Address            string `json:"address"`
	PhoneNumber        string `json:"phone_number"`
	FaxNumber          string `json:"fax_number"`
	Email              string `json:"email"`
	Website            string `json:"website"`
	DirectorName       string `json:"director_name"`
	RegistrationNumber string `json:"registration_number"`
	LogoURL            string `json:"logo_url"`
}

type updateClinicInput struct {
	Name               string `json:"name"`
	PostalCode         string `json:"postal_code"`
	Address            string `json:"address"`
	PhoneNumber        string `json:"phone_number"`
	FaxNumber          string `json:"fax_number"`
	RegistrationNumber string `json:"registration_number"`
	DirectorName       string `json:"director_name"`
	Email              string `json:"email"`
	Website            string `json:"website"`
	LogoURL            string `json:"logo_url"`
	IsActive           *bool  `json:"is_active"`
}

// GetCompany godoc
func (h *Handler) GetCompany(c *gin.Context) {
	company, err := h.svc.Clinic.GetCompany(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, company)
}

// UpdateCompany godoc
func (h *Handler) UpdateCompany(c *gin.Context) {
	var input updateCompanyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company := &model.Company{
		Name:               input.Name,
		PostalCode:         input.PostalCode,
		Address:            input.Address,
		PhoneNumber:        input.PhoneNumber,
		FaxNumber:          input.FaxNumber,
		Email:              input.Email,
		Website:            input.Website,
		DirectorName:       input.DirectorName,
		RegistrationNumber: input.RegistrationNumber,
		LogoURL:            input.LogoURL,
	}

	if err := h.svc.Clinic.UpdateCompany(c.Request.Context(), company); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, company)
}

// ListClinics godoc
func (h *Handler) ListClinics(c *gin.Context) {
	clinics, err := h.svc.Clinic.ListClinics(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinics)
}

// GetClinic godoc
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	clinic, err := h.svc.Clinic.GetClinicByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinic)
}

// UpdateClinic godoc
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input updateClinicInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clinic := &model.Clinic{
		Name:               input.Name,
		PostalCode:         input.PostalCode,
		Address:            input.Address,
		PhoneNumber:        input.PhoneNumber,
		FaxNumber:          input.FaxNumber,
		RegistrationNumber: input.RegistrationNumber,
		DirectorName:       input.DirectorName,
		Email:              input.Email,
		Website:            input.Website,
		LogoURL:            input.LogoURL,
	}
	if input.IsActive != nil {
		clinic.IsActive = *input.IsActive
	}

	result, err := h.svc.Clinic.UpdateClinic(c.Request.Context(), id, clinic)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

type createClinicRequest struct {
	Name               string `json:"name"                binding:"required"`
	PostalCode         string `json:"postal_code"`
	Address            string `json:"address"`
	PhoneNumber        string `json:"phone_number"`
	FaxNumber          string `json:"fax_number"`
	RegistrationNumber string `json:"registration_number"`
	DirectorName       string `json:"director_name"`
	Email              string `json:"email"`
	Website            string `json:"website"`
}

// CreateClinic godoc
func (h *Handler) CreateClinic(c *gin.Context) {
	var req createClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clinic := &model.Clinic{
		Name:               req.Name,
		PostalCode:         req.PostalCode,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		FaxNumber:          req.FaxNumber,
		RegistrationNumber: req.RegistrationNumber,
		DirectorName:       req.DirectorName,
		Email:              req.Email,
		Website:            req.Website,
		IsActive:           true,
	}
	result, err := h.svc.Clinic.CreateClinic(c.Request.Context(), clinic)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// DeleteClinic godoc
func (h *Handler) DeleteClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Clinic.DeleteClinic(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
