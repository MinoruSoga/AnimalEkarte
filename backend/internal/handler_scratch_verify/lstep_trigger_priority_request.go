package handler

import "github.com/animal-ekarte/backend/internal/service"

// updateTriggerPriorityItemRequest は PATCH /lstep/trigger-priorities のリクエスト内1件。
type updateTriggerPriorityItemRequest struct {
	TriggerType string `json:"trigger_type" binding:"required"`
	Priority    int    `json:"priority"     binding:"required,min=1"`
}

// updateTriggerPrioritiesRequest は PATCH /lstep/trigger-priorities のリクエストボディ。
type updateTriggerPrioritiesRequest struct {
	Items []updateTriggerPriorityItemRequest `json:"items" binding:"required,min=1,dive"`
}

func (r updateTriggerPrioritiesRequest) toServiceInput() service.UpdateTriggerPrioritiesInput {
	items := make([]service.TriggerPriorityItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = service.TriggerPriorityItem{
			TriggerType: item.TriggerType,
			Priority:    item.Priority,
		}
	}
	return service.UpdateTriggerPrioritiesInput{Items: items}
}
