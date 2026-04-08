package handler

type createPermissionGroupRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	Color       string `json:"color"       binding:"required"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type updatePermissionGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
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

// reorderPermissionGroupRequest は権限グループ並び替えリクエスト
type reorderPermissionGroupRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
