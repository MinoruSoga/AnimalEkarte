package lstep

type createAutoManagedPrefixRequest struct {
	Prefix      string  `json:"prefix"       binding:"required"`
	Category    string  `json:"category"     binding:"required"`
	Description *string `json:"description"`
}

func (r createAutoManagedPrefixRequest) toServiceInput() CreateAutoManagedPrefixInput {
	return CreateAutoManagedPrefixInput(r)
}

type createConditionTagMappingRequest struct {
	ConditionCode string  `json:"condition_code" binding:"required"`
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
