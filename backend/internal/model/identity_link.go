package model

import (
	"time"

	"gorm.io/gorm"
)

// OwnerIdentityGroup anchors a manual multi-clinic owner identity link (#239).
// created_clinic_id is the RLS/write anchor and is immutable after insert.
type OwnerIdentityGroup struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedClinicID uint64         `gorm:"column:created_clinic_id;not null" json:"created_clinic_id"`
	Version         int64          `gorm:"not null;default:1" json:"version"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-"`

	Members []OwnerIdentityGroupMember `gorm:"foreignKey:GroupID;references:ID" json:"members,omitempty"`
}

func (OwnerIdentityGroup) TableName() string { return "owner_identity_groups" }

// OwnerIdentityGroupMember is a clinic-scoped owner row participating in a group.
type OwnerIdentityGroupMember struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupCreatedClinicID uint64         `gorm:"column:group_created_clinic_id;not null" json:"group_created_clinic_id"`
	GroupID              uint64         `gorm:"column:group_id;not null" json:"group_id"`
	ClinicID             uint64         `gorm:"column:clinic_id;not null" json:"clinic_id"`
	OwnerID              uint64         `gorm:"column:owner_id;not null" json:"owner_id"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-"`
}

func (OwnerIdentityGroupMember) TableName() string { return "owner_identity_group_members" }

// PetIdentityGroup anchors a pet identity link that must hang under an owner identity group.
type PetIdentityGroup struct {
	ID                        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedClinicID           uint64         `gorm:"column:created_clinic_id;not null" json:"created_clinic_id"`
	OwnerGroupCreatedClinicID uint64         `gorm:"column:owner_group_created_clinic_id;not null" json:"owner_group_created_clinic_id"`
	OwnerGroupID              uint64         `gorm:"column:owner_group_id;not null" json:"owner_group_id"`
	Version                   int64          `gorm:"not null;default:1" json:"version"`
	CreatedAt                 time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt                 gorm.DeletedAt `json:"-"`

	Members []PetIdentityGroupMember `gorm:"foreignKey:GroupID;references:ID" json:"members,omitempty"`
}

func (PetIdentityGroup) TableName() string { return "pet_identity_groups" }

// PetIdentityGroupMember is a clinic-scoped pet row participating in a pet identity group.
type PetIdentityGroupMember struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	GroupCreatedClinicID uint64         `gorm:"column:group_created_clinic_id;not null" json:"group_created_clinic_id"`
	GroupID              uint64         `gorm:"column:group_id;not null" json:"group_id"`
	ClinicID             uint64         `gorm:"column:clinic_id;not null" json:"clinic_id"`
	PetID                uint64         `gorm:"column:pet_id;not null" json:"pet_id"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-"`
}

func (PetIdentityGroupMember) TableName() string { return "pet_identity_group_members" }
