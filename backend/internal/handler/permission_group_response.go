package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type permissionGroupResponse struct {
	ID          string                        `json:"id"`
	ClinicID    string                        `json:"clinic_id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Color       string                        `json:"color"`
	IsActive    bool                          `json:"is_active"`
	SortOrder   int                           `json:"sort_order"`
	Rules       []permissionGroupRuleResponse `json:"rules,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

type permissionGroupRuleResponse struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	Resource  string    `json:"resource"`
	CanView   bool      `json:"can_view"`
	CanCreate bool      `json:"can_create"`
	CanEdit   bool      `json:"can_edit"`
	CanDelete bool      `json:"can_delete"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toPermissionGroupResponse(pg *model.PermissionGroup) permissionGroupResponse {
	rules := make([]permissionGroupRuleResponse, 0)
	if len(pg.Rules) > 0 {
		for i := range pg.Rules {
			rules = append(rules, toPermissionGroupRuleResponse(&pg.Rules[i]))
		}
	}
	return permissionGroupResponse{
		ID:          strconv.FormatUint(pg.ID, 10),
		ClinicID:    strconv.FormatUint(pg.ClinicID, 10),
		Name:        pg.Name,
		Description: pg.Description,
		Color:       pg.Color,
		IsActive:    pg.IsActive,
		SortOrder:   pg.SortOrder,
		Rules:       rules,
		CreatedAt:   pg.CreatedAt,
		UpdatedAt:   pg.UpdatedAt,
	}
}

func toPermissionGroupResponseList(items []model.PermissionGroup) []permissionGroupResponse {
	list := make([]permissionGroupResponse, 0, len(items))
	for i := range items {
		list = append(list, toPermissionGroupResponse(&items[i]))
	}
	return list
}

func toPermissionGroupRuleResponse(rule *model.PermissionGroupRule) permissionGroupRuleResponse {
	return permissionGroupRuleResponse{
		ID:        strconv.FormatUint(rule.ID, 10),
		GroupID:   strconv.FormatUint(rule.GroupID, 10),
		Resource:  rule.Resource,
		CanView:   rule.CanView,
		CanCreate: rule.CanCreate,
		CanEdit:   rule.CanEdit,
		CanDelete: rule.CanDelete,
		CreatedAt: rule.CreatedAt,
		UpdatedAt: rule.UpdatedAt,
	}
}
