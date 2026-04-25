import { Link } from "react-router";
import { Check, Minus } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import type { LtvOwner } from "../api/get-ltv-owners";

interface LtvOwnerTableProps {
  owners: LtvOwner[];
  selectedOwnerIds: Set<string>;
  onSelectAll: (checked: boolean) => void;
  onSelectOwner: (ownerId: string, checked: boolean) => void;
  isLoading: boolean;
}

type CpmStage = "encounter" | "growing" | "core" | "noah" | "spot" | "dormant";

const CPM_STAGE_LABEL: Record<CpmStage, string> = {
  encounter: "エンカウンター",
  growing: "グロウィング",
  core: "コア",
  noah: "ノア",
  spot: "スポット",
  dormant: "休眠",
};

// design-tokens の C クラスで対応困難なため Tailwind 直書き
const CPM_STAGE_CLASS: Record<CpmStage, string> = {
  encounter: "bg-cyan-100 text-cyan-800 border-cyan-200",
  growing: "bg-blue-100 text-blue-800 border-blue-200",
  core: "bg-green-100 text-green-800 border-green-200",
  noah: "bg-purple-100 text-purple-800 border-purple-200",
  spot: "bg-orange-100 text-orange-800 border-orange-200",
  dormant: "bg-yellow-100 text-yellow-800 border-yellow-200",
};

function CpmStageBadge({ stage }: { stage: CpmStage | null }) {
  if (stage === null) {
    return <span className={`text-sm ${C.text40}`}>—</span>;
  }
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${CPM_STAGE_CLASS[stage]}`}
    >
      {CPM_STAGE_LABEL[stage]}
    </span>
  );
}

function formatFee(fee: number): string {
  return `¥${fee.toLocaleString("ja-JP")}`;
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return "—";
  return dateStr.slice(0, 10);
}

export function LtvOwnerTable({
  owners,
  selectedOwnerIds,
  onSelectAll,
  onSelectOwner,
  isLoading,
}: LtvOwnerTableProps) {
  const allSelected = owners.length > 0 && owners.every((o) => selectedOwnerIds.has(o.owner_id));
  const someSelected = owners.some((o) => selectedOwnerIds.has(o.owner_id)) && !allSelected;

  return (
    <div className={STYLE.tableContainer}>
      <Table>
        <TableHeader>
          <TableRow className={STYLE.tableHeaderRow}>
            <TableHead className={`${STYLE.tableHeaderCell} w-12 px-4`}>
              <Checkbox
                checked={someSelected ? "indeterminate" : allSelected}
                onCheckedChange={(checked) => onSelectAll(!!checked)}
                aria-label="全選択"
                className={`${C.dataCheckedBgAccent} ${C.dataCheckedBorderAccent}`}
              />
            </TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4`}>飼い主名</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-24`}>LINE</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-32`}>CPMステージ</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-32 text-right`}>累計診療費</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-24 text-right`}>年間来院</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-24 text-right`}>累計来院</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-28`}>最終来院日</TableHead>
            <TableHead className={`${STYLE.tableHeaderCell} px-4 w-28`}>初診日</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow>
              <TableCell colSpan={9} className={STYLE.tableEmpty}>
                読み込み中...
              </TableCell>
            </TableRow>
          ) : owners.length === 0 ? (
            <TableRow>
              <TableCell colSpan={9} className={STYLE.tableEmpty}>
                データが見つかりません
              </TableCell>
            </TableRow>
          ) : (
            owners.map((owner) => (
              <TableRow
                key={owner.owner_id}
                className={`${STYLE.tableRow} ${selectedOwnerIds.has(owner.owner_id) ? C.bgBrand10 : ""}`}
              >
                <TableCell className="px-4 py-2">
                  <Checkbox
                    checked={selectedOwnerIds.has(owner.owner_id)}
                    onCheckedChange={(checked) =>
                      onSelectOwner(owner.owner_id, !!checked)
                    }
                    aria-label={`${owner.owner_name}を選択`}
                    className={`${C.dataCheckedBgAccent} ${C.dataCheckedBorderAccent}`}
                  />
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4`}>
                  <Link
                    to={`/owners/${owner.owner_id}`}
                    className={`${C.accent} hover:underline`}
                    onClick={(e) => e.stopPropagation()}
                  >
                    {owner.owner_name}
                  </Link>
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4`}>
                  {owner.has_line ? (
                    <Check className="size-4 text-[#06C755]" />
                  ) : (
                    <Minus className={`size-4 ${C.text30}`} />
                  )}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4`}>
                  <CpmStageBadge stage={owner.cpm_stage} />
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4 text-right font-mono`}>
                  {formatFee(owner.total_fee)}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4 text-right font-mono`}>
                  {owner.annual_visit_count}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4 text-right font-mono`}>
                  {owner.total_visit_count}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4 font-mono`}>
                  {formatDate(owner.last_visit_date)}
                </TableCell>
                <TableCell className={`${STYLE.tableCell} px-4 font-mono`}>
                  {formatDate(owner.first_visit_date)}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
