import { memo } from "react";

import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C } from "@/lib/design-tokens";

import {
  ALL_PERMISSION_RESOURCES,
  PERMISSION_ACTION_COLUMNS,
  buildPermissionRuleMap,
  createEmptyPermissionRule,
  type PermissionRule,
  type PermissionRuleField,
} from "../lib/permission-rule-table-model";
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
      <h3 className={`flex items-center gap-1 text-sm font-semibold ${C.text} mb-3`}>
        権限設定
        <FieldHelp
          label="権限設定"
          content="この権限グループに許可する操作です。リソース（画面・機能）ごとに、チェックで各操作権限を付与します。"
        />
      </h3>
      <Table className="border rounded-lg">
        <TableHeader>
          <TableRow className={C.borderLight}>
            <TableHead className={C.text50}>リソース</TableHead>
            {PERMISSION_ACTION_COLUMNS.map(({ field, label }) => (
              <TableHead key={field} className={`text-center ${C.text50}`}>
                {label}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {ALL_PERMISSION_RESOURCES.map((resource) => (
            <PermissionRuleTableRow
              key={resource}
              resource={resource}
              rule={ruleMap.get(resource) ?? createEmptyPermissionRule(resource)}
              onRuleChange={onRuleChange}
              disabled={disabled}
            />
          ))}
        </TableBody>
      </Table>
    </div>
  );
});
