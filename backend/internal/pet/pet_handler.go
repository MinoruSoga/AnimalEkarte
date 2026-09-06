package pet

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type ownerReportAnimalSpeciesResponse struct {
	Name string `json:"name"`
}

type ownerReportInsuranceResponse struct {
	Name         string `json:"name"`
	CoverageRate int    `json:"coverage_rate"`
}

type ownerReportResponse struct {
	ID              uint64                            `json:"id"`
	Name            string                            `json:"name"`
	PetNameKana     string                            `json:"pet_name_kana"`
	Gender          string                            `json:"gender"`
	Status          string                            `json:"status"`
	BirthDate       *time.Time                        `json:"birth_date"`
	Breed           string                            `json:"breed"`
	Color           string                            `json:"color"`
	BloodType       *string                           `json:"blood_type"`
	MicrochipNumber *string                           `json:"microchip_number"`
	Weight          *float64                          `json:"weight"`
	NeuteredDate    *time.Time                        `json:"neutered_date"`
	AcquisitionType *string                           `json:"acquisition_type"`
	Food            string                            `json:"food"`
	Environment     string                            `json:"environment"`
	LastVisit       *time.Time                        `json:"last_visit"`
	Remarks         string                            `json:"remarks"`
	AnimalSpecies   *ownerReportAnimalSpeciesResponse `json:"animal_species"`
	Insurance       *ownerReportInsuranceResponse     `json:"insurance"`
}

type ownerReportPetsResponse struct {
	Data []ownerReportResponse `json:"data"`
}

func toOwnerReportResponse(p *model.Pet) ownerReportResponse {
	var acquisitionType *string
	if p.AcquisitionType != nil {
		value := string(*p.AcquisitionType)
		acquisitionType = &value
	}

	response := ownerReportResponse{
		ID:              p.ID,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		BirthDate:       httpapi.LocalTimePtr(p.BirthDate),
		Breed:           p.Breed,
		Color:           p.Color,
		BloodType:       p.BloodType,
		MicrochipNumber: p.MicrochipNumber,
		Weight:          p.Weight,
		NeuteredDate:    httpapi.LocalTimePtr(p.NeuteredDate),
		AcquisitionType: acquisitionType,
		Food:            p.Food,
		Environment:     p.Environment,
		LastVisit:       httpapi.LocalTimePtr(p.LastVisit),
		Remarks:         p.Remarks,
	}
	if p.AnimalSpecies != nil {
		response.AnimalSpecies = &ownerReportAnimalSpeciesResponse{Name: p.AnimalSpecies.Name}
	}
	if p.Insurance != nil {
		response.Insurance = &ownerReportInsuranceResponse{
			Name:         p.Insurance.Name,
			CoverageRate: p.Insurance.CoverageRate,
		}
	}
	return response
}

// ListPets godoc
// #86: 拠点横断一覧 — 所属かつ owners:view を持つ医院だけをスコープにする
// （OwnersList.tsx が useClinicScope 経由でこの一覧を消費するため owners API と同一パターンを維持する）。
func (h *Handler) ListPets(c *gin.Context) {
	clinicIDs, ok := httpapi.ResolveListClinicIDsForPermission(
		c,
		string(model.ResourceOwners),
		"view",
	)
	if !ok {
		return
	}
	page, limit, err := httpapi.ParsePagination(c)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	filters, err := newListPetQuery(c.Request.URL.Query()).toServiceFilters()
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	pets, total, err := h.pets.List(c.Request.Context(), clinicIDs, filters, page, limit)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpapi.NewPaginatedResponse(httpapi.MapSlice(pets, toPetListResponse), total, page, limit))
}

// ListOwnerReportPets godoc
func (h *Handler) ListOwnerReportPets(c *gin.Context) {
	clinicIDs, ok := httpapi.ResolveAllClinicIDsForPermission(
		c,
		string(model.ResourceOwners),
		"view",
	)
	if !ok {
		return
	}
	ownerID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	pets, err := h.pets.ListOwnerReportPets(c.Request.Context(), clinicIDs, ownerID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ownerReportPetsResponse{
		Data: httpapi.MapSlice(pets, toOwnerReportResponse),
	})
}

// GetPet godoc
func (h *Handler) GetPet(c *gin.Context) {
	// #86: 詳細画面の拠点横断閲覧 — 所属かつ owners:view を持つ医院だけをスコープにする
	clinicIDs, ok := httpapi.ResolveAllClinicIDsForPermission(
		c,
		string(model.ResourceOwners),
		"view",
	)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	pet, err := h.pets.GetByIDForClinics(c.Request.Context(), clinicIDs, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(pet))
}

// GetPetFirstVisit godoc
// GET /pets/:id/first-visit (#158 飼主レポート: 初診日 = 最古の有効カルテ date)
// 値は medical_records から導出する派生値で、捏造しない（カルテ無し = null）。
// clinic 隔離・論理削除除外は service/repository が担保する。
func (h *Handler) GetPetFirstVisit(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	petID, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	date, err := h.pets.GetFirstVisitDate(c.Request.Context(), clinicID, petID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toPetFirstVisitResponse(date))
}

// CreatePet godoc
func (h *Handler) CreatePet(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var req createPetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput()
	pet, err := h.pets.Create(c.Request.Context(), clinicID, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/pets/%d", pet.ID))
	c.JSON(http.StatusCreated, toResponse(pet))
}

// UpdatePet godoc
func (h *Handler) UpdatePet(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req updatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	input := req.toServiceInput()
	pet, err := h.pets.Update(c.Request.Context(), clinicID, id, input)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(pet))
}

// DeletePet godoc
func (h *Handler) DeletePet(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.pets.Delete(c.Request.Context(), clinicID, id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
