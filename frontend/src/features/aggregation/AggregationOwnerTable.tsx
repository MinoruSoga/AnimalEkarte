import { Link } from "react-router";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, STYLE, BADGE } from "@/lib/design-tokens";
import type { AggregationOwner, LastVisitBucket } from "./api/get-aggregations";

export type AggregationTab = "revenue" | "visit" | "last_visit";

interface AggregationOwnerTableProps {
  owners: AggregationOwner[];
  selectedOwnerIds: Set<string>;
  onSelectAll: (checked: boolean) => void;
  onSelectOwner: (ownerId: string, checked: boolean) => void;
  isLoading: boolean;
  activeTab: AggregationTab;
  isError?: boolean;
  errorMessage?: string;
}




const LAST_VISIT_BUCKET_LABEL: Record<LastVisitBucket, string> = {
  within_3m: "3ヶ月以内",
  over_3m: "3ヶ月以上",
  over_6m: "6ヶ月以上",
  over_1y: "1年以上",
  no_visit: "来院なし",
};

const LAST_VISIT_BUCKET_CLASS: Record<LastVisitBucket, string> = {
  within_3m: BADGE.green,
  over_3m: BADGE.yellow,
  over_6m: BADGE.orange,
  over_1y: BADGE.red,
  no_visit: BADGE.gray,
};

function LastVisitBucketBadge({ bucket }: { bucket: LastVisitBucket | null }) {
  if (bucket === null) {
    return <span className={`text-sm ${C.text40}`}>—</span>;
  }
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${LAST_VISIT_BUCKET_CLASS[bucket]}`}
    >
      {LAST_VISIT_BUCKET_LABEL[bucket]}
    </span>
  );
}

function formatFee(fee: number | undefined): string {
  if (fee === undefined) return "—";
  return `¥${fee.toLocaleString("ja-JP")}`;
}

function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return "—";
  return dateStr.slice(0, 10);
}

function formatDaysSince(days: number | null | undefined): string {
  if (days === null || days === undefined) return "—";
  return `${days}日`;
}

interface ColumnDef {
  key: string;
  label: string;
  width?: string;
  render: (owner: AggregationOwner) => React.ReactNode;
  textAlign?: "left" | "right";
}

const COMMON_COLUMNS: ColumnDef[] = [
  {
    key: "checkbox",
    label: "",
    width: "w-12",
    render: () => null, // handled separately
  },
  {
    key: "owner_name",
    label: "飼い主名",
    render: (owner) => (
      <Link
        to={`/owners/${owner.owner_id}`}
        className={`${C.accent} hover:underline`}
        onClick={(e) => e.stopPropagation()}
      >
        {owner.owner_name}
      </Link>
    ),
  },
];

const TAB_SPECIFIC_COLUMNS: Record<AggregationTab, ColumnDef[]> = {
  revenue: [
    ...COMMON_COLUMNS,
    {
      key: "annual_amount",
      label: "期間診療費",
      width: "w-32",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{formatFee(owner.annual_amount)}</span>
      ),
    },
    {
      key: "billing_count",
      label: "会計件数",
      width: "w-24",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{owner.billing_count ?? "—"}</span>
      ),
    },
    {
      key: "period_visit_count",
      label: "来院回数",
      width: "w-24",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{owner.period_visit_count ?? "—"}</span>
      ),
    },
    {
      key: "last_visit_date",
      label: "最終来院日",
      width: "w-28",
      render: (owner) => (
        <span className="font-mono">{formatDate(owner.last_visit_date)}</span>
      ),
    },
    {
      key: "first_visit_date",
      label: "初診日",
      width: "w-28",
      render: (owner) => (
        <span className="font-mono">{formatDate(owner.first_visit_date)}</span>
      ),
    },
  ],
  visit: [
    ...COMMON_COLUMNS,
    {
      key: "period_visit_count",
      label: "来院回数(期間)",
      width: "w-24",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{owner.period_visit_count ?? "—"}</span>
      ),
    },
    {
      key: "total_visit_count",
      label: "累計来院",
      width: "w-24",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{owner.total_visit_count}</span>
      ),
    },
    {
      key: "annual_visit_count",
      label: "年間来院",
      width: "w-24",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{owner.annual_visit_count}</span>
      ),
    },
    {
      key: "last_visit_date",
      label: "最終来院日",
      width: "w-28",
      render: (owner) => (
        <span className="font-mono">{formatDate(owner.last_visit_date)}</span>
      ),
    },
    {
      key: "first_visit_date",
      label: "初診日",
      width: "w-28",
      render: (owner) => (
        <span className="font-mono">{formatDate(owner.first_visit_date)}</span>
      ),
    },
  ],
  last_visit: [
    ...COMMON_COLUMNS,
    {
      key: "last_visit_date",
      label: "最終来院日",
      width: "w-28",
      render: (owner) => (
        <span className="font-mono">{formatDate(owner.last_visit_date)}</span>
      ),
    },
    {
      key: "days_since_last_visit",
      label: "経過日数",
      width: "w-24",
      textAlign: "right",
      render: (owner) => (
        <span className="font-mono">{formatDaysSince(owner.days_since_last_visit)}</span>
      ),
    },
    {
      key: "last_visit_bucket",
      label: "分類",
      width: "w-32",
      render: (owner) => (
        <LastVisitBucketBadge bucket={owner.last_visit_bucket ?? null} />
      ),
    },
    {
      key: "first_visit_date",
      label: "初診日",
      width: "w-28",
      render: (owner) => (
        <span className="font-mono">{formatDate(owner.first_visit_date)}</span>
      ),
    },
  ],
};

export function AggregationOwnerTable({
  owners,
  selectedOwnerIds,
  onSelectAll,
  onSelectOwner,
  isLoading,
  activeTab,
  isError,
  errorMessage,
}: AggregationOwnerTableProps) {
  const columns = TAB_SPECIFIC_COLUMNS[activeTab];
  const colSpan = columns.length + 1; // +1 for checkbox
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
            {columns
              .filter((col) => col.key !== "checkbox")
              .map((col) => (
                <TableHead
                  key={col.key}
                  className={`${STYLE.tableHeaderCell} px-4 ${col.width ?? ""} ${
                    col.textAlign === "right" ? "text-right" : ""
                  }`}
                >
                  {col.label}
                </TableHead>
              ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {isError ? (
            <TableRow>
              <TableCell colSpan={colSpan} className={`${STYLE.tableEmpty} ${C.danger}`}>
                {errorMessage || "エラーが発生しました"}
              </TableCell>
            </TableRow>
          ) : isLoading ? (
            <TableRow>
              <TableCell colSpan={colSpan} className={STYLE.tableEmpty}>
                読み込み中...
              </TableCell>
            </TableRow>
          ) : owners.length === 0 ? (
            <TableRow>
              <TableCell colSpan={colSpan} className={STYLE.tableEmpty}>
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
                {columns
                  .filter((col) => col.key !== "checkbox")
                  .map((col) => (
                    <TableCell
                      key={col.key}
                      className={`${STYLE.tableCell} px-4 ${col.width ?? ""} ${
                        col.textAlign === "right" ? "text-right" : ""
                      }`}
                    >
                      {col.render(owner)}
                    </TableCell>
                  ))}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
