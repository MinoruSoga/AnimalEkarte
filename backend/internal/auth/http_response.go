package auth

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// LoginInput is the login request body.
type LoginInput struct {
	Email    string `json:"email"    binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginResponse is returned after successful authentication.
type LoginResponse struct {
	IsSystemAdmin bool        `json:"is_system_admin"`
	User          *MeResponse `json:"user"`
}

// MeResponse is the authenticated staff profile.
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

// MeClinicInfo is the selected clinic embedded in /me.
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

	StandardTaxRate                           float64  `json:"standard_tax_rate"`
	ReducedTaxRate                            float64  `json:"reduced_tax_rate"`
	AccountingDocumentShowLogo                bool     `json:"accounting_document_show_logo"`
	AccountingDocumentShowRegistrationWarning bool     `json:"accounting_document_show_registration_warning"`
	AccountingDocumentShowItemCategory        bool     `json:"accounting_document_show_item_category"`
	AccountingDocumentFooterNote              string   `json:"accounting_document_footer_note"`
	AccountingDocumentShowClinicHeader        bool     `json:"accounting_document_show_clinic_header"`
	AccountingDocumentShowOwnerPetInfo        bool     `json:"accounting_document_show_owner_pet_info"`
	AccountingDocumentShowItemsTable          bool     `json:"accounting_document_show_items_table"`
	AccountingDocumentShowPaymentSummary      bool     `json:"accounting_document_show_payment_summary"`
	AccountingDocumentSectionOrder            []string `json:"accounting_document_section_order"`
}

// MeClinicMembership is one clinic switch target.
type MeClinicMembership struct {
	ClinicID   string `json:"clinic_id"`
	ClinicName string `json:"clinic_name"`
	IsMain     bool   `json:"is_main"`
}

// EffectivePermissions is the HTTP resource-keyed permission map.
type EffectivePermissions map[string]ResourcePermission

// ResourcePermission is one resource's effective CRUD permission.
type ResourcePermission struct {
	View   bool `json:"view"`
	Create bool `json:"create"`
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
}

// ToMeResponse maps domain models to the public authenticated profile shape.
func ToMeResponse(
	staff *model.Staff,
	account *model.Account,
	mainClinicID string,
	clinicNameMap map[string]string,
	allClinics []model.Clinic,
	effectivePerms EffectivePermissions,
) *MeResponse {
	isSystemAdmin := account != nil && account.IsSystemAdmin
	permissions := effectivePerms
	if permissions == nil {
		permissions = make(EffectivePermissions)
	}

	var occupation *string
	if staff != nil && staff.Occupation != nil {
		name := staff.Occupation.Name
		occupation = &name
	}

	var staffID uint64
	var displayName string
	if staff != nil {
		staffID = staff.ID
		displayName = staff.Name
	}

	var email string
	if account != nil {
		email = account.Email
	}

	return &MeResponse{
		ID:            strconv.FormatUint(staffID, 10),
		Email:         email,
		DisplayName:   displayName,
		IsSystemAdmin: isSystemAdmin,
		Occupation:    occupation,
		MainClinicID:  mainClinicID,
		Clinic:        selectedMeClinicInfo(allClinics, mainClinicID),
		Clinics:       meClinicMemberships(staff, isSystemAdmin, mainClinicID, clinicNameMap, allClinics),
		Permissions:   permissions,
	}
}

func meClinicMemberships(
	staff *model.Staff,
	isSystemAdmin bool,
	mainClinicID string,
	clinicNameMap map[string]string,
	allClinics []model.Clinic,
) []MeClinicMembership {
	meClinicList := make([]MeClinicMembership, 0)
	if isSystemAdmin {
		for i := range allClinics {
			clinic := &allClinics[i]
			if clinic.ID == 0 || !clinic.IsActive {
				continue
			}
			clinicID := strconv.FormatUint(clinic.ID, 10)
			meClinicList = append(meClinicList, MeClinicMembership{
				ClinicID:   clinicID,
				ClinicName: clinic.Name,
				IsMain:     clinicID == mainClinicID,
			})
		}
		return meClinicList
	}
	if staff == nil {
		return meClinicList
	}
	for i := range staff.ClinicAssignments {
		assignment := &staff.ClinicAssignments[i]
		clinicID := strconv.FormatUint(assignment.ClinicID, 10)
		meClinicList = append(meClinicList, MeClinicMembership{
			ClinicID:   clinicID,
			ClinicName: clinicNameMap[clinicID],
			IsMain:     clinicID == mainClinicID,
		})
	}
	return meClinicList
}

func selectedMeClinicInfo(allClinics []model.Clinic, mainClinicID string) *MeClinicInfo {
	for i := range allClinics {
		clinic := &allClinics[i]
		if strconv.FormatUint(clinic.ID, 10) != mainClinicID {
			continue
		}
		var logoURL *string
		if clinic.LogoURL != "" {
			logo := clinic.LogoURL
			logoURL = &logo
		}
		return &MeClinicInfo{
			ID:                         strconv.FormatUint(clinic.ID, 10),
			Name:                       clinic.Name,
			PostalCode:                 clinic.PostalCode,
			Address:                    clinic.Address,
			PhoneNumber:                clinic.PhoneNumber,
			FaxNumber:                  clinic.FaxNumber,
			RegistrationNumber:         clinic.RegistrationNumber,
			DirectorName:               clinic.DirectorName,
			Email:                      clinic.Email,
			Website:                    clinic.Website,
			LogoURL:                    logoURL,
			StandardTaxRate:            clinic.StandardTaxRate,
			ReducedTaxRate:             clinic.ReducedTaxRate,
			AccountingDocumentShowLogo: clinic.AccountingDocumentShowLogo,
			AccountingDocumentShowRegistrationWarning: clinic.AccountingDocumentShowRegistrationWarning,
			AccountingDocumentShowItemCategory:        clinic.AccountingDocumentShowItemCategory,
			AccountingDocumentFooterNote:              clinic.AccountingDocumentFooterNote,
			AccountingDocumentShowClinicHeader:        clinic.AccountingDocumentShowClinicHeader,
			AccountingDocumentShowOwnerPetInfo:        clinic.AccountingDocumentShowOwnerPetInfo,
			AccountingDocumentShowItemsTable:          clinic.AccountingDocumentShowItemsTable,
			AccountingDocumentShowPaymentSummary:      clinic.AccountingDocumentShowPaymentSummary,
			AccountingDocumentSectionOrder:            append([]string{}, clinic.AccountingDocumentSectionOrder...),
		}
	}
	return nil
}

// BuildAllPermissions returns the system-administrator permission map.
func BuildAllPermissions() EffectivePermissions {
	return ToHTTPEffectivePermissions(buildAuthAllPermissions())
}

// CalculateEffectivePermissions returns the fail-closed HTTP permission map.
func (h *HTTPHandler) CalculateEffectivePermissions(
	c *gin.Context,
	isSystemAdmin bool,
	staffID uint64,
) EffectivePermissions {
	if isSystemAdmin {
		return ToHTTPEffectivePermissions(
			h.authService().CalculateEffectivePermissions(c.Request.Context(), true, staffID, 0),
		)
	}

	// Peek only: caller (GetMe/Login) owns the success response; do not write an error body here (AUS-09).
	clinicID, ok := httpapi.PeekClinicID(c)
	if !ok {
		slog.ErrorContext(
			c.Request.Context(),
			"failed to extract clinic_id for effective permissions",
			"staff_id", staffID,
		)
		return make(EffectivePermissions)
	}
	return ToHTTPEffectivePermissions(
		h.authService().CalculateEffectivePermissions(
			c.Request.Context(),
			false,
			staffID,
			clinicID,
		),
	)
}

// ToHTTPEffectivePermissions maps the auth use-case permission representation.
func ToHTTPEffectivePermissions(perms AuthEffectivePermissions) EffectivePermissions {
	if perms == nil {
		return make(EffectivePermissions)
	}
	result := make(EffectivePermissions, len(perms))
	for resource, permission := range perms {
		result[resource] = ResourcePermission{
			View:   permission.View,
			Create: permission.Create,
			Edit:   permission.Edit,
			Delete: permission.Delete,
		}
	}
	return result
}
