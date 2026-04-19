package handler

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
