package medicalrecord

import (
	"context"
	"math"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type CreateExamTypeFieldInput struct {
	Name            string
	InspectionValue string
	NormalValue     string
	Unit            string
	SortOrder       int
}

type CreateExamTypeFieldCommand struct {
	ExamTypeID uint64
	Field      CreateExamTypeFieldInput
}

type UpdateExamTypeFieldInput struct {
	Name            *string
	InspectionValue *string
	NormalValue     *string
	Unit            *string
	SortOrder       *int
}

type ReferenceRangeInput struct {
	AnimalSpeciesID uint64
	RefMin          *float64
	RefMax          *float64
	QualitativeMin  *string
	QualitativeMax  *string
}

type ReplaceReferenceRangesCommand struct {
	ExamTypeFieldID uint64
	Ranges          []ReferenceRangeInput
}

type ExamTypeFieldResult struct {
	Field           model.ExamTypeField
	ReferenceRanges []model.ExamReferenceRange
}

func validateReferenceRangeInputs(inputs []ReferenceRangeInput) error {
	seen := make(map[uint64]struct{}, len(inputs))
	for _, input := range inputs {
		if input.AnimalSpeciesID == 0 {
			return apperrors.WrapInvalidInput("animal_species_id must be positive")
		}
		if _, ok := seen[input.AnimalSpeciesID]; ok {
			return apperrors.WrapInvalidInput("animal_species_id must be unique")
		}
		seen[input.AnimalSpeciesID] = struct{}{}

		hasNumeric := input.RefMin != nil || input.RefMax != nil
		hasQualitative := input.QualitativeMin != nil || input.QualitativeMax != nil
		if hasNumeric && hasQualitative {
			return apperrors.WrapInvalidInput("numeric and qualitative bounds cannot coexist")
		}
		if input.RefMin != nil && (math.IsNaN(*input.RefMin) || math.IsInf(*input.RefMin, 0)) {
			return apperrors.WrapInvalidInput("ref_min must be finite")
		}
		if input.RefMax != nil && (math.IsNaN(*input.RefMax) || math.IsInf(*input.RefMax, 0)) {
			return apperrors.WrapInvalidInput("ref_max must be finite")
		}
		if input.RefMin != nil && input.RefMax != nil && *input.RefMin > *input.RefMax {
			return apperrors.WrapInvalidInput("ref_min must be less than or equal to ref_max")
		}

		var minIndex, maxIndex int
		if input.QualitativeMin != nil {
			var ok bool
			minIndex, ok = qualitativeValueIndex(*input.QualitativeMin)
			if !ok {
				return apperrors.WrapInvalidInput("qualitative_min is unsupported")
			}
		}
		if input.QualitativeMax != nil {
			var ok bool
			maxIndex, ok = qualitativeValueIndex(*input.QualitativeMax)
			if !ok {
				return apperrors.WrapInvalidInput("qualitative_max is unsupported")
			}
		}
		if input.QualitativeMin != nil && input.QualitativeMax != nil && minIndex > maxIndex {
			return apperrors.WrapInvalidInput("qualitative_min must be less than or equal to qualitative_max")
		}
	}
	return nil
}

func (s *examTypeService) CreateField(
	ctx context.Context,
	clinicID uint64,
	command *CreateExamTypeFieldCommand,
) (*model.ExamTypeField, error) {
	if command == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateRequiredName(command.Field.Name); err != nil {
		return nil, err
	}
	var field *model.ExamTypeField
	err := s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.FindByID(txCtx, clinicID, command.ExamTypeID); err != nil {
			return err
		}
		created := &model.ExamTypeField{
			ClinicID:        clinicID,
			ExamTypeID:      command.ExamTypeID,
			Name:            command.Field.Name,
			InspectionValue: command.Field.InspectionValue,
			NormalValue:     command.Field.NormalValue,
			Unit:            command.Field.Unit,
			SortOrder:       command.Field.SortOrder,
		}
		if err := s.repo.CreateField(txCtx, created); err != nil {
			return err
		}
		field = created
		return nil
	})
	return field, err
}

func (s *examTypeService) UpdateField(
	ctx context.Context,
	clinicID, examTypeID, fieldID uint64,
	input *UpdateExamTypeFieldInput,
) (*ExamTypeFieldResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateOptionalName(input.Name); err != nil {
		return nil, err
	}
	fields := make(map[string]any)
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.InspectionValue != nil {
		fields["inspection_value"] = *input.InspectionValue
	}
	if input.NormalValue != nil {
		fields["normal_value"] = *input.NormalValue
	}
	if input.Unit != nil {
		fields["unit"] = *input.Unit
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	var result *ExamTypeFieldResult
	err := s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockFieldByID(txCtx, clinicID, examTypeID, fieldID); err != nil {
			return err
		}
		updated, err := s.repo.UpdateField(txCtx, clinicID, examTypeID, fieldID, fields)
		if err != nil {
			return err
		}
		ranges, err := s.repo.FindReferenceRangesByFieldIDs(txCtx, clinicID, []uint64{fieldID})
		if err != nil {
			return err
		}
		result = &ExamTypeFieldResult{Field: *updated, ReferenceRanges: ranges[fieldID]}
		return nil
	})
	return result, err
}

func (s *examTypeService) DeleteField(ctx context.Context, clinicID, examTypeID, fieldID uint64) error {
	return s.withTx(ctx, func(txCtx context.Context) error {
		if _, err := s.repo.LockFieldByID(txCtx, clinicID, examTypeID, fieldID); err != nil {
			return err
		}
		resultCount, err := s.repo.CountExamResultsByFieldID(txCtx, fieldID)
		if err != nil {
			return err
		}
		rangeCount, err := s.repo.CountReferenceRangesByFieldID(txCtx, fieldID)
		if err != nil {
			return err
		}
		if resultCount > 0 || rangeCount > 0 {
			return apperrors.WrapConflict("この検査項目は検査結果または基準値で使用中のため削除できません")
		}
		return s.repo.DeleteField(txCtx, clinicID, examTypeID, fieldID)
	})
}

func (s *examTypeService) ReorderFields(
	ctx context.Context,
	clinicID, examTypeID uint64,
	ids []uint64,
) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput(errMsgIDsNotEmpty)
	}
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return apperrors.WrapInvalidInput("ids must be positive")
		}
		if _, ok := seen[id]; ok {
			return apperrors.WrapInvalidInput("ids must be unique")
		}
		seen[id] = struct{}{}
	}
	return s.withTx(ctx, func(txCtx context.Context) error {
		return s.repo.ReorderFields(txCtx, clinicID, examTypeID, ids)
	})
}

func (s *examTypeService) ReplaceReferenceRanges(
	ctx context.Context,
	clinicID, examTypeID uint64,
	command *ReplaceReferenceRangesCommand,
) (*ExamTypeFieldResult, error) {
	if command == nil {
		return nil, apperrors.WrapInvalidInput(errMsgInputNotNil)
	}
	if err := validateReferenceRangeInputs(command.Ranges); err != nil {
		return nil, err
	}
	ranges := make([]model.ExamReferenceRange, 0, len(command.Ranges))
	for _, input := range command.Ranges {
		ranges = append(ranges, model.ExamReferenceRange{
			ClinicID:        clinicID,
			ExamTypeFieldID: command.ExamTypeFieldID,
			AnimalSpeciesID: input.AnimalSpeciesID,
			RefMin:          input.RefMin,
			RefMax:          input.RefMax,
			QualitativeMin:  input.QualitativeMin,
			QualitativeMax:  input.QualitativeMax,
		})
	}
	var result *ExamTypeFieldResult
	err := s.withTx(ctx, func(txCtx context.Context) error {
		field, err := s.repo.LockFieldByID(txCtx, clinicID, examTypeID, command.ExamTypeFieldID)
		if err != nil {
			return err
		}
		for _, input := range command.Ranges {
			exists, err := s.repo.AnimalSpeciesExists(txCtx, input.AnimalSpeciesID)
			if err != nil {
				return err
			}
			if !exists {
				return apperrors.WrapInvalidInput("animal_species_id does not exist")
			}
		}
		if err := s.repo.ReplaceReferenceRanges(txCtx, clinicID, command.ExamTypeFieldID, ranges); err != nil {
			return err
		}
		saved, err := s.repo.FindReferenceRangesByFieldIDs(
			txCtx,
			clinicID,
			[]uint64{command.ExamTypeFieldID},
		)
		if err != nil {
			return err
		}
		result = &ExamTypeFieldResult{
			Field:           *field,
			ReferenceRanges: saved[command.ExamTypeFieldID],
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *examTypeService) ListReferenceRanges(
	ctx context.Context,
	clinicID uint64,
	fieldIDs []uint64,
) (map[uint64][]model.ExamReferenceRange, error) {
	return s.repo.FindReferenceRangesByFieldIDs(ctx, clinicID, fieldIDs)
}

func (s *examTypeService) withTx(ctx context.Context, fn func(context.Context) error) error {
	// MRB-07: nil transactor is composition/wiring failure (server internal), not client input → 500.
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("exam type transaction dependency is required")
	}
	return s.transactor.WithTx(ctx, fn)
}
