package auth

// PermissionGroupCreateRequest is the presence-aware create body for permission groups.
// IsActive is *bool so JSON binding can distinguish omitted / false / true.
// Live handler binds this type (via CreatePermissionGroupRequest alias).
type PermissionGroupCreateRequest struct {
	Name        string                     `json:"name"        binding:"required,min=1,max=255"`
	Description string                     `json:"description" binding:"max=2000"`
	Color       string                     `json:"color"       binding:"required,hexcolor,max=7"`
	IsActive    *bool                      `json:"is_active"`
	SortOrder   int                        `json:"sort_order"`
	Rules       []PermissionGroupRuleInput `json:"rules" binding:"omitempty,max=100,dive"`
}

// ToInput maps the request to the auth use-case input.
// Omitted is_active (nil) resolves to true; explicit false/true are preserved.
func (r PermissionGroupCreateRequest) ToInput() *CreatePermissionGroupInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &CreatePermissionGroupInput{
		Name:        r.Name,
		Description: r.Description,
		Color:       r.Color,
		IsActive:    isActive,
		SortOrder:   r.SortOrder,
		Rules:       permissionGroupRuleInputs(r.Rules),
	}
}
