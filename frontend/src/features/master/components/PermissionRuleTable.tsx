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
  dashboard: "当日の受付",
  owners: "飼主・ペット",
  reservations: "予約管理",
  "medical-records": "カルテ",
  hospitalization: "入院・ホテル",
  trimming: "トリミング",
  examinations: "検査管理",
  accounting: "会計管理",
  vaccinations: "予防接種",
  checkups: "定期健診",
  inventory: "在庫管理",
  estimates: "見積書",
  shifts: "シフト管理",
  "hospital-settings": "医院",
  "master-animal-species": "動物種類",
  "master-medical": "カルテ関連",
  "master-service-type": "診療サービス",
  "master-hospitalization": "入院・ケージ",
  "master-trimming": "トリミングマスタ",
  "master-permission": "権限グループ",
  "master-staff": "スタッフ管理",
  "master-insurance": "保険マスタ",
  "master-merchandise": "物販・フード",
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
              <th className={`px-4 py-2 text-left text-xs font-semibold ${C.text50}`}>
                リソース
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.text50}`}>
                表示
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.text50}`}>
                作成
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.text50}`}>
                編集
              </th>
              <th className={`px-4 py-2 text-center text-xs font-semibold ${C.text50}`}>
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
