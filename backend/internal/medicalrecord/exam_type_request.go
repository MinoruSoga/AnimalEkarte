package medicalrecord

type createExaminationTypeRequest struct {
	Name           string  `json:"name"             binding:"required"`
	Price          *int64  `json:"price"`
	IsActive       bool    `json:"is_active"`
	Description    string  `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	SortOrder      int     `json:"sort_order"`
	IsNonInsurance bool    `json:"is_non_insurance"`
}

func (r createExaminationTypeRequest) toServiceInput() *CreateExamTypeInput {
	return &CreateExamTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		ParentID:       r.ParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance,
	}
}

type updateExaminationTypeRequest struct {
	Name           *string `json:"name"`
	Price          *int64  `json:"price"`
	IsActive       *bool   `json:"is_active"`
	Description    *string `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	ClearParentID  bool    `json:"clear_parent_id"`
	SortOrder      *int    `json:"sort_order"`
	IsNonInsurance *bool   `json:"is_non_insurance"`
}

func (r updateExaminationTypeRequest) toServiceInput() *UpdateExamTypeInput {
	return &UpdateExamTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		ParentID:       r.ParentID,
		ClearParentID:  r.ClearParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance,
	}
}

type createExamTypeFieldRequest struct {
	Name            string `json:"name" binding:"required"`
	InspectionValue string `json:"inspection_value"`
	NormalValue     string `json:"normal_value"`
	Unit            string `json:"unit"`
	SortOrder       int    `json:"sort_order"`
}

func (r createExamTypeFieldRequest) toServiceCommand(examTypeID uint64) *CreateExamTypeFieldCommand {
	return &CreateExamTypeFieldCommand{
		ExamTypeID: examTypeID,
		Field: CreateExamTypeFieldInput{
			Name:            r.Name,
			InspectionValue: r.InspectionValue,
			NormalValue:     r.NormalValue,
			Unit:            r.Unit,
			SortOrder:       r.SortOrder,
		},
	}
}

type updateExamTypeFieldRequest struct {
	Name            *string `json:"name"`
	InspectionValue *string `json:"inspection_value"`
	NormalValue     *string `json:"normal_value"`
	Unit            *string `json:"unit"`
	SortOrder       *int    `json:"sort_order"`
}

func (r updateExamTypeFieldRequest) toServiceInput() *UpdateExamTypeFieldInput {
	return &UpdateExamTypeFieldInput{
		Name:            r.Name,
		InspectionValue: r.InspectionValue,
		NormalValue:     r.NormalValue,
		Unit:            r.Unit,
		SortOrder:       r.SortOrder,
	}
}

type examReferenceRangeRequest struct {
	AnimalSpeciesID uint64   `json:"animal_species_id"`
	RefMin          *float64 `json:"ref_min"`
	RefMax          *float64 `json:"ref_max"`
	QualitativeMin  *string  `json:"qualitative_min"`
	QualitativeMax  *string  `json:"qualitative_max"`
}

func (r examReferenceRangeRequest) toServiceInput() ReferenceRangeInput {
	return ReferenceRangeInput{
		AnimalSpeciesID: r.AnimalSpeciesID,
		RefMin:          r.RefMin,
		RefMax:          r.RefMax,
		QualitativeMin:  r.QualitativeMin,
		QualitativeMax:  r.QualitativeMax,
	}
}

type replaceExamReferenceRangesRequest struct {
	Ranges *[]examReferenceRangeRequest `json:"ranges"`
}

func (r replaceExamReferenceRangesRequest) toServiceCommand(
	fieldID uint64,
) *ReplaceReferenceRangesCommand {
	inputs := make([]ReferenceRangeInput, 0, len(*r.Ranges))
	for _, item := range *r.Ranges {
		inputs = append(inputs, item.toServiceInput())
	}
	return &ReplaceReferenceRangesCommand{
		ExamTypeFieldID: fieldID,
		Ranges:          inputs,
	}
}
