package lstep

type createAutoManagedPrefixRequest struct {
	Prefix string `json:"prefix" binding:"required"`
	// Category is constrained to DDL-documented values (LSA-14); C2 drives pet-death tag cleanup.
	Category    string  `json:"category" binding:"required,oneof=B C1 C2 C3"`
	Description *string `json:"description"`
}

func (r createAutoManagedPrefixRequest) toServiceInput() CreateAutoManagedPrefixInput {
	return CreateAutoManagedPrefixInput(r)
}

type createConditionTagMappingRequest struct {
	// condition_code matches DDL lstep_condition_tag_mappings.condition_code VARCHAR(50).
	// max=50 rejects overflow at bind time (BUG-025: DB length violation was leaking as 500).
	ConditionCode string  `json:"condition_code" binding:"required,max=50"`
	TagName       string  `json:"tag_name"       binding:"required"`
	Description   *string `json:"description"`
}

func (r createConditionTagMappingRequest) toServiceInput() CreateConditionTagMappingInput {
	return CreateConditionTagMappingInput(r)
}

type createSendPurposeTagPrefixRequest struct {
	Purpose     string  `json:"purpose"     binding:"required"`
	TagPrefix   string  `json:"tag_prefix"  binding:"required"`
	Description *string `json:"description"`
}

func (r createSendPurposeTagPrefixRequest) toServiceInput() CreateSendPurposeTagPrefixInput {
	return CreateSendPurposeTagPrefixInput(r)
}
