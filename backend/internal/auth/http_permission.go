package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// HasPermission evaluates one resource/action pair against the selected clinic.
// Context peeks never write a response (AUS-09); RequirePermission owns the single write.
func (h *HTTPHandler) HasPermission(
	c *gin.Context,
	resource, action string,
) bool {
	clinicID, _ := httpapi.PeekClinicID(c)
	return h.HasPermissionInClinic(c, clinicID, resource, action)
}

// HasPermissionInClinic evaluates one resource/action pair for a destination clinic.
// system_admin remains an explicit bypass. Missing identity, a zero clinic, a nil
// EffectivePermissions dependency, or a repository error fail closed.
func (h *HTTPHandler) HasPermissionInClinic(
	c *gin.Context,
	clinicID uint64,
	resource, action string,
) bool {
	isSystemAdmin, ok := httpapi.PeekIsSystemAdmin(c)
	if !ok {
		return false
	}
	if isSystemAdmin {
		return true
	}

	staffID, ok := httpapi.PeekStaffID(c)
	if !ok || clinicID == 0 || h.deps.EffectivePermissions == nil {
		return false
	}
	rules, err := h.deps.EffectivePermissions.GetEffectivePermissions(
		c.Request.Context(),
		staffID,
		clinicID,
	)
	if err != nil {
		return false
	}
	return permissionRulesAllow(rules, resource, action)
}

func permissionRulesAllow(
	rules []model.PermissionGroupRule,
	resource, action string,
) bool {
	for i := range rules {
		rule := &rules[i]
		if rule.Resource != resource {
			continue
		}
		switch action {
		case "view":
			return rule.CanView
		case "create":
			return rule.CanCreate
		case "edit":
			return rule.CanEdit
		case "delete":
			return rule.CanDelete
		}
	}
	return false
}

func (h *HTTPHandler) attachClinicPermissionChecker(c *gin.Context) {
	httpapi.SetClinicPermissionChecker(c, h.HasPermissionInClinic)
}

// RequirePermission rejects requests lacking one resource/action permission.
// Writes still require the grant in the selected clinic. Safe methods (GET/HEAD)
// may proceed when another authorized clinic holds the grant so list/detail
// expansion can Filter/Authorize destination clinics instead of reusing the
// selected clinic's deny.
func (h *HTTPHandler) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.attachClinicPermissionChecker(c)
		if h.HasPermission(c, resource, action) {
			c.Next()
			return
		}
		if isSafeHTTPMethod(c) && h.hasPermissionInAuthorizedClinics(c, resource, action) {
			c.Next()
			return
		}
		httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
		c.Abort()
	}
}

// RequirePermissionAny permits a request when at least one requirement matches.
func (h *HTTPHandler) RequirePermissionAny(
	permissions ...PermissionRequirement,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.attachClinicPermissionChecker(c)
		for _, permission := range permissions {
			if h.HasPermission(c, permission.Resource, permission.Action) {
				c.Next()
				return
			}
		}
		if isSafeHTTPMethod(c) {
			for _, permission := range permissions {
				if h.hasPermissionInAuthorizedClinics(c, permission.Resource, permission.Action) {
					c.Next()
					return
				}
			}
		}
		httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
		c.Abort()
	}
}

func isSafeHTTPMethod(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead:
		return true
	default:
		return false
	}
}

func (h *HTTPHandler) hasPermissionInAuthorizedClinics(c *gin.Context, resource, action string) bool {
	val, exists := c.Get("clinic_ids")
	if !exists {
		return false
	}
	ids, ok := val.([]uint64)
	if !ok {
		return false
	}
	for _, id := range ids {
		if id != 0 && h.HasPermissionInClinic(c, id, resource, action) {
			return true
		}
	}
	return false
}

// RequireDiscountEditFloat enforces discount:edit for changed float values.
func (h *HTTPHandler) RequireDiscountEditFloat(
	c *gin.Context,
	newValue *float64,
	oldValue float64,
) error {
	return httpapi.RequireDiscountEditFloat(c, h.HasPermission, newValue, oldValue)
}

// RequireDiscountEditInt enforces discount:edit for changed integer values.
func (h *HTTPHandler) RequireDiscountEditInt(
	c *gin.Context,
	newValue *int64,
	oldValue int64,
) error {
	return httpapi.RequireDiscountEditInt(c, h.HasPermission, newValue, oldValue)
}

// RequireDiscountCreateFloat enforces discount:create for non-zero float values.
func (h *HTTPHandler) RequireDiscountCreateFloat(
	c *gin.Context,
	value float64,
) error {
	return httpapi.RequireDiscountCreateFloat(c, h.HasPermission, value)
}

// RequireDiscountCreateInt enforces discount:create for non-zero integer values.
func (h *HTTPHandler) RequireDiscountCreateInt(c *gin.Context, value int64) error {
	return httpapi.RequireDiscountCreateInt(c, h.HasPermission, value)
}

// FloatEquals retains the discount epsilon contract.
func FloatEquals(left, right float64) bool {
	return httpapi.FloatEquals(left, right)
}

// CreatePermissionGroupRequest is the live create-body alias for the
// presence-aware PermissionGroupCreateRequest (omitted / false / true).
type CreatePermissionGroupRequest = PermissionGroupCreateRequest

// UpdatePermissionGroupRequest is the permission-group PATCH body.
type UpdatePermissionGroupRequest struct {
	Name        *string                    `json:"name"        binding:"omitempty,min=1,max=255"`
	Description *string                    `json:"description" binding:"omitempty,max=2000"`
	Color       *string                    `json:"color"       binding:"omitempty,hexcolor,max=7"`
	IsActive    *bool                      `json:"is_active"`
	SortOrder   *int                       `json:"sort_order"`
	Rules       []PermissionGroupRuleInput `json:"rules" binding:"omitempty,max=100,dive"`
}

// ToInput maps the request to the auth use-case input.
func (r UpdatePermissionGroupRequest) ToInput() *UpdatePermissionGroupInput {
	return &UpdatePermissionGroupInput{
		Name:        r.Name,
		Description: r.Description,
		Color:       r.Color,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
		Rules:       permissionGroupRuleInputs(r.Rules),
	}
}

// SetPermissionGroupRulesRequest replaces every rule in one permission group.
type SetPermissionGroupRulesRequest struct {
	Rules []PermissionGroupRuleInput `json:"rules" binding:"required,min=1,max=100,dive"`
}

// PermissionGroupRuleInput is one replacement rule.
type PermissionGroupRuleInput struct {
	Resource  string `json:"resource"   binding:"required,min=1,max=50"`
	CanView   bool   `json:"can_view"`
	CanCreate bool   `json:"can_create"`
	CanEdit   bool   `json:"can_edit"`
	CanDelete bool   `json:"can_delete"`
}

// ReorderPermissionGroupsRequest bounds the auth-owned reorder collection.
type ReorderPermissionGroupsRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1,max=100"`
}

// ToInput maps the request to auth use-case rule inputs.
func (r SetPermissionGroupRulesRequest) ToInput() []SetPermissionGroupRulesInput {
	return permissionGroupRuleInputs(r.Rules)
}

func permissionGroupRuleInputs(
	rules []PermissionGroupRuleInput,
) []SetPermissionGroupRulesInput {
	if rules == nil {
		return nil
	}
	inputs := make([]SetPermissionGroupRulesInput, 0, len(rules))
	for _, rule := range rules {
		inputs = append(inputs, SetPermissionGroupRulesInput{
			Resource:  rule.Resource,
			CanView:   rule.CanView,
			CanCreate: rule.CanCreate,
			CanEdit:   rule.CanEdit,
			CanDelete: rule.CanDelete,
		})
	}
	return inputs
}

// PermissionGroupResponse is the public permission-group representation.
type PermissionGroupResponse struct {
	ID          uint64                        `json:"id"`
	ClinicID    uint64                        `json:"clinic_id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Color       string                        `json:"color"`
	IsActive    bool                          `json:"is_active"`
	SortOrder   int                           `json:"sort_order"`
	Rules       []PermissionGroupRuleResponse `json:"rules,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

// PermissionGroupRuleResponse is one public permission rule.
type PermissionGroupRuleResponse struct {
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

// ToPermissionGroupResponse maps a permission group and its rules.
func ToPermissionGroupResponse(group *model.PermissionGroup) PermissionGroupResponse {
	rules := make([]PermissionGroupRuleResponse, 0, len(group.Rules))
	for i := range group.Rules {
		rules = append(rules, ToPermissionGroupRuleResponse(&group.Rules[i]))
	}
	return PermissionGroupResponse{
		ID:          group.ID,
		ClinicID:    group.ClinicID,
		Name:        group.Name,
		Description: group.Description,
		Color:       group.Color,
		IsActive:    group.IsActive,
		SortOrder:   group.SortOrder,
		Rules:       rules,
		CreatedAt:   httpapi.LocalTime(group.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(group.UpdatedAt),
	}
}

// ToPermissionGroupRuleResponse maps one permission group rule.
func ToPermissionGroupRuleResponse(
	rule *model.PermissionGroupRule,
) PermissionGroupRuleResponse {
	return PermissionGroupRuleResponse{
		ID:        rule.ID,
		GroupID:   rule.GroupID,
		Resource:  rule.Resource,
		CanView:   rule.CanView,
		CanCreate: rule.CanCreate,
		CanEdit:   rule.CanEdit,
		CanDelete: rule.CanDelete,
		CreatedAt: httpapi.LocalTime(rule.CreatedAt),
		UpdatedAt: httpapi.LocalTime(rule.UpdatedAt),
	}
}

// RegisterPermissionGroupRoutes registers the clinic-scoped RBAC master surface.
func (h *HTTPHandler) RegisterPermissionGroupRoutes(masters *gin.RouterGroup) {
	resource := string(model.ResourceMasterPermission)
	masters.GET(
		"/permission-groups",
		h.RequirePermission(resource, "view"),
		h.ListPermissionGroups,
	)
	masters.GET(
		"/permission-groups/:id",
		h.RequirePermission(resource, "view"),
		h.GetPermissionGroup,
	)
	masters.POST(
		"/permission-groups",
		h.RequirePermission(resource, "create"),
		h.CreatePermissionGroup,
	)
	masters.PATCH(
		"/permission-groups/reorder",
		h.RequirePermission(resource, "edit"),
		h.ReorderPermissionGroups,
	)
	masters.PATCH(
		"/permission-groups/:id",
		h.RequirePermission(resource, "edit"),
		h.UpdatePermissionGroup,
	)
	masters.DELETE(
		"/permission-groups/:id",
		h.RequirePermission(resource, "delete"),
		h.DeletePermissionGroup,
	)
	masters.PUT(
		"/permission-groups/:id/rules",
		h.RequirePermission(resource, "edit"),
		h.SetPermissionGroupRules,
	)
}

// GetPermissionGroup returns one clinic-owned permission group.
func (h *HTTPHandler) GetPermissionGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	group, err := h.deps.PermissionGroups.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToPermissionGroupResponse(group))
}

// ListPermissionGroups returns clinic-owned permission groups.
func (h *HTTPHandler) ListPermissionGroups(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	groups, err := h.deps.PermissionGroups.List(c.Request.Context(), clinicID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(groups, ToPermissionGroupResponse))
}

// CreatePermissionGroup creates one clinic-owned permission group.
func (h *HTTPHandler) CreatePermissionGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var request CreatePermissionGroupRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if request.Rules != nil &&
		!h.HasPermission(
			c,
			string(model.ResourceMasterPermission),
			"edit",
		) {
		httpapi.RespondError(c, apperrors.WrapForbidden("forbidden"))
		return
	}
	audit, ok := permissionMutationAuditFromContext(
		c,
		clinicID,
		model.AuditActionPermissionGroupCreate,
		"permission_group",
	)
	if !ok {
		return
	}
	group, err := h.deps.PermissionGroups.Create(
		c.Request.Context(),
		clinicID,
		request.ToInput(),
		audit,
	)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/v1/masters/permission-groups/%d", group.ID))
	c.JSON(http.StatusCreated, ToPermissionGroupResponse(group))
}

// UpdatePermissionGroup updates one clinic-owned permission group.
func (h *HTTPHandler) UpdatePermissionGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var request UpdatePermissionGroupRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	audit, ok := permissionMutationAuditFromContext(
		c,
		clinicID,
		model.AuditActionPermissionGroupUpdate,
		"permission_group",
	)
	if !ok {
		return
	}
	updated, err := h.deps.PermissionGroups.Update(
		c.Request.Context(),
		clinicID,
		id,
		request.ToInput(),
		audit,
	)
	if err != nil {
		httpapi.RespondErrorPreferringConflictCode(c, err)
		return
	}

	c.JSON(http.StatusOK, ToPermissionGroupResponse(updated))
}

// DeletePermissionGroup deletes one clinic-owned permission group.
func (h *HTTPHandler) DeletePermissionGroup(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}
	audit, ok := permissionMutationAuditFromContext(
		c,
		clinicID,
		model.AuditActionPermissionGroupDelete,
		"permission_group",
	)
	if !ok {
		return
	}
	if err := h.deps.PermissionGroups.Delete(
		c.Request.Context(),
		clinicID,
		id,
		audit,
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SetPermissionGroupRules replaces one clinic-owned group's rules.
func (h *HTTPHandler) SetPermissionGroupRules(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "id")
	if !ok {
		return
	}

	var request SetPermissionGroupRulesRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	audit, ok := permissionMutationAuditFromContext(
		c,
		clinicID,
		model.AuditActionPermissionRulesUpdate,
		"permission_group_rules",
	)
	if !ok {
		return
	}
	rules := request.ToInput()
	group, err := h.deps.PermissionGroups.UpdateRules(
		c.Request.Context(),
		clinicID,
		id,
		rules,
		staffID,
		audit,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToPermissionGroupResponse(group))
}

// ReorderPermissionGroups replaces clinic-owned permission-group sort order.
func (h *HTTPHandler) ReorderPermissionGroups(c *gin.Context) {
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	var request ReorderPermissionGroupsRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := h.deps.PermissionGroups.Reorder(
		c.Request.Context(),
		clinicID,
		request.IDs,
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func permissionMutationAuditFromContext(
	c *gin.Context,
	clinicID uint64,
	action, resource string,
) (PermissionMutationAudit, bool) {
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return PermissionMutationAudit{}, false
	}
	return PermissionMutationAudit{
		ClinicID:     clinicID,
		ActorStaffID: staffID,
		Action:       action,
		Resource:     resource,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.Header.Get("User-Agent"),
	}, true
}
