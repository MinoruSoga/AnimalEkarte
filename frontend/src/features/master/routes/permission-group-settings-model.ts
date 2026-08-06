import type {
  CreatePermissionGroupRequest,
  PermissionGroup,
  UpdatePermissionGroupRequest,
} from "../api/permission-groups";
import type { PermissionGroupFormData } from "../components/permission-group-side-panel-model";
import {
  ALL_PERMISSION_RESOURCES,
  createEmptyPermissionRule,
  type PermissionRule,
} from "../components/permission-rule-table-model";

type PermissionGroupRuleRequest = NonNullable<
  CreatePermissionGroupRequest["rules"]
>[number];

/**
 * Build the complete rule set the matrix UI shows.
 * Overlay form edits on every known resource, and keep any extra resources
 * returned from the API so replace-all cannot silently drop them.
 */
export function expandPermissionGroupRules(
  formRules: PermissionRule[],
): PermissionRule[] {
  const byResource = new Map<string, PermissionRule>();
  for (const resource of ALL_PERMISSION_RESOURCES) {
    byResource.set(resource, createEmptyPermissionRule(resource));
  }
  for (const rule of formRules) {
    byResource.set(rule.resource, {
      resource: rule.resource,
      canView: rule.canView === true,
      canCreate: rule.canCreate === true,
      canEdit: rule.canEdit === true,
      canDelete: rule.canDelete === true,
    });
  }
  return Array.from(byResource.values());
}

function buildPermissionGroupRuleRequests(
  data: PermissionGroupFormData,
): PermissionGroupRuleRequest[] {
  return expandPermissionGroupRules(data.rules).map((rule) => ({
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

/**
 * BUG-024: reject false success when PATCH returns parent fields (updated_at)
 * but the rule matrix did not round-trip.
 */
export function permissionRulesMatchRequest(
  requested: PermissionGroupRuleRequest[] | undefined,
  saved: PermissionGroup["rules"],
): boolean {
  if (requested === undefined) {
    return true;
  }
  if (requested.length !== saved.length) {
    return false;
  }
  const savedByResource = new Map(saved.map((rule) => [rule.resource, rule]));
  for (const req of requested) {
    const savedRule = savedByResource.get(req.resource);
    if (!savedRule) {
      return false;
    }
    if (
      savedRule.canView !== req.can_view ||
      savedRule.canCreate !== req.can_create ||
      savedRule.canEdit !== req.can_edit ||
      savedRule.canDelete !== req.can_delete
    ) {
      return false;
    }
  }
  return true;
}

export function assertSavedPermissionRulesMatch(
  requested: PermissionGroupRuleRequest[] | undefined,
  saved: PermissionGroup,
): void {
  if (!permissionRulesMatchRequest(requested, saved.rules)) {
    throw new Error(
      "権限マトリクスの保存結果がサーバ応答と一致しません。再読み込みして確認してください。",
    );
  }
}
