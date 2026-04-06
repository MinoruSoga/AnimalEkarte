import { memo } from "react";
import { Checkbox } from "@/components/ui/checkbox";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { C } from "@/lib/design-tokens";
import type { PermissionGroup } from "@/features/master/api/permission-groups";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const RESOURCE_LABELS: Record<string, string> = {
  dashboard: "ダッシュボード",
  owners: "オーナー",
  reservations: "予約",
  "medical-records": "電子カルテ",
  hospitalization: "入院管理",
  trimming: "トリミング",
  examinations: "診察",
  accounting: "会計",
  vaccinations: "ワクチン",
  checkups: "検査",
  inventory: "在庫",
  estimates: "見積",
  shifts: "シフト",
  master: "マスタ設定",
  "hospital-settings": "病院設定",
};

// All available resources for permission configuration
const AllResources = Object.keys(RESOURCE_LABELS);

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface PermissionRule {
  resource: string;
  canView: boolean;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

interface PermissionRuleTableProps {
  group: PermissionGroup | null;
  rules: PermissionRule[];
  onRuleChange: (resource: string, field: keyof Omit<PermissionRule, "resource">, value: boolean) => void;
}

// ─────────────────────────────────────────────────
// PermissionRuleTable
// ─────────────────────────────────────────────────

export const PermissionRuleTable = memo(function PermissionRuleTable({
  group,
  rules,
  onRuleChange,
}: PermissionRuleTableProps) {
  // Show permission table regardless of group state (both edit and new modes)
  // For new groups (group === null), display empty rules with all permissions unchecked

  const ruleMap = new Map(rules.map((r) => [r.resource, r]));

  return (
    <div className="mt-6">
      <h3 className={`text-sm font-semibold ${C.text} mb-3`}>権限設定</h3>
      <div className="overflow-x-auto border rounded-lg">
        <table className="w-full">
          <thead>
            <tr className={`border-b ${C.borderLight}`}>
              <th className={`px-4 py-2 text-left text-xs font-semibold ${C.textSecondary}`}>
                リソース
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.textSecondary}`}>
                表示
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.textSecondary}`}>
                作成
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.textSecondary}`}>
                編集
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.textSecondary}`}>
                削除
              </th>
            </tr>
          </thead>
          <tbody>
            {AllResources.map((resource) => {
              const rule = ruleMap.get(resource) || {
                resource,
                canView: false,
                canCreate: false,
                canEdit: false,
                canDelete: false,
              };

              return (
                <DataTableRow key={resource}>
                  <TableCell className={`text-sm ${C.text}`}>
                    {RESOURCE_LABELS[resource] || resource}
                  </TableCell>
                  <TableCell className="text-center">
                    <Checkbox
                      checked={rule.canView}
                      onCheckedChange={(checked) =>
                        onRuleChange(resource, "canView", checked === true)
                      }
                    />
                  </TableCell>
                  <TableCell className="text-center">
                    <Checkbox
                      checked={rule.canCreate}
                      onCheckedChange={(checked) =>
                        onRuleChange(resource, "canCreate", checked === true)
                      }
                    />
                  </TableCell>
                  <TableCell className="text-center">
                    <Checkbox
                      checked={rule.canEdit}
                      onCheckedChange={(checked) =>
                        onRuleChange(resource, "canEdit", checked === true)
                      }
                    />
                  </TableCell>
                  <TableCell className="text-center">
                    <Checkbox
                      checked={rule.canDelete}
                      onCheckedChange={(checked) =>
                        onRuleChange(resource, "canDelete", checked === true)
                      }
                    />
                  </TableCell>
                </DataTableRow>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
});
