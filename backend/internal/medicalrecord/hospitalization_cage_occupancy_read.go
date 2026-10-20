package medicalrecord

import (
	"context"
	"errors"
	"math"
	"slices"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

var errCageOccupancyReadFailed = errors.New("cage occupancy read failed")

type hospitalizationCageAssignmentFinder interface {
	FindByID(context.Context, uint64, uint64) (*model.Hospitalization, error)
}

type hospitalizationCageOccupancySource interface {
	FindAll(context.Context, uint64, *string) ([]model.Cage, error)
	ListOccupiedCageIDsExcludingHospitalization(context.Context, uint64, *uint64) ([]uint64, error)
}

func readHospitalizationCageOccupancy(ctx context.Context, clinicID uint64, exclude *uint64, finder hospitalizationCageAssignmentFinder, source hospitalizationCageOccupancySource) (hospitalizationCageOccupancyResponse, error) {
	result := hospitalizationCageOccupancyResponse{ClinicID: strconv.FormatUint(clinicID, 10), State: "unknown"}
	if !validCageProjectionID(clinicID) || source == nil {
		return result, errCageOccupancyReadFailed
	}
	if exclude != nil {
		if !validCageProjectionID(*exclude) || finder == nil {
			return result, errCageOccupancyReadFailed
		}
		target, err := finder.FindByID(ctx, clinicID, *exclude)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return result, apperrors.WrapNotFound("hospitalization", strconv.FormatUint(*exclude, 10))
			}
			return result, errors.Join(errCageOccupancyReadFailed, err)
		}
		if target == nil || target.ID != *exclude || target.ClinicID != clinicID || target.DeletedAt.Valid {
			return result, apperrors.WrapNotFound("hospitalization", strconv.FormatUint(*exclude, 10))
		}
		targetID := strconv.FormatUint(*exclude, 10)
		result.ExcludeHospitalizationID = &targetID
		result.CurrentAssignment = &hospitalizationCageCurrentAssignment{State: "known"}
		if target.CageID != nil {
			if validCageProjectionID(*target.CageID) {
				cageID := strconv.FormatUint(*target.CageID, 10)
				result.CurrentAssignment.CageID = &cageID
			} else {
				result.CurrentAssignment.State = "unknown"
			}
		}
	}
	occupied, err := source.ListOccupiedCageIDsExcludingHospitalization(ctx, clinicID, exclude)
	if err != nil {
		var unknown *cageOccupancyUnknown
		if errors.As(err, &unknown) {
			return result, nil
		}
		return result, errors.Join(errCageOccupancyReadFailed, err)
	}
	cages, err := source.FindAll(ctx, clinicID, nil)
	if err != nil {
		return result, errors.Join(errCageOccupancyReadFailed, err)
	}
	candidates, known := projectCageAssignmentCandidates(clinicID, cages, occupied)
	if !known {
		return result, nil
	}
	result.State = "known"
	result.OccupiedCageIDs = cageProjectionStrings(occupied)
	result.CandidateCageIDs = cageProjectionStrings(candidates)
	return result, nil
}

func validCageProjectionID(id uint64) bool { return id > 0 && id <= math.MaxInt64 }

func cageProjectionStrings(ids []uint64) []string {
	ids = slices.Clone(ids)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, strconv.FormatUint(id, 10))
	}
	return result
}

// A group's representative is a candidate only when none of its members retain
// another stay's occupancy. Current assignment is deliberately not an input.
func projectCageAssignmentCandidates(clinicID uint64, cages []model.Cage, occupied []uint64) ([]uint64, bool) {
	byID := make(map[uint64]model.Cage, len(cages))
	for _, cage := range cages {
		if _, duplicate := byID[cage.ID]; duplicate || !validCageProjectionID(cage.ID) || cage.ClinicID != clinicID || cage.DeletedAt.Valid {
			return nil, false
		}
		byID[cage.ID] = cage
	}
	for _, cage := range cages {
		if cage.MergedIntoCageID != nil {
			parent, found := byID[*cage.MergedIntoCageID]
			if !found || parent.ID == cage.ID || parent.MergedIntoCageID != nil {
				return nil, false
			}
		}
	}
	blocked := make(map[uint64]bool, len(occupied))
	for _, id := range occupied {
		cage, found := byID[id]
		if !found {
			return nil, false
		}
		if cage.MergedIntoCageID != nil {
			blocked[*cage.MergedIntoCageID] = true
		} else {
			blocked[id] = true
		}
	}
	candidates := make([]uint64, 0, len(cages))
	for _, cage := range cages {
		if cage.IsActive && cage.MergedIntoCageID == nil && !blocked[cage.ID] {
			candidates = append(candidates, cage.ID)
		}
	}
	slices.Sort(candidates)
	return candidates, true
}
