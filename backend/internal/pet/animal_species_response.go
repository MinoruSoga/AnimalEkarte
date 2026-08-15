package pet

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type animalSpeciesResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toAnimalSpeciesResponse(s *model.AnimalSpecies) animalSpeciesResponse {
	return animalSpeciesResponse{
		ID:        s.ID,
		Name:      s.Name,
		IsActive:  s.IsActive,
		SortOrder: s.SortOrder,
		CreatedAt: httpapi.LocalTime(s.CreatedAt),
		UpdatedAt: httpapi.LocalTime(s.UpdatedAt),
	}
}
