package owner

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// SummaryResponse is the owner summary embedded by pet and other domain
// responses.
// BUG-374 同様: フロントの generated/models.Owner.name を期待するため json:"name" に統一。
type SummaryResponse struct {
	ID        uint64 `json:"id"`
	OwnerName string `json:"name"`
}

// PetAnimalSpeciesNested is the species embed on owner detail pets.
type PetAnimalSpeciesNested struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// PetInsuranceNested is the insurance embed on owner detail pets.
type PetInsuranceNested struct {
	ID           uint64  `json:"id"`
	Name         string  `json:"name"`
	CoverageRate float64 `json:"coverage_rate"`
	ContactPhone string  `json:"contact_phone"`
}

// PetInOwnerResponse is a pet row nested under OwnerResponse.
type PetInOwnerResponse struct {
	ID              uint64     `json:"id"`
	OwnerID         uint64     `json:"owner_id"`
	AnimalSpeciesID uint64     `json:"animal_species_id"`
	PetNumber       string     `json:"pet_number"`
	Name            string     `json:"name"`
	PetNameKana     string     `json:"pet_name_kana"`
	Gender          string     `json:"gender"`
	Status          string     `json:"status"`
	BirthDate       *time.Time `json:"birth_date,omitempty"`
	Breed           string     `json:"breed"`
	Color           string     `json:"color"`
	BloodType       *string    `json:"blood_type,omitempty"`
	MicrochipNumber *string    `json:"microchip_number,omitempty"`
	DangerLevel     string     `json:"danger_level"`
	Weight          *float64   `json:"weight,omitempty"`
	NeuteredDate    *time.Time `json:"neutered_date,omitempty"`
	AcquisitionType *string    `json:"acquisition_type,omitempty"`
	Food            string     `json:"food"`
	Environment     string     `json:"environment"`
	LastVisit       *time.Time `json:"last_visit,omitempty"`
	InsuranceID     *uint64    `json:"insurance_id,omitempty"`
	Remarks         string     `json:"remarks"`
	// DeceasedReason は含めない（セキュリティレビュー指摘: この構造体は未curationの
	// LIFF LinkLiffAccount 経路でも再利用されるため、スタッフ用の死因自由記述を飼主向け
	// レスポンスに載せてしまう。UI側でも死亡バナー表示に必要なのは DeceasedAt のみで
	// DeceasedReason の読み取り消費者は存在しない — 意図的に未追加）。
	DeceasedAt    *time.Time              `json:"deceased_at,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	AnimalSpecies *PetAnimalSpeciesNested `json:"animal_species,omitempty"`
	Insurance     *PetInsuranceNested     `json:"insurance,omitempty"`
}

// OwnerResponse is the owner detail/list HTTP response DTO (domain-owned wire).
// Field names use owner_name / owner_name_kana — not model.Owner's name / name_kana.
type OwnerResponse struct {
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
	DMPreference           *bool                `json:"dm_preference,omitempty"`
	Pets                   []PetInOwnerResponse `json:"pets"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
}

func toPetInOwnerResponse(p *model.Pet) PetInOwnerResponse {
	resp := PetInOwnerResponse{
		ID:              p.ID,
		OwnerID:         p.OwnerID,
		AnimalSpeciesID: p.AnimalSpeciesID,
		PetNumber:       p.PetNumber,
		Name:            p.Name,
		PetNameKana:     p.NameKana,
		Gender:          string(p.Gender),
		Status:          string(p.Status),
		BirthDate:       httpapi.LocalTimePtr(p.BirthDate),
		Breed:           p.Breed,
		Color:           p.Color,
		BloodType:       p.BloodType,
		MicrochipNumber: p.MicrochipNumber,
		DangerLevel:     string(p.DangerLevel),
		Weight:          p.Weight,
		NeuteredDate:    httpapi.LocalTimePtr(p.NeuteredDate),
		Food:            p.Food,
		Environment:     p.Environment,
		LastVisit:       httpapi.LocalTimePtr(p.LastVisit),
		InsuranceID:     p.InsuranceID,
		Remarks:         p.Remarks,
		DeceasedAt:      httpapi.LocalTimePtr(p.DeceasedAt),
		CreatedAt:       httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:       httpapi.LocalTime(p.UpdatedAt),
	}
	if p.AcquisitionType != nil {
		s := string(*p.AcquisitionType)
		resp.AcquisitionType = &s
	}
	if p.AnimalSpecies != nil {
		resp.AnimalSpecies = &PetAnimalSpeciesNested{
			ID:        p.AnimalSpecies.ID,
			Name:      p.AnimalSpecies.Name,
			SortOrder: p.AnimalSpecies.SortOrder,
		}
	}
	if p.Insurance != nil {
		resp.Insurance = &PetInsuranceNested{
			ID:           p.Insurance.ID,
			Name:         p.Insurance.Name,
			CoverageRate: float64(p.Insurance.CoverageRate),
			ContactPhone: p.Insurance.ContactPhone,
		}
	}
	return resp
}

func toOwnerResponse(o *model.Owner) OwnerResponse {
	pets := make([]PetInOwnerResponse, 0, len(o.Pets))
	for i := range o.Pets {
		pets = append(pets, toPetInOwnerResponse(&o.Pets[i]))
	}
	return OwnerResponse{
		ID:                     o.ID,
		ClinicID:               o.ClinicID,
		OwnerName:              o.Name,
		OwnerNameKana:          o.NameKana,
		BirthDate:              httpapi.LocalTimePtr(o.BirthDate),
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
		LineIDConfirmedAt:      httpapi.LocalTimePtr(o.LineIDConfirmedAt),
		LineIDConfirmedBy:      o.LineIDConfirmedBy,
		DeliveryExcluded:       o.DeliveryExcluded,
		DeliveryExcludedReason: o.DeliveryExcludedReason,
		DeliveryCaution:        o.DeliveryCaution,
		DeliveryCautionReason:  o.DeliveryCautionReason,
		IsTransferred:          o.IsTransferred,
		TransferAt:             httpapi.LocalTimePtr(o.TransferAt),
		DMPreference:           o.DMPreference,
		Pets:                   pets,
		CreatedAt:              httpapi.LocalTime(o.CreatedAt),
		UpdatedAt:              httpapi.LocalTime(o.UpdatedAt),
	}
}
