package model

import (
	"time"

	"github.com/google/uuid"
)

type UserType string

const (
	UserTypeSystemAdmin UserType = "system_admin"
	UserTypeClinicAdmin UserType = "clinic_admin"
	UserTypeStaff       UserType = "staff"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusLocked   AccountStatus = "locked"
)

type JobTitle string

const (
	JobTitleVeterinarian JobTitle = "veterinarian"
	JobTitleNurse        JobTitle = "nurse"
	JobTitleTrimmer      JobTitle = "trimmer"
	JobTitleReception    JobTitle = "reception"
	JobTitleGeneralStaff JobTitle = "general_staff"
)

type PermissionType string

const (
	PermissionAccountAdmin    PermissionType = "account_admin"
	PermissionMedical         PermissionType = "medical"
	PermissionMedicalRead     PermissionType = "medical_read"
	PermissionTrimming        PermissionType = "trimming"
	PermissionBilling         PermissionType = "billing"
	PermissionReception       PermissionType = "reception"
	PermissionHospitalization PermissionType = "hospitalization"
	PermissionMasterAdmin     PermissionType = "master_admin"
	PermissionShiftAdmin      PermissionType = "shift_admin"
	PermissionInventory       PermissionType = "inventory"
)

type ClinicInfo struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name               string    `gorm:"not null"                                        json:"name"`
	BranchName         string    `gorm:"default:''"                                      json:"branch_name"`
	PostalCode         string    `gorm:"default:''"                                      json:"postal_code"`
	Address            string    `gorm:"default:''"                                      json:"address"`
	PhoneNumber        string    `gorm:"default:''"                                      json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                                      json:"fax_number"`
	RegistrationNumber string    `gorm:"default:''"                                      json:"registration_number"`
	DirectorName       string    `gorm:"default:''"                                      json:"director_name"`
	Email              string    `gorm:"default:''"                                      json:"email"`
	Website            string    `gorm:"default:''"                                      json:"website"`
	LogoURL            string    `gorm:"default:''"                                      json:"logo_url"`
	CreatedAt          time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (ClinicInfo) TableName() string { return "clinic_info" }

type Clinic struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name               string    `gorm:"not null"                                        json:"name"`
	BranchName         string    `gorm:"default:''"                                      json:"branch_name"`
	PostalCode         string    `gorm:"default:''"                                      json:"postal_code"`
	Address            string    `gorm:"default:''"                                      json:"address"`
	PhoneNumber        string    `gorm:"default:''"                                      json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                                      json:"fax_number"`
	RegistrationNumber string    `gorm:"default:''"                                      json:"registration_number"`
	DirectorName       string    `gorm:"default:''"                                      json:"director_name"`
	Email              string    `gorm:"default:''"                                      json:"email"`
	Website            string    `gorm:"default:''"                                      json:"website"`
	LogoURL            string    `gorm:"default:''"                                      json:"logo_url"`
	IsActive           bool      `gorm:"default:true"                                    json:"is_active"`
	CreatedAt          time.Time `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (Clinic) TableName() string { return "clinics" }

type UserAccount struct {
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email           string        `gorm:"not null;uniqueIndex"                           json:"email"`
	DisplayName     string        `gorm:"not null"                                       json:"display_name"`
	DisplayNameKana string        `gorm:"default:''"                                     json:"display_name_kana"`
	UserType        UserType      `gorm:"type:user_type;not null;default:'staff'"        json:"user_type"`
	JobTitle        *JobTitle     `gorm:"type:job_title"                                 json:"job_title,omitempty"`
	Status          AccountStatus `gorm:"type:account_status;default:'active'"           json:"status"`
	AvatarURL       string        `gorm:"default:''"                                     json:"avatar_url"`
	StaffID         *uuid.UUID    `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (UserAccount) TableName() string { return "user_accounts" }

type UserClinicMembership struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;not null"                             json:"user_id"`
	ClinicID uuid.UUID `gorm:"type:uuid;not null"                             json:"clinic_id"`
	IsMain   bool      `gorm:"default:false"                                  json:"is_main"`
	JoinedAt time.Time `gorm:"default:now()"                                  json:"joined_at"`
}

func (UserClinicMembership) TableName() string { return "user_clinic_memberships" }

type UserPermission struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"                             json:"user_id"`
	ClinicID   uuid.UUID      `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Permission PermissionType `gorm:"type:permission_type;not null"                  json:"permission"`
	GrantedBy  *uuid.UUID     `gorm:"type:uuid"                                      json:"granted_by,omitempty"`
	GrantedAt  time.Time      `gorm:"default:now()"                                  json:"granted_at"`
}

func (UserPermission) TableName() string { return "user_permissions" }
