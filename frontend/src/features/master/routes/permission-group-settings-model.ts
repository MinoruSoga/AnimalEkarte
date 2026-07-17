import type {
  CreatePermissionGroupRequest,
  UpdatePermissionGroupRequest,
  UpdatePermissionGroupRulesRequest,
} from "../api/permission-groups";
import type { PermissionGroupFormData } from "../components/permission-group-side-panel-model";

export function buildPermissionGroupCreateRequest(
  data: PermissionGroupFormData,
): CreatePermissionGroupRequest {
  return {
    name: data.name,
    description: data.description,
    color: data.color,
    is_active: data.isActive,
  };
}

export function buildPermissionGroupUpdateRequest(
  data: PermissionGroupFormData,
): UpdatePermissionGroupRequest {
  return buildPermissionGroupCreateRequest(data);
}

export function buildPermissionGroupRulesRequest(
  data: PermissionGroupFormData,
): UpdatePermissionGroupRulesRequest {
  return {
    rules: data.rules.map((rule) => ({
      resource: rule.resource,
      can_view: rule.canView,
      can_create: rule.canCreate,
      can_edit: rule.canEdit,
      can_delete: rule.canDelete,
    })),
  };
}
