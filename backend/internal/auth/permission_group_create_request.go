package auth

// PermissionGroupCreateRequest is a presence-aware create body for permission groups.
// IsActive is *bool so JSON binding can distinguish omitted / false / true.
// This DTO is intentionally unconnected from the live handler; wire-up is AUTH-B.
type PermissionGroupCreateRequest struct {
	Name        string                     `json:"name"        binding:"required,min=1,max=255"`
	Description string                     `json:"description" binding:"max=2000"`
	Color       string                     `json:"color"       binding:"required,hexcolor,max=7"`
	IsActive    *bool                      `json:"is_active"`
	SortOrder   int                        `json:"sort_order"`
	Rules       []PermissionGroupRuleInput `json:"rules" binding:"omitempty,max=100,dive"`
}
