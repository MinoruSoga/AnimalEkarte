package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/trimming"
)

// BE9-2E compatibility aliases; remove with the service aggregator in BE9-2F.
type CreateTrimmingOptionInput = trimming.CreateTrimmingOptionInput
type UpdateTrimmingOptionInput = trimming.UpdateTrimmingOptionInput
type TrimmingOptionService = trimming.TrimmingOptionService

func NewTrimmingOptionService(repo repository.TrimmingOptionRepository) TrimmingOptionService {
	return trimming.NewTrimmingOptionService(repo)
}
