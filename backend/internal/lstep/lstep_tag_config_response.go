package lstep

import "github.com/animal-ekarte/backend/internal/model"

type autoManagedPrefixResponse struct {
	ID          uint64  `json:"id"`
	Prefix      string  `json:"prefix"`
	Category    string  `json:"category"`
	Description *string `json:"description"`
}

type conditionTagMappingResponse struct {
	ID            uint64  `json:"id"`
	ConditionCode string  `json:"condition_code"`
	TagName       string  `json:"tag_name"`
	Description   *string `json:"description"`
}

type sendPurposeTagPrefixResponse struct {
	ID          uint64  `json:"id"`
	Purpose     string  `json:"purpose"`
	TagPrefix   string  `json:"tag_prefix"`
	Description *string `json:"description"`
}

func toAutoManagedPrefixResponse(m *model.LstepAutoManagedPrefix) autoManagedPrefixResponse {
	return autoManagedPrefixResponse{
		ID:          m.ID,
		Prefix:      m.Prefix,
		Category:    m.Category,
		Description: m.Description,
	}
}

func toAutoManagedPrefixListResponse(ms []*model.LstepAutoManagedPrefix) []autoManagedPrefixResponse {
	out := make([]autoManagedPrefixResponse, len(ms))
	for i, m := range ms {
		out[i] = toAutoManagedPrefixResponse(m)
	}
	return out
}

func toConditionTagMappingResponse(m *model.LstepConditionTagMapping) conditionTagMappingResponse {
	return conditionTagMappingResponse{
		ID:            m.ID,
		ConditionCode: m.ConditionCode,
		TagName:       m.TagName,
		Description:   m.Description,
	}
}

func toConditionTagMappingListResponse(ms []*model.LstepConditionTagMapping) []conditionTagMappingResponse {
	out := make([]conditionTagMappingResponse, len(ms))
	for i, m := range ms {
		out[i] = toConditionTagMappingResponse(m)
	}
	return out
}

func toSendPurposeTagPrefixResponse(m *model.LstepSendPurposeTagPrefix) sendPurposeTagPrefixResponse {
	return sendPurposeTagPrefixResponse{
		ID:          m.ID,
		Purpose:     m.Purpose,
		TagPrefix:   m.TagPrefix,
		Description: m.Description,
	}
}

func toSendPurposeTagPrefixListResponse(ms []*model.LstepSendPurposeTagPrefix) []sendPurposeTagPrefixResponse {
	out := make([]sendPurposeTagPrefixResponse, len(ms))
	for i, m := range ms {
		out[i] = toSendPurposeTagPrefixResponse(m)
	}
	return out
}
