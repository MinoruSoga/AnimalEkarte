package lstep

// updateTriggerPriorityItemRequest は PATCH /lstep/trigger-priorities のリクエスト内1件。
type updateTriggerPriorityItemRequest struct {
	TriggerType string `json:"trigger_type" binding:"required"`
	Priority    int    `json:"priority"     binding:"required,min=1"`
}

// updateTriggerPrioritiesRequest は PATCH /lstep/trigger-priorities のリクエストボディ。
type updateTriggerPrioritiesRequest struct {
	Items []updateTriggerPriorityItemRequest `json:"items" binding:"required,min=1,dive"`
}

func (r updateTriggerPrioritiesRequest) toServiceInput() UpdateTriggerPrioritiesInput {
	items := make([]TriggerPriorityItem, len(r.Items))
	for i, item := range r.Items {
		items[i] = TriggerPriorityItem(item)
	}
	return UpdateTriggerPrioritiesInput{Items: items}
}
