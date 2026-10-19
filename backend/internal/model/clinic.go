package model

import (
	"github.com/lib/pq"
	"time"
)

type Clinic struct {
	ID                                        uint64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	CompanyID                                 uint64         `gorm:"column:company_id;type:bigint;not null" json:"company_id"`
	Name                                      string         `gorm:"column:name;type:text;not null" json:"name"`
	PostalCode                                string         `gorm:"column:postal_code;type:text;not null;default:''" json:"postal_code"`
	Address                                   string         `gorm:"column:address;type:text;not null;default:''" json:"address"`
	PhoneNumber                               string         `gorm:"column:phone_number;type:text;not null;default:''" json:"phone_number"`
	FaxNumber                                 string         `gorm:"column:fax_number;type:text;not null;default:''" json:"fax_number"`
	RegistrationNumber                        string         `gorm:"column:registration_number;type:text;not null;default:''" json:"registration_number"`
	DirectorName                              string         `gorm:"column:director_name;type:text;not null;default:''" json:"director_name"`
	Email                                     string         `gorm:"column:email;type:text;not null;default:''" json:"email"`
	Website                                   string         `gorm:"column:website;type:text;not null;default:''" json:"website"`
	LogoURL                                   string         `gorm:"column:logo_url;type:text;not null;default:''" json:"logo_url"`
	IsActive                                  bool           `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
	StandardTaxRate                           ExactNumeric   `gorm:"column:standard_tax_rate;type:numeric;not null;default:0.10" json:"standard_tax_rate"`
	ReducedTaxRate                            ExactNumeric   `gorm:"column:reduced_tax_rate;type:numeric;not null;default:0.08" json:"reduced_tax_rate"`
	AccountingDocumentShowLogo                bool           `gorm:"column:accounting_document_show_logo;type:boolean;not null;default:false" json:"accounting_document_show_logo"`
	AccountingDocumentShowRegistrationWarning bool           `gorm:"column:accounting_document_show_registration_warning;type:boolean;not null;default:true" json:"accounting_document_show_registration_warning"`
	AccountingDocumentShowItemCategory        bool           `gorm:"column:accounting_document_show_item_category;type:boolean;not null;default:true" json:"accounting_document_show_item_category"`
	AccountingDocumentFooterNote              string         `gorm:"column:accounting_document_footer_note;type:text;not null;default:''" json:"accounting_document_footer_note"`
	AccountingDocumentShowClinicHeader        bool           `gorm:"column:accounting_document_show_clinic_header;type:boolean;not null;default:true" json:"accounting_document_show_clinic_header"`
	AccountingDocumentShowOwnerPetInfo        bool           `gorm:"column:accounting_document_show_owner_pet_info;type:boolean;not null;default:true" json:"accounting_document_show_owner_pet_info"`
	AccountingDocumentShowItemsTable          bool           `gorm:"column:accounting_document_show_items_table;type:boolean;not null;default:true" json:"accounting_document_show_items_table"`
	AccountingDocumentShowPaymentSummary      bool           `gorm:"column:accounting_document_show_payment_summary;type:boolean;not null;default:true" json:"accounting_document_show_payment_summary"`
	AccountingDocumentSectionOrder            pq.StringArray `gorm:"column:accounting_document_section_order;type:text[];not null;default:'{}'" json:"accounting_document_section_order"`
	CreatedAt                                 time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt                                 time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (Clinic) TableName() string { return "clinics" }
