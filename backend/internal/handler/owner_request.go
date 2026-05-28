package handler

import (
	"net/url"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

type listOwnersQuery struct {
	Search string
}

func newListOwnersQuery(values url.Values) listOwnersQuery {
	return listOwnersQuery{Search: values.Get("search")}
}

// createPetForOwnerRequest は飼主登録時のペット入力バインド struct
type createPetForOwnerRequest struct {
	Name            string    `json:"name"              binding:"required"`
	AnimalSpeciesID uint64    `json:"animal_species_id" binding:"required"`
	NameKana        string    `json:"name_kana"`
	Breed           string    `json:"breed"`
	Color           string    `json:"color"`
	Gender          string    `json:"gender"            binding:"omitempty,oneof=male female unknown"`
	Status          string    `json:"status"            binding:"omitempty,oneof=alive deceased"`
	BirthDate       *jsonDate `json:"birth_date"`
	Weight          *float64  `json:"weight"`
	NeuteredDate    *jsonDate `json:"neutered_date"`
	AcquisitionType string    `json:"acquisition_type"`
	DangerLevel     string    `json:"danger_level"`
	Food            string    `json:"food"`
	Environment     string    `json:"environment"`
	InsuranceID     *uint64   `json:"insurance_id"`
	Remarks         string    `json:"remarks"`
}

func (r createPetForOwnerRequest) toServiceInput() service.CreatePetForOwnerInput {
	return service.CreatePetForOwnerInput{
		Name:            r.Name,
		AnimalSpeciesID: r.AnimalSpeciesID,
		PetNameKana:     r.NameKana,
		Breed:           r.Breed,
		Color:           r.Color,
		Gender:          r.Gender,
		Status:          r.Status,
		BirthDate:       jsonDatePtr(r.BirthDate),
		Weight:          r.Weight,
		NeuteredDate:    jsonDatePtr(r.NeuteredDate),
		AcquisitionType: r.AcquisitionType,
		DangerLevel:     r.DangerLevel,
		Food:            r.Food,
		Environment:     r.Environment,
		InsuranceID:     r.InsuranceID,
		Remarks:         r.Remarks,
	}
}

// createOwnerRequest は飼主作成のバインド struct
type createOwnerRequest struct {
	OwnerName      string                     `json:"owner_name"       binding:"required"`
	OwnerNameKana  string                     `json:"owner_name_kana"`
	BirthDate      *jsonDate                  `json:"birth_date"`
	Company        string                     `json:"company"`
	PostalCode     string                     `json:"postal_code"`
	Address1       string                     `json:"address1"`
	Address2       string                     `json:"address2"`
	HomePostalCode string                     `json:"home_postal_code"`
	HomeAddress1   string                     `json:"home_address1"`
	HomeAddress2   string                     `json:"home_address2"`
	Phone          string                     `json:"phone"`
	CompanyPhone   string                     `json:"company_phone"`
	Email          string                     `json:"email"`
	Remarks        string                     `json:"remarks"`
	IsDangerous    bool                       `json:"is_dangerous"`
	DiscountRate   float64                    `json:"discount_rate"`
	MembershipType string                     `json:"membership_type"  binding:"omitempty,oneof=non_member member deceased transferred"`
	Pets           []createPetForOwnerRequest `json:"pets"`
}

func (r createOwnerRequest) toServiceInput() service.CreateOwnerInput {
	pets := make([]service.CreatePetForOwnerInput, 0, len(r.Pets))
	for i := range r.Pets {
		pets = append(pets, r.Pets[i].toServiceInput())
	}

	return service.CreateOwnerInput{
		OwnerName:      r.OwnerName,
		OwnerNameKana:  r.OwnerNameKana,
		BirthDate:      jsonDatePtr(r.BirthDate),
		Company:        r.Company,
		PostalCode:     r.PostalCode,
		Address1:       r.Address1,
		Address2:       r.Address2,
		HomePostalCode: r.HomePostalCode,
		HomeAddress1:   r.HomeAddress1,
		HomeAddress2:   r.HomeAddress2,
		Phone:          r.Phone,
		CompanyPhone:   r.CompanyPhone,
		Email:          r.Email,
		Remarks:        r.Remarks,
		IsDangerous:    r.IsDangerous,
		DiscountRate:   r.DiscountRate,
		MembershipType: model.MembershipType(r.MembershipType),
		Pets:           pets,
	}
}

// updateOwnerRequest は飼主更新のバインド struct（全フィールドポインタ型）
type updateOwnerRequest struct {
	OwnerName      *string   `json:"owner_name"`
	OwnerNameKana  *string   `json:"owner_name_kana"`
	BirthDate      *jsonDate `json:"birth_date"`
	Company        *string   `json:"company"`
	PostalCode     *string   `json:"postal_code"`
	Address1       *string   `json:"address1"`
	Address2       *string   `json:"address2"`
	HomePostalCode *string   `json:"home_postal_code"`
	HomeAddress1   *string   `json:"home_address1"`
	HomeAddress2   *string   `json:"home_address2"`
	Phone          *string   `json:"phone"`
	CompanyPhone   *string   `json:"company_phone"`
	Email          *string   `json:"email"`
	Remarks        *string   `json:"remarks"`
	IsDangerous    *bool     `json:"is_dangerous"`
	DiscountRate   *float64  `json:"discount_rate"`
	MembershipType *string   `json:"membership_type"  binding:"omitempty,oneof=non_member member deceased transferred"`
}

func (r updateOwnerRequest) toServiceInput() *service.UpdateOwnerInput {
	var membershipType *model.MembershipType
	if r.MembershipType != nil {
		mt := model.MembershipType(*r.MembershipType)
		membershipType = &mt
	}

	return &service.UpdateOwnerInput{
		OwnerName:      r.OwnerName,
		OwnerNameKana:  r.OwnerNameKana,
		BirthDate:      jsonDatePtr(r.BirthDate),
		Company:        r.Company,
		PostalCode:     r.PostalCode,
		Address1:       r.Address1,
		Address2:       r.Address2,
		HomePostalCode: r.HomePostalCode,
		HomeAddress1:   r.HomeAddress1,
		HomeAddress2:   r.HomeAddress2,
		Phone:          r.Phone,
		CompanyPhone:   r.CompanyPhone,
		Email:          r.Email,
		Remarks:        r.Remarks,
		IsDangerous:    r.IsDangerous,
		DiscountRate:   r.DiscountRate,
		MembershipType: membershipType,
	}
}

// patchOwnerLineUserIDRequest は LINE User ID 連携リクエスト（BE-005）。nil で連携解除。
type patchOwnerLineUserIDRequest struct {
	LineUserID *string `json:"line_user_id"`
}

// patchOwnerDeliveryExclusionRequest は配信除外フラグ更新リクエスト（FEAT-381）。
type patchOwnerDeliveryExclusionRequest struct {
	Excluded bool    `json:"excluded"`
	Reason   *string `json:"reason"   binding:"omitempty,max=100"`
}

func (r patchOwnerDeliveryExclusionRequest) toServiceInput() service.UpdateDeliveryExclusionInput {
	return service.UpdateDeliveryExclusionInput{
		Excluded: r.Excluded,
		Reason:   r.Reason,
	}
}

// patchOwnerDeliveryCautionRequest は配信注意フラグ更新リクエスト（FEAT-381-2）。
type patchOwnerDeliveryCautionRequest struct {
	Caution bool   `json:"caution"`
	Reason  string `json:"reason" binding:"omitempty,max=100"`
}

func (r patchOwnerDeliveryCautionRequest) toServiceInput() service.UpdateDeliveryCautionInput {
	return service.UpdateDeliveryCautionInput{
		Caution: r.Caution,
		Reason:  r.Reason,
	}
}

// patchOwnerTransferStatusRequest は転院フラグ更新リクエスト（FEAT-381）。
type patchOwnerTransferStatusRequest struct {
	IsTransferred bool `json:"is_transferred"`
}

func (r patchOwnerTransferStatusRequest) toServiceInput() service.UpdateTransferStatusInput {
	return service.UpdateTransferStatusInput{
		IsTransferred: r.IsTransferred,
	}
}
