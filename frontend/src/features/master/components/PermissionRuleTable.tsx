import { memo } from "react";

import { TableHead } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";

import {
  ALL_PERMISSION_RESOURCES,
  PERMISSION_ACTION_COLUMNS,
  buildPermissionRuleMap,
  createEmptyPermissionRule,
  type PermissionRule,
  type PermissionRuleField,
} from "./permission-rule-table-model";
import { PermissionRuleTableRow } from "./PermissionRuleTableRow";

interface PermissionRuleTableProps {
  rules: PermissionRule[];
  onRuleChange: (resource: string, field: PermissionRuleField, value: boolean) => void;
  disabled?: boolean;
}

export const PermissionRuleTable = memo(function PermissionRuleTable({
  rules,
  onRuleChange,
  disabled,
}: PermissionRuleTableProps) {
  const ruleMap = buildPermissionRuleMap(rules);

  return (
    <div className="mt-6">
      <h3 className={`text-sm font-semibold ${C.text} mb-3`}>権限設定</h3>
      <div className="overflow-x-auto border rounded-lg">
        <table className="w-full">
          <thead>
            <tr className={`border-b ${C.borderLight}`}>
              <TableHead className={C.text50}>
                リソース
              </TableHead>
              {PERMISSION_ACTION_COLUMNS.map(({ field, label }) => (
                <TableHead
                  key={field}
                  className={`text-center ${C.text50}`}
                >
                  {label}
                </TableHead>
              ))}
            </tr>
          </thead>
          <tbody>
            {ALL_PERMISSION_RESOURCES.map((resource) => (
              <PermissionRuleTableRow
                key={resource}
                resource={resource}
                rule={ruleMap.get(resource) ?? createEmptyPermissionRule(resource)}
                onRuleChange={onRuleChange}
                disabled={disabled}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
});
