package model

import (
	"time"

	"github.com/google/uuid"
)

type Pet struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name      string     `json:"name" gorm:"size:100;not null"`
	Species   string     `json:"species" gorm:"size:50;not null"`
	Breed     string     `json:"breed" gorm:"size:100"`
	BirthDate *time.Time `json:"birth_date"`
	Gender    string     `json:"gender" gorm:"size:10"`
	Weight    *float64   `json:"weight" gorm:"type:decimal(5,2)"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreatePetRequest struct {
	Name      string  `json:"name" binding:"required"`
	Species   string  `json:"species" binding:"required"`
	Breed     string  `json:"breed"`
	BirthDate string  `json:"birth_date"`
	Gender    string  `json:"gender"`
	Weight    float64 `json:"weight"`
}

type UpdatePetRequest struct {
	Name      string  `json:"name"`
	Species   string  `json:"species"`
	Breed     string  `json:"breed"`
	BirthDate string  `json:"birth_date"`
	Gender    string  `json:"gender"`
	Weight    float64 `json:"weight"`
}
