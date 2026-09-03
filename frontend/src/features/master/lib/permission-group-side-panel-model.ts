import { PALETTE } from "@/lib/design-tokens";

import type { PermissionGroup } from "../api/permission-groups";
import type { PermissionRule, PermissionRuleField } from "./permission-rule-table-model";

export interface PermissionGroupFormData {
  name: string;
  description: string;
  color: string;
  isActive: boolean;
  rules: PermissionRule[];
}

export function permissionGroupToFormData(item: PermissionGroup | null): PermissionGroupFormData {
  return {
    name: item?.name ?? "",
    description: item?.description ?? "",
    color: item?.color ?? PALETTE.pickerDefaultGray,
    isActive: item?.isActive ?? true,
    rules:
      item?.rules?.map((rule) => ({
        resource: rule.resource,
        canView: rule.canView,
        canCreate: rule.canCreate,
        canEdit: rule.canEdit,
        canDelete: rule.canDelete,
      })) ?? [],
  };
}

export function updatePermissionRule(
  rules: PermissionRule[],
  resource: string,
  field: PermissionRuleField,
  value: boolean,
): PermissionRule[] {
  const existingRule = rules.find((rule) => rule.resource === resource);

  if (existingRule) {
    return rules.map((rule) => (rule.resource === resource ? { ...rule, [field]: value } : rule));
  }

  return [
    ...rules,
    {
      resource,
      canView: field === "canView" ? value : false,
      canCreate: field === "canCreate" ? value : false,
      canEdit: field === "canEdit" ? value : false,
      canDelete: field === "canDelete" ? value : false,
    },
  ];
}
