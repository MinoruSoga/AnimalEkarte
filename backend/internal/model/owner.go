package model

import (
	"time"

	"github.com/google/uuid"
)

// Owner 飼い主モデル
type Owner struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	NameKana  string    `json:"name_kana" gorm:"type:varchar(100)"`
	Phone     string    `json:"phone" gorm:"type:varchar(20)"`
	Email     string    `json:"email" gorm:"type:varchar(255)"`
	Address   string    `json:"address" gorm:"type:text"`
	Notes     string    `json:"notes" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Pets []Pet `json:"pets,omitempty" gorm:"foreignKey:OwnerID"`
}

// TableName テーブル名を指定
func (Owner) TableName() string {
	return "owners"
}

// CreateOwnerRequest 飼い主作成リクエスト
type CreateOwnerRequest struct {
	Name     string `json:"name" binding:"required"`
	NameKana string `json:"name_kana"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Notes    string `json:"notes"`
}

// UpdateOwnerRequest 飼い主更新リクエスト
type UpdateOwnerRequest struct {
	Name     string `json:"name"`
	NameKana string `json:"name_kana"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Notes    string `json:"notes"`
}
