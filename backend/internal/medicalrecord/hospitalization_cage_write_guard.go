package medicalrecord

import (
	"context"
	"errors"
	"slices"
	"strconv"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

const (
	codeCageAssignmentChanged = "cage_assignment_changed"
	codeCageOccupied          = "cage_occupied"
	codeOccupancyUnknown      = "occupancy_unknown"
)

func hospitalizationCageConflict(code, message string) error {
	return &apperrors.AppError{Code: code, Message: message, Err: apperrors.ErrConflict}
}

// lockCageAssignmentGroup precedes both the stay read and FindByID's FK SHARE
// lock. Save callers retain their ambient transaction and cage -> stay order.
func (s *hospitalizationService) lockCageAssignmentGroup(ctx context.Context, clinicID, cageID uint64) ([]model.Cage, error) {
	mergeRepo, ok := s.cageRepo.(CageMergeRepository)
	if !ok {
		return nil, apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	group, err := mergeRepo.LockCageGroupForUpdate(ctx, clinicID, cageID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock cage group")
	}
	for _, cage := range group {
		if cage.ID == cageID {
			if cage.MergedIntoCageID != nil {
				return nil, apperrors.WrapUnprocessable("連結中の構成ケージには割り当てできません")
			}
			return group, nil
		}
	}
	return nil, apperrors.WrapNotFound("cage", strconv.FormatUint(cageID, 10))
}

// validateCageAssignmentSubjects runs after all target cage groups are locked,
// before any subject is excluded from occupancy. Legacy status/dates are never
// evidence. This is the current normal-case guard, not the future 499 write CAS.
func (s *hospitalizationService) validateCageAssignmentSubjects(ctx context.Context, clinicID uint64, subjects []*model.Hospitalization) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("cage assignment requires ambient transaction")
	}
	reader, ok := s.cageRepo.(StayCageEffectiveStateReader)
	if !ok {
		return apperrors.WrapInternalServerError("stay effective-state reader dependency is required")
	}
	ids := make([]uint64, 0, len(subjects))
	for _, subject := range subjects {
		if subject == nil || subject.ClinicID != clinicID || subject.StayCaseID == nil || *subject.StayCaseID == 0 {
			return hospitalizationCageConflict(codeOccupancyUnknown, "割当対象の滞在を確認できません")
		}
		ids = append(ids, *subject.StayCaseID)
	}
	slices.Sort(ids)
	if len(slices.Compact(slices.Clone(ids))) != len(ids) {
		return hospitalizationCageConflict(codeOccupancyUnknown, "割当対象の滞在が重複しています")
	}
	states, err := reader.ReadEffectiveStatesForCageAssignment(ctx, clinicID, ids)
	if err != nil {
		return cageAssignmentReadError(err)
	}
	if len(states) != len(ids) {
		return hospitalizationCageConflict(codeOccupancyUnknown, "割当対象の滞在の根拠が不足しています")
	}
	byStay := make(map[uint64]StayCageEffectiveState, len(states))
	for _, state := range states {
		_, duplicate := byStay[state.StayCaseID]
		if state.ClinicID != clinicID || !slices.Contains(ids, state.StayCaseID) || duplicate ||
			state.Evidence.StayVersion <= 0 || state.Evidence.MovementHeadRevision < 0 {
			return hospitalizationCageConflict(codeOccupancyUnknown, "割当対象の滞在の根拠を確認できません")
		}
		// Head 0 is legitimate empty history. This subset cannot establish a
		// known lifecycle there until 499 connects the accepted reservation proof.
		if state.State != StayCageUnknown && state.Evidence.MovementHeadRevision == 0 {
			return hospitalizationCageConflict(codeOccupancyUnknown, "割当対象の滞在の証跡が未接続です")
		}
		byStay[state.StayCaseID] = state
	}
	for _, id := range ids {
		switch byStay[id].State {
		case StayCageReserved, StayCageAdmitted:
		case StayCageReleased:
			// A known released stay is ineligible, rather than unresolved.
			return apperrors.WrapInvalidInput("終了済みの滞在にはケージを割り当てできません")
		default:
			return hospitalizationCageConflict(codeOccupancyUnknown, "割当対象の滞在の有効状態を確認できません")
		}
	}
	return nil
}

func cageAssignmentReadError(err error) error {
	var unknown *cageOccupancyUnknown
	if errors.As(err, &unknown) {
		return hospitalizationCageConflict(codeOccupancyUnknown, "滞在の有効状態を確認できません")
	}
	return apperrors.Wrap(err, "failed to read effective cage occupancy")
}

// Always recheck others, independent of the removed hospitalization status.
func (s *hospitalizationService) validateCageAssignmentOccupancy(ctx context.Context, clinicID uint64, group []model.Cage, excluded []uint64) error {
	mergeRepo, ok := s.cageRepo.(CageMergeRepository)
	if !ok {
		return apperrors.WrapInternalServerError("cage merge repository dependency is required")
	}
	ids := make([]uint64, 0, len(group))
	for _, cage := range group {
		ids = append(ids, cage.ID)
	}
	occupied, err := mergeRepo.CountOccupiedCagesExcludingHospitalizations(ctx, clinicID, ids, excluded)
	if err != nil {
		return cageAssignmentReadError(err)
	}
	if len(occupied) > 0 {
		return hospitalizationCageConflict(codeCageOccupied, "ケージまたは連結グループが使用中のため割り当てできません")
	}
	return nil
}
