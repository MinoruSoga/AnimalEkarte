package handler

// LoginResponse はログインレスポンスの構造
type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	UserType  string      `json:"user_type"`
	User      *MeResponse `json:"user"`
}

// MeResponse はユーザー情報レスポンスの構造
type MeResponse struct {
	ID           string               `json:"id"`
	Email        string               `json:"email"`
	DisplayName  string               `json:"display_name"`
	UserType     string               `json:"user_type"`
	StaffRole    *string              `json:"staff_role,omitempty"`
	JobTitle     *string              `json:"job_title,omitempty"`
	MainClinicID string               `json:"main_clinic_id"`
	Clinic       *MeClinicInfo        `json:"clinic,omitempty"`
	Clinics      []MeClinicMembership `json:"clinics,omitempty"`
	Permissions  EffectivePermissions `json:"permissions"`
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
