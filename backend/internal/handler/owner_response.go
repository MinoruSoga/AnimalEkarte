package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// ownerSummaryResponse はネストされたレスポンスで使用するオーナーの要約型。
// BUG-374 同様: フロントの generated/models.Owner.name を期待するため json:"name" に統一。
type ownerSummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// toOwnerSummary は *model.Owner を *ownerSummaryResponse に変換する。nilの場合はnilを返す。
func toOwnerSummary(o *model.Owner) *ownerSummaryResponse {
	if o == nil {
		return nil
	}
	return &ownerSummaryResponse{
		ID:        o.ID,
		OwnerName: o.Name,
	}
}

type petInOwnerResponse struct {
	ID              uint64                  `json:"id"`
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
	DangerLevel     string                  `json:"danger_level"`
	Weight          *float64                `json:"weight,omitempty"`
	NeuteredDate    *time.Time              `json:"neutered_date,omitempty"`
	AcquisitionType *string                 `json:"acquisition_type,omitempty"`
	Food            string                  `json:"food"`
	Environment     string                  `json:"environment"`
	LastVisit       *time.Time              `json:"last_visit,omitempty"`
	InsuranceID     *uint64                 `json:"insurance_id,omitempty"`
	Remarks         string                  `json:"remarks"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	AnimalSpecies   *petAnimalSpeciesNested `json:"animal_species,omitempty"`
	Insurance       *petInsuranceNested     `json:"insurance,omitempty"`
}

type ownerResponse struct {
	ID uint64 `json:"id"`
	// ClinicID は所属医院 (#86 拠点横断一覧で医院名表示に使用)
	ClinicID               uint64               `json:"clinic_id"`
	OwnerName              string               `json:"owner_name"`
	OwnerNameKana          string               `json:"owner_name_kana"`
	BirthDate              *time.Time           `json:"birth_date,omitempty"`
	Company                string               `json:"company"`
	PostalCode             string               `json:"postal_code"`
	Address1               string               `json:"address1"`
	Address2               string               `json:"address2"`
	HomePostalCode         string               `json:"home_postal_code"`
	HomeAddress1           string               `json:"home_address1"`
	HomeAddress2           string               `json:"home_address2"`
	Phone                  string               `json:"phone"`
	CompanyPhone           string               `json:"company_phone"`
	Email                  string               `json:"email"`
	Remarks                string               `json:"remarks"`
	IsDangerous            bool                 `json:"is_dangerous"`
	DiscountRate           float64              `json:"discount_rate"`
	MembershipType         string               `json:"membership_type"`
	LineIDConfirmedAt      *time.Time           `json:"line_id_confirmed_at,omitempty"`
	LineIDConfirmedBy      *uint64              `json:"line_id_confirmed_by,omitempty"`
	DeliveryExcluded       bool                 `json:"delivery_excluded"`
	DeliveryExcludedReason *string              `json:"delivery_excluded_reason,omitempty"`
	DeliveryCaution        bool                 `json:"delivery_caution"`
	DeliveryCautionReason  *string              `json:"delivery_caution_reason,omitempty"`
	IsTransferred          bool                 `json:"is_transferred"`
	TransferAt             *time.Time           `json:"transfer_at,omitempty"`
	Pets                   []petInOwnerResponse `json:"pets"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
}

func toPetInOwnerResponse(p *model.Pet) petInOwnerResponse {
	resp := petInOwnerResponse{
		ID:              p.ID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		BirthDate:       localTimePtr(p.BirthDate),
		Breed:           p.Breed,
		Color:           p.Color,
		DangerLevel:     string(p.DangerLevel),
		Weight:          p.Weight,
		NeuteredDate:    localTimePtr(p.NeuteredDate),
		Food:            p.Food,
		Environment:     p.Environment,
		LastVisit:       localTimePtr(p.LastVisit),
		InsuranceID:     p.InsuranceID,
		Remarks:         p.Remarks,
		CreatedAt:       localTime(p.CreatedAt),
		UpdatedAt:       localTime(p.UpdatedAt),
	}
	if p.AcquisitionType != nil {
		s := string(*p.AcquisitionType)
		resp.AcquisitionType = &s
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

func toOwnerResponse(o *model.Owner) ownerResponse {
	pets := make([]petInOwnerResponse, 0, len(o.Pets))
	for i := range o.Pets {
		pets = append(pets, toPetInOwnerResponse(&o.Pets[i]))
	}
	return ownerResponse{
		ID:                     o.ID,
		ClinicID:               o.ClinicID,
		OwnerName:              o.Name,
		OwnerNameKana:          o.NameKana,
		BirthDate:              localTimePtr(o.BirthDate),
		Company:                o.Company,
		PostalCode:             o.PostalCode,
		Address1:               o.Address1,
		Address2:               o.Address2,
		HomePostalCode:         o.HomePostalCode,
		HomeAddress1:           o.HomeAddress1,
		HomeAddress2:           o.HomeAddress2,
		Phone:                  o.Phone,
		CompanyPhone:           o.CompanyPhone,
		Email:                  o.Email,
		Remarks:                o.Remarks,
		IsDangerous:            o.IsDangerous,
		DiscountRate:           o.DiscountRate,
		MembershipType:         string(o.MembershipType),
		LineIDConfirmedAt:      localTimePtr(o.LineIDConfirmedAt),
		LineIDConfirmedBy:      o.LineIDConfirmedBy,
		DeliveryExcluded:       o.DeliveryExcluded,
		DeliveryExcludedReason: o.DeliveryExcludedReason,
		DeliveryCaution:        o.DeliveryCaution,
		DeliveryCautionReason:  o.DeliveryCautionReason,
		IsTransferred:          o.IsTransferred,
		TransferAt:             localTimePtr(o.TransferAt),
		Pets:                   pets,
		CreatedAt:              localTime(o.CreatedAt),
		UpdatedAt:              localTime(o.UpdatedAt),
	}
}
