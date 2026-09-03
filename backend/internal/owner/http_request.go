package owner

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type listOwnersQuery struct {
	Search string
}

// nullableBoolRequestField は PATCH DTO で「未送信」と「JSON null」を区別する。
// Go 標準の **bool だけでは encoding/json が null と未送信の両方を nil にするため、
// set フラグでフィールド存在を保持し、service 層の **bool（nil / &nil / &&value）へ変換する。
type nullableBoolRequestField struct {
	set   bool
	value *bool
}

func (f *nullableBoolRequestField) UnmarshalJSON(data []byte) error {
	f.set = true
	if string(data) == "null" {
		f.value = nil
		return nil
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		// Go内部のエラー文字列を漏洩させないため、汎用メッセージを返す
		return apperrors.WrapInvalidInput("真偽値の形式が正しくありません（true または false を指定してください）")
	}
	f.value = &value
	return nil
}

func (f nullableBoolRequestField) toServiceInput() **bool {
	if !f.set {
		return nil
	}
	return &f.value
}

// nullableUint64RequestField は PATCH DTO で「未送信」と「JSON null」を区別する。
// Go 標準の **uint64 だけでは encoding/json が null と未送信の両方を nil にするため、
// set フラグでフィールド存在を保持し、service 層の **uint64（nil / &nil / &&value）へ変換する。
type nullableUint64RequestField struct {
	set   bool
	value *uint64
}

func (f *nullableUint64RequestField) UnmarshalJSON(data []byte) error {
	f.set = true
	if string(data) == "null" {
		f.value = nil
		return nil
	}
	var value uint64
	if err := json.Unmarshal(data, &value); err != nil {
		return apperrors.WrapInvalidInput("数値IDの形式が正しくありません")
	}
	f.value = &value
	return nil
}

func (f nullableUint64RequestField) toServiceInput() **uint64 {
	if !f.set {
		return nil
	}
	return &f.value
}

// nullableDateRequestField は PATCH DTO で「未送信」と「JSON null」を区別する。
// set フラグでフィールド存在を保持し、service 層の **time.Time
// （nil / &nil / &&value）へ変換する。
type nullableDateRequestField struct {
	set   bool
	value *time.Time
}

func (f *nullableDateRequestField) UnmarshalJSON(data []byte) error {
	f.set = true
	if string(data) == "null" {
		f.value = nil
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return apperrors.WrapInvalidInput(httpapi.FlexibleDateInvalidInputMsg)
	}
	value, err := httpapi.ParseFlexibleDate(raw)
	if err != nil {
		return apperrors.WrapInvalidInput(httpapi.FlexibleDateInvalidInputMsg)
	}
	f.value = &value
	return nil
}

func (f nullableDateRequestField) toServiceInput() **time.Time {
	if !f.set {
		return nil
	}
	return &f.value
}

func newListOwnersQuery(values url.Values) (listOwnersQuery, error) {
	search := values.Get("search")
	if len(search) > 255 {
		return listOwnersQuery{}, apperrors.WrapInvalidInput("search must be at most 255 characters")
	}
	return listOwnersQuery{Search: search}, nil
}

// createPetForOwnerRequest は飼主登録時のペット入力バインド struct
type createPetForOwnerRequest struct {
	Name            string    `json:"name"              binding:"required,max=255"`
	AnimalSpeciesID uint64    `json:"animal_species_id" binding:"required"`
	NameKana        string    `json:"name_kana"        binding:"omitempty,max=100"`
	Breed           string    `json:"breed"            binding:"omitempty,max=100"`
	Color           string    `json:"color"            binding:"omitempty,max=100"`
	BloodType       string    `json:"blood_type"       binding:"omitempty,max=32"`
	MicrochipNumber string    `json:"microchip_number" binding:"omitempty,max=64"`
	Gender          string    `json:"gender"            binding:"omitempty,oneof=male female unknown"`
	Status          string    `json:"status"            binding:"omitempty,oneof=alive deceased"`
	BirthDate       *jsonDate `json:"birth_date"`
	Weight          *float64  `json:"weight"`
	NeuteredDate    *jsonDate `json:"neutered_date"`
	AcquisitionType string    `json:"acquisition_type"`
	DangerLevel     string    `json:"danger_level"`
	DangerReason    *string   `json:"danger_reason"`
	Food            string    `json:"food"             binding:"omitempty,max=500"`
	Environment     string    `json:"environment"      binding:"omitempty,max=500"`
	InsuranceID     *uint64   `json:"insurance_id"`
	Remarks         string    `json:"remarks"          binding:"omitempty,max=2000"`
}

func (r *createPetForOwnerRequest) toServiceInput() CreatePetForOwnerInput {
	return CreatePetForOwnerInput{
		Name:            r.Name,
		AnimalSpeciesID: r.AnimalSpeciesID,
		PetNameKana:     r.NameKana,
		Breed:           r.Breed,
		Color:           r.Color,
		BloodType:       r.BloodType,
		MicrochipNumber: r.MicrochipNumber,
		Gender:          r.Gender,
		Status:          r.Status,
		BirthDate:       jsonDatePtr(r.BirthDate),
		Weight:          r.Weight,
		NeuteredDate:    jsonDatePtr(r.NeuteredDate),
		AcquisitionType: r.AcquisitionType,
		DangerLevel:     r.DangerLevel,
		DangerReason:    r.DangerReason,
		Food:            r.Food,
		Environment:     r.Environment,
		InsuranceID:     r.InsuranceID,
		Remarks:         r.Remarks,
	}
}

// createOwnerRequest は飼主作成のバインド struct
type createOwnerRequest struct {
	// ClinicID は登録先医院の任意指定 (#84 Q12=A: 登録フォームでのみ医院指定可)。
	// 未指定時は JWT/X-Clinic-ID 由来の clinic_id を使用する。
	// 指定時はハンドラで所属医院 (clinic_ids) との一致を必ず検証すること。
	ClinicID       *uint64                    `json:"clinic_id"`
	OwnerName      string                     `json:"owner_name"       binding:"required,max=255"`
	OwnerNameKana  string                     `json:"owner_name_kana"  binding:"omitempty,max=100"`
	BirthDate      *jsonDate                  `json:"birth_date"`
	Company        string                     `json:"company"          binding:"omitempty,max=200"`
	PostalCode     string                     `json:"postal_code"`
	Address1       string                     `json:"address1"         binding:"omitempty,max=200"`
	Address2       string                     `json:"address2"         binding:"omitempty,max=200"`
	HomePostalCode string                     `json:"home_postal_code" binding:"omitempty,max=10"`
	HomeAddress1   string                     `json:"home_address1"    binding:"omitempty,max=200"`
	HomeAddress2   string                     `json:"home_address2"    binding:"omitempty,max=200"`
	Phone          string                     `json:"phone"            binding:"omitempty,max=30"`
	CompanyPhone   string                     `json:"company_phone"    binding:"omitempty,max=30"`
	Email          string                     `json:"email"            binding:"omitempty,max=254"`
	Remarks        string                     `json:"remarks"          binding:"omitempty,max=2000"`
	IsDangerous    bool                       `json:"is_dangerous"`
	DiscountRate   float64                    `json:"discount_rate"`
	MembershipType string                     `json:"membership_type"  binding:"omitempty,oneof=non_member member deceased transferred"`
	DMPreference   *bool                      `json:"dm_preference"`
	Pets           []createPetForOwnerRequest `json:"pets"`
}

func (r *createOwnerRequest) toServiceInput() CreateOwnerInput {
	pets := make([]CreatePetForOwnerInput, 0, len(r.Pets))
	for i := range r.Pets {
		pets = append(pets, r.Pets[i].toServiceInput())
	}

	return CreateOwnerInput{
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
		DMPreference:   r.DMPreference,
		Pets:           pets,
	}
}

// updateOwnerRequest は飼主更新のバインド struct（全フィールドポインタ型）
type updateOwnerRequest struct {
	OwnerName      *string                  `json:"owner_name"       binding:"omitempty,max=255"`
	OwnerNameKana  *string                  `json:"owner_name_kana"  binding:"omitempty,max=100"`
	BirthDate      nullableDateRequestField `json:"birth_date"`
	Company        *string                  `json:"company"          binding:"omitempty,max=200"`
	PostalCode     *string                  `json:"postal_code"`
	Address1       *string                  `json:"address1"         binding:"omitempty,max=200"`
	Address2       *string                  `json:"address2"         binding:"omitempty,max=200"`
	HomePostalCode *string                  `json:"home_postal_code" binding:"omitempty,max=10"`
	HomeAddress1   *string                  `json:"home_address1"    binding:"omitempty,max=200"`
	HomeAddress2   *string                  `json:"home_address2"    binding:"omitempty,max=200"`
	Phone          *string                  `json:"phone"            binding:"omitempty,max=30"`
	CompanyPhone   *string                  `json:"company_phone"    binding:"omitempty,max=30"`
	Email          *string                  `json:"email"            binding:"omitempty,max=254"`
	Remarks        *string                  `json:"remarks"          binding:"omitempty,max=2000"`
	IsDangerous    *bool                    `json:"is_dangerous"`
	DiscountRate   *float64                 `json:"discount_rate"`
	MembershipType *string                  `json:"membership_type"  binding:"omitempty,oneof=non_member member deceased transferred"`
	DMPreference   nullableBoolRequestField `json:"dm_preference"`
}

func (r *updateOwnerRequest) toServiceInput() *UpdateOwnerInput {
	var membershipType *model.MembershipType
	if r.MembershipType != nil {
		mt := model.MembershipType(*r.MembershipType)
		membershipType = &mt
	}

	return &UpdateOwnerInput{
		OwnerName:      r.OwnerName,
		OwnerNameKana:  r.OwnerNameKana,
		BirthDate:      r.BirthDate.toServiceInput(),
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
		DMPreference:   r.DMPreference.toServiceInput(),
	}
}

// patchOwnerLineUserIDRequest は LINE User ID 連携リクエスト（BE-005）。nil で連携解除。
type patchOwnerLineUserIDRequest struct {
	// INF-03: bound length at HTTP boundary; service validates character class.
	LineUserID *string `json:"line_user_id" binding:"omitempty,max=64"`
}

// patchOwnerDeliveryExclusionRequest は配信除外フラグ更新リクエスト（FEAT-381）。
type patchOwnerDeliveryExclusionRequest struct {
	Excluded bool    `json:"excluded"`
	Reason   *string `json:"reason"   binding:"omitempty,max=100"`
}

func (r patchOwnerDeliveryExclusionRequest) toServiceInput() UpdateDeliveryExclusionInput {
	return UpdateDeliveryExclusionInput(r)
}

// patchOwnerDeliveryCautionRequest は配信注意フラグ更新リクエスト（FEAT-381-2）。
type patchOwnerDeliveryCautionRequest struct {
	Caution bool   `json:"caution"`
	Reason  string `json:"reason" binding:"omitempty,max=100"`
}

func (r patchOwnerDeliveryCautionRequest) toServiceInput() UpdateDeliveryCautionInput {
	return UpdateDeliveryCautionInput(r)
}

// patchOwnerTransferStatusRequest は転院フラグ更新リクエスト（FEAT-381）。
type patchOwnerTransferStatusRequest struct {
	IsTransferred bool `json:"is_transferred"`
}

func (r patchOwnerTransferStatusRequest) toServiceInput() UpdateTransferStatusInput {
	return UpdateTransferStatusInput(r)
}
