package handler

// LoginResponse はログインレスポンスの構造
type LoginResponse struct {
	IsSystemAdmin bool        `json:"is_system_admin"`
	User          *MeResponse `json:"user"`
}

// MeResponse はユーザー情報レスポンスの構造
type MeResponse struct {
	ID            string               `json:"id"`
	Email         string               `json:"email"`
	DisplayName   string               `json:"display_name"`
	IsSystemAdmin bool                 `json:"is_system_admin"`
	Occupation    *string              `json:"occupation,omitempty"`
	MainClinicID  string               `json:"main_clinic_id"`
	Clinic        *MeClinicInfo        `json:"clinic,omitempty"`
	Clinics       []MeClinicMembership `json:"clinics,omitempty"`
	Permissions   EffectivePermissions `json:"permissions"`
}

// MeClinicInfo はユーザー所属クリニックの詳細情報
type MeClinicInfo struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	PostalCode         string  `json:"postal_code"`
	Address            string  `json:"address"`
	PhoneNumber        string  `json:"phone_number"`
	FaxNumber          string  `json:"fax_number"`
	RegistrationNumber string  `json:"registration_number"`
	DirectorName       string  `json:"director_name"`
	Email              string  `json:"email"`
	Website            string  `json:"website"`
	LogoURL            *string `json:"logo_url,omitempty"`
	// BUG-367: インボイス対応帳票用（税率別内訳計算に使用）
	StandardTaxRate                           float64 `json:"standard_tax_rate"`
	ReducedTaxRate                            float64 `json:"reduced_tax_rate"`
	AccountingDocumentShowLogo                bool    `json:"accounting_document_show_logo"`
	AccountingDocumentShowRegistrationWarning bool    `json:"accounting_document_show_registration_warning"`
	AccountingDocumentShowItemCategory        bool    `json:"accounting_document_show_item_category"`
	AccountingDocumentFooterNote              string  `json:"accounting_document_footer_note"`
	// #190: セクション表示/非表示トグルと表示順
	AccountingDocumentShowClinicHeader   bool     `json:"accounting_document_show_clinic_header"`
	AccountingDocumentShowOwnerPetInfo   bool     `json:"accounting_document_show_owner_pet_info"`
	AccountingDocumentShowItemsTable     bool     `json:"accounting_document_show_items_table"`
	AccountingDocumentShowPaymentSummary bool     `json:"accounting_document_show_payment_summary"`
	AccountingDocumentSectionOrder       []string `json:"accounting_document_section_order"`
}

// MeClinicMembership はユーザーのクリニック所属情報
type MeClinicMembership struct {
	ClinicID   string `json:"clinic_id"`
	ClinicName string `json:"clinic_name"`
	IsMain     bool   `json:"is_main"`
}

// EffectivePermissions はユーザーの実効権限マップ
type EffectivePermissions map[string]ResourcePermission

// ResourcePermission はリソース単位の権限
type ResourcePermission struct {
	View   bool `json:"view"`
	Create bool `json:"create"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}
