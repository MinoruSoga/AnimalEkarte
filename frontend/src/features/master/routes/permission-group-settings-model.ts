import type {
  CreatePermissionGroupRequest,
  UpdatePermissionGroupRequest,
} from "../api/permission-groups";
import type { PermissionGroupFormData } from "../components/permission-group-side-panel-model";

function buildPermissionGroupRuleRequests(data: PermissionGroupFormData) {
  return data.rules.map((rule) => ({
    resource: rule.resource,
    can_view: rule.canView,
    can_create: rule.canCreate,
    can_edit: rule.canEdit,
    can_delete: rule.canDelete,
  }));
}

export function buildPermissionGroupCreateRequest(
  data: PermissionGroupFormData,
): CreatePermissionGroupRequest {
  return {
    name: data.name,
    description: data.description,
    color: data.color,
    is_active: data.isActive,
    rules: buildPermissionGroupRuleRequests(data),
  };
}

export function buildPermissionGroupUpdateRequest(
  data: PermissionGroupFormData,
): UpdatePermissionGroupRequest {
  return buildPermissionGroupCreateRequest(data);
}
