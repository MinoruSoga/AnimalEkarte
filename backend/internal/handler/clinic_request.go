package handler

// createClinicRequest はクリニック作成リクエスト。
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

// updateClinicRequest はクリニック更新リクエスト。
// PATCH セマンティクス: 未送信フィールドは nil → 既存値を保持する。
type updateClinicRequest struct {
	Name               *string  `json:"name"`
	PostalCode         *string  `json:"postal_code"`
	Address            *string  `json:"address"`
	PhoneNumber        *string  `json:"phone_number"`
	FaxNumber          *string  `json:"fax_number"`
	RegistrationNumber *string  `json:"registration_number"`
	DirectorName       *string  `json:"director_name"`
	Email              *string  `json:"email"`
	Website            *string  `json:"website"`
	LogoURL            *string  `json:"logo_url"`
	IsActive           *bool    `json:"is_active"`
	StandardTaxRate    *float64 `json:"standard_tax_rate"`
	ReducedTaxRate     *float64 `json:"reduced_tax_rate"`
}
