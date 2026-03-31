import { memo, useCallback } from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { PERMISSION_RESOURCES } from "@/features/master/types/permission-resources";
import type { RuleInput } from "@/features/master/api/permission-groups/permission-groups";

interface PermissionRuleTableProps {
  rules: RuleInput[];
  onChange: (rules: RuleInput[]) => void;
}

type ActionKey = "can_view" | "can_create" | "can_edit" | "can_delete";

const ACTION_COLUMNS: { key: ActionKey; label: string }[] = [
  { key: "can_view", label: "表示" },
  { key: "can_create", label: "登録" },
  { key: "can_edit", label: "編集" },
  { key: "can_delete", label: "削除" },
];

const TABLE_HEADER = (
  <thead>
    <tr className="border-b">
      <th className="text-left py-2 pr-4 font-medium text-sm text-muted-foreground">ページ</th>
      {ACTION_COLUMNS.map((col) => (
        <th
          key={col.key}
          className="text-center w-16 font-medium text-sm text-muted-foreground pb-2"
        >
          {col.label}
        </th>
      ))}
    </tr>
  </thead>
);

export const PermissionRuleTable = memo(function PermissionRuleTable({
  rules,
  onChange,
}: PermissionRuleTableProps) {
  const handleChange = useCallback(
    (resource: string, action: ActionKey, value: boolean) => {
      onChange(
        rules.map((r) =>
          r.resource === resource ? { ...r, [action]: value } : r,
        ),
      );
    },
    [rules, onChange],
  );

  return (
    <table className="w-full text-sm">
      {TABLE_HEADER}
      <tbody>
        {PERMISSION_RESOURCES.map((res) => {
          const rule = rules.find((r) => r.resource === res.key) ?? {
            resource: res.key,
            can_view: false,
            can_create: false,
            can_edit: false,
            can_delete: false,
          };
          return (
            <tr key={res.key} className="border-b last:border-0">
              <td className="py-2.5 pr-4 text-sm">{res.label}</td>
              {ACTION_COLUMNS.map((col) => (
                <td key={col.key} className="text-center">
                  <Checkbox
                    checked={rule[col.key]}
                    onCheckedChange={(v) => handleChange(res.key, col.key, v === true)}
                  />
                </td>
              ))}
            </tr>
          );
        })}
      </tbody>
    </table>
  );
});
