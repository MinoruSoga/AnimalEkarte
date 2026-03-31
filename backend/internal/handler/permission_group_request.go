package handler

// createPermissionGroupRequest はグループ作成リクエスト
type createPermissionGroupRequest struct {
	Name        string `json:"name"        binding:"required,max=100"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// updatePermissionGroupRequest はグループ更新リクエスト（全フィールドオプション）
type updatePermissionGroupRequest struct {
	Name        *string `json:"name"        binding:"omitempty,max=100"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

// setPermissionGroupRulesRequest はルール一括更新リクエスト
type setPermissionGroupRulesRequest struct {
	Rules []ruleRequest `json:"rules" binding:"required"`
}

// ruleRequest は個別ルールのリクエスト
type ruleRequest struct {
	Resource  string `json:"resource"   binding:"required"`
	CanView   bool   `json:"can_view"`
	CanCreate bool   `json:"can_create"`
	CanEdit   bool   `json:"can_edit"`
	CanDelete bool   `json:"can_delete"`
}
