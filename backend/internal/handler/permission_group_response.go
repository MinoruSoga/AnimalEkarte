package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type permissionGroupResponse struct {
	ID          uint64                        `json:"id"`
	ClinicID    uint64                        `json:"clinic_id"`
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
	ID        uint64    `json:"id"`
	GroupID   uint64    `json:"group_id"`
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
		ID:          pg.ID,
		ClinicID:    pg.ClinicID,
		Name:        pg.Name,
		Description: pg.Description,
		Color:       pg.Color,
		IsActive:    pg.IsActive,
		SortOrder:   pg.SortOrder,
		Rules:       rules,
		CreatedAt:   localTime(pg.CreatedAt),
		UpdatedAt:   localTime(pg.UpdatedAt),
	}
}

func toPermissionGroupRuleResponse(rule *model.PermissionGroupRule) permissionGroupRuleResponse {
	return permissionGroupRuleResponse{
		ID:        rule.ID,
		GroupID:   rule.GroupID,
		Resource:  rule.Resource,
		CanView:   rule.CanView,
		CanCreate: rule.CanCreate,
		CanEdit:   rule.CanEdit,
		CanDelete: rule.CanDelete,
		CreatedAt: localTime(rule.CreatedAt),
		UpdatedAt: localTime(rule.UpdatedAt),
	}
}
