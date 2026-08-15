import { Checkbox } from "@/components/ui/checkbox";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { C } from "@/lib/design-tokens";

import {
  PERMISSION_ACTION_COLUMNS,
  RESOURCE_LABELS,
  type PermissionRule,
  type PermissionRuleField,
} from "./permission-rule-table-model";

interface PermissionRuleTableRowProps {
  resource: string;
  rule: PermissionRule;
  onRuleChange: (resource: string, field: PermissionRuleField, value: boolean) => void;
  disabled?: boolean;
}

export function PermissionRuleTableRow({
  resource,
  rule,
  onRuleChange,
  disabled,
}: PermissionRuleTableRowProps) {
  return (
    <DataTableRow>
      <TableCell className={`text-sm ${C.text}`}>
        {RESOURCE_LABELS[resource] || resource}
      </TableCell>
      {PERMISSION_ACTION_COLUMNS.map(({ field, label }) => (
        <PermissionCheckboxCell
          key={field}
          accessibleName={`${RESOURCE_LABELS[resource] || resource} ${label}`}
          checked={rule[field]}
          disabled={disabled}
          onChange={(checked) => onRuleChange(resource, field, checked)}
        />
      ))}
    </DataTableRow>
  );
}

interface PermissionCheckboxCellProps {
  accessibleName: string;
  checked: boolean;
  disabled?: boolean;
  onChange: (checked: boolean) => void;
}

function PermissionCheckboxCell({
  accessibleName,
  checked,
  disabled,
  onChange,
}: PermissionCheckboxCellProps) {
  return (
    <TableCell className="text-center">
      <Checkbox
        aria-label={accessibleName}
        checked={checked}
        disabled={disabled}
        onCheckedChange={(value) => onChange(value === true)}
      />
    </TableCell>
  );
}
