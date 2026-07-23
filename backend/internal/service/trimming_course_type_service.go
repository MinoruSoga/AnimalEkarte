package service

import (
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/trimming"
)

// BE9-2E compatibility aliases; remove with the service aggregator in BE9-2F.
type CreateTrimmingCourseTypeInput = trimming.CreateTrimmingCourseTypeInput
type UpdateTrimmingCourseTypeInput = trimming.UpdateTrimmingCourseTypeInput
type TrimmingCourseTypeService = trimming.TrimmingCourseTypeService

func NewTrimmingCourseTypeService(
	repo repository.TrimmingCourseTypeRepository,
	tx repository.Transactor,
) TrimmingCourseTypeService {
	return trimming.NewTrimmingCourseTypeService(repo, tx)
}
