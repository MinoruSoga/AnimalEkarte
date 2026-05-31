package handler

import "github.com/animal-ekarte/backend/internal/service"

type createAutoManagedPrefixRequest struct {
	Prefix      string  `json:"prefix"       binding:"required"`
	Category    string  `json:"category"     binding:"required"`
	Description *string `json:"description"`
}

func (r createAutoManagedPrefixRequest) toServiceInput() service.CreateAutoManagedPrefixInput {
	return service.CreateAutoManagedPrefixInput{
		Prefix:      r.Prefix,
		Category:    r.Category,
		Description: r.Description,
	}
}

type createConditionTagMappingRequest struct {
	ConditionCode string  `json:"condition_code" binding:"required"`
	TagName       string  `json:"tag_name"       binding:"required"`
	Description   *string `json:"description"`
}

func (r createConditionTagMappingRequest) toServiceInput() service.CreateConditionTagMappingInput {
	return service.CreateConditionTagMappingInput{
		ConditionCode: r.ConditionCode,
		TagName:       r.TagName,
		Description:   r.Description,
	}
}

type createSendPurposeTagPrefixRequest struct {
	Purpose     string  `json:"purpose"     binding:"required"`
	TagPrefix   string  `json:"tag_prefix"  binding:"required"`
	Description *string `json:"description"`
}

func (r createSendPurposeTagPrefixRequest) toServiceInput() service.CreateSendPurposeTagPrefixInput {
	return service.CreateSendPurposeTagPrefixInput{
		Purpose:     r.Purpose,
		TagPrefix:   r.TagPrefix,
		Description: r.Description,
	}
}
