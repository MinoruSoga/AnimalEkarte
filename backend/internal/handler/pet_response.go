package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type petAnimalSpeciesNested struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type petInsuranceNested struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	CoverageRate int    `json:"coverage_rate"`
	ContactPhone string `json:"contact_phone"`
}

type petOwnerNested struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type petResponse struct {
	ID              uint64                  `json:"id"`
	ClinicID        uint64                  `json:"clinic_id"`
	OwnerID         uint64                  `json:"owner_id"`
	AnimalSpeciesID uint64                  `json:"animal_species_id"`
	PetNumber       string                  `json:"pet_number"`
	Name            string                  `json:"name"`
	PetNameKana     string                  `json:"pet_name_kana"`
	Gender          string                  `json:"gender"`
	Status          string                  `json:"status"`
	BirthDate       *time.Time              `json:"birth_date,omitempty"`
	Breed           string                  `json:"breed"`
	Color           string                  `json:"color"`
	Weight          *float64                `json:"weight,omitempty"`
	NeuteredDate    *time.Time              `json:"neutered_date,omitempty"`
	AcquisitionType *string                 `json:"acquisition_type,omitempty"`
	DangerLevel     string                  `json:"danger_level"`
	Food            string                  `json:"food"`
	Environment     string                  `json:"environment"`
	Phone           string                  `json:"phone"`
	LastVisit       *time.Time              `json:"last_visit,omitempty"`
	InsuranceID     *uint64                 `json:"insurance_id,omitempty"`
	Remarks         string                  `json:"remarks"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	Owner           *petOwnerNested         `json:"owner,omitempty"`
	AnimalSpecies   *petAnimalSpeciesNested `json:"animal_species,omitempty"`
	Insurance       *petInsuranceNested     `json:"insurance,omitempty"`
}

// petListResponse はリスト表示に必要な最小限フィールドのみ返す（GET /v1/pets 専用）
type petListResponse struct {
	ID              uint64                  `json:"id"`
	OwnerID         uint64                  `json:"owner_id"`
	AnimalSpeciesID uint64                  `json:"animal_species_id"`
	PetNumber       string                  `json:"pet_number"`
	Name            string                  `json:"name"`
	PetNameKana     string                  `json:"pet_name_kana"`
	Gender          string                  `json:"gender"`
	Status          string                  `json:"status"`
	BirthDate       *time.Time              `json:"birth_date,omitempty"`
	Weight          *float64                `json:"weight,omitempty"`
	Environment     string                  `json:"environment"`
	LastVisit       *time.Time              `json:"last_visit,omitempty"`
	Owner           *petOwnerNested         `json:"owner,omitempty"`
	AnimalSpecies   *petAnimalSpeciesNested `json:"animal_species,omitempty"`
}

func toPetListResponse(p *model.Pet) petListResponse {
	resp := petListResponse{
		ID:              p.ID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		BirthDate:       p.BirthDate,
		Weight:          p.Weight,
		Environment:     p.Environment,
		LastVisit:       p.LastVisit,
	}
	if p.Owner != nil {
		resp.Owner = &petOwnerNested{
			ID:    p.Owner.ID,
			Name:  p.Owner.Name,
			Phone: p.Owner.Phone,
		}
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &petAnimalSpeciesNested{
			ID:        p.AnimalSpecies.ID,
			Name:      p.AnimalSpecies.Name,
			SortOrder: p.AnimalSpecies.SortOrder,
		}
	}
	return resp
}

// petSummaryResponse はネストされたレスポンスで使用するペットの要約型
type petSummaryResponse struct {
	ID            uint64                        `json:"id"`
	Name          string                        `json:"name"`
	AnimalSpecies *animalSpeciesSummaryResponse `json:"animal_species,omitempty"`
}

// animalSpeciesSummaryResponse はペット要約内で使用する種別の要約型
type animalSpeciesSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toPetSummary は *model.Pet を *petSummaryResponse に変換する。nilの場合はnilを返す。
func toPetSummary(p *model.Pet) *petSummaryResponse {
	if p == nil {
		return nil
	}
	resp := &petSummaryResponse{
		ID:   p.ID,
		Name: p.Name,
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &animalSpeciesSummaryResponse{
			ID:   p.AnimalSpecies.ID,
			Name: p.AnimalSpecies.Name,
		}
	}
	return resp
}

func toPetResponse(p *model.Pet) petResponse {
	var acquisitionType *string
	if p.AcquisitionType != nil {
		s := string(*p.AcquisitionType)
		acquisitionType = &s
	}
	resp := petResponse{
		ID:              p.ID,
		ClinicID:        p.ClinicID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		BirthDate:       p.BirthDate,
		Breed:           p.Breed,
		Color:           p.Color,
		Weight:          p.Weight,
		NeuteredDate:    p.NeuteredDate,
		AcquisitionType: acquisitionType,
		DangerLevel:     string(p.DangerLevel),
		Food:            p.Food,
		Environment:     p.Environment,
		Phone:           p.Phone,
		LastVisit:       p.LastVisit,
		InsuranceID:     p.InsuranceID,
		Remarks:         p.Remarks,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
	if p.Owner != nil {
		resp.Owner = &petOwnerNested{
			ID:    p.Owner.ID,
			Name:  p.Owner.Name,
			Phone: p.Owner.Phone,
		}
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &petAnimalSpeciesNested{
			ID:        p.AnimalSpecies.ID,
			Name:      p.AnimalSpecies.Name,
			SortOrder: p.AnimalSpecies.SortOrder,
		}
	}
	if p.Insurance != nil {
		resp.Insurance = &petInsuranceNested{
			ID:           p.Insurance.ID,
			Name:         p.Insurance.Name,
			CoverageRate: p.Insurance.CoverageRate,
			ContactPhone: p.Insurance.ContactPhone,
		}
	}
	return resp
}
