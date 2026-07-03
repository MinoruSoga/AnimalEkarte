package handler

import "github.com/animal-ekarte/backend/internal/service"

type createPermissionGroupRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Color       string `json:"color"       binding:"required"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

func (r createPermissionGroupRequest) toServiceInput() *service.CreatePermissionGroupInput {
	return &service.CreatePermissionGroupInput{
		Name:        r.Name,
		Description: r.Description,
		Color:       r.Color,
		IsActive:    r.IsActive,
		SortOrder:   r.SortOrder,
	}
}

type updatePermissionGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updatePermissionGroupRequest) toServiceInput() *service.UpdatePermissionGroupInput {
	return &service.UpdatePermissionGroupInput{
		Name:        r.Name,
		Description: r.Description,
		Color:       r.Color,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
	}
}

// setPermissionGroupRulesRequest は権限グループのルール設定リクエスト
type setPermissionGroupRulesRequest struct {
	Rules []permissionGroupRuleInput `json:"rules" binding:"required"`
}

type permissionGroupRuleInput struct {
	Resource  string `json:"resource"   binding:"required,min=1,max=50"`
	CanView   bool   `json:"can_view"`
	CanCreate bool   `json:"can_create"`
	CanEdit   bool   `json:"can_edit"`
	CanDelete bool   `json:"can_delete"`
}

func (r setPermissionGroupRulesRequest) toServiceInput() []service.SetPermissionGroupRulesInput {
	inputRules := make([]service.SetPermissionGroupRulesInput, 0, len(r.Rules))
	for _, rule := range r.Rules {
		inputRules = append(inputRules, service.SetPermissionGroupRulesInput{
			Resource:  rule.Resource,
			CanView:   rule.CanView,
			CanCreate: rule.CanCreate,
			CanEdit:   rule.CanEdit,
			CanDelete: rule.CanDelete,
		})
	}
	return inputRules
}
