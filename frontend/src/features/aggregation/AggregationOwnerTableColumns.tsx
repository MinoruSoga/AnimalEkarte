import type { ReactNode } from "react";
import { Link } from "react-router";

import { BADGE, C } from "@/lib/design-tokens";

import type { AggregationOwner, LastVisitBucket } from "./api/get-aggregations";
import type { AggregationTab } from "./AggregationDashboardPage";

export interface AggregationOwnerColumn {
  key: string;
  label: string;
  width?: string;
  render: (owner: AggregationOwner) => ReactNode;
  textAlign?: "left" | "right";
}

const LAST_VISIT_BUCKET_LABEL: Record<LastVisitBucket, string> = {
  within_3m: "3ヶ月未満",
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

function renderLastVisitBucketBadge(bucket: LastVisitBucket | null) {
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

const OWNER_NAME_COLUMN: AggregationOwnerColumn = {
  key: "owner_name",
  label: "飼主名",
  render: (owner) => (
    <Link
      to={`/owners/${owner.owner_id}`}
      className={`${C.text} hover:underline decoration-dotted underline-offset-2`}
      onClick={(e) => e.stopPropagation()}
    >
      {owner.owner_name}
    </Link>
  ),
};

const TAB_SPECIFIC_COLUMNS: Record<AggregationTab, AggregationOwnerColumn[]> = {
  revenue: [
    OWNER_NAME_COLUMN,
    {
      key: "annual_amount",
      label: "年間診療費",
      width: "w-32",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{formatFee(owner.annual_amount)}</span>,
    },
    {
      key: "billing_count",
      label: "会計件数",
      width: "w-24",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.billing_count ?? "—"}</span>,
    },
    {
      key: "period_visit_count",
      label: "期間内来院回数",
      width: "w-28",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.period_visit_count ?? "—"}</span>,
    },
    {
      key: "last_visit_date",
      label: "最終来院日",
      width: "w-28",
      render: (owner) => <span className="font-mono">{formatDate(owner.last_visit_date)}</span>,
    },
    {
      key: "first_visit_date",
      label: "初診日",
      width: "w-28",
      render: (owner) => <span className="font-mono">{formatDate(owner.first_visit_date)}</span>,
    },
  ],
  visit: [
    OWNER_NAME_COLUMN,
    {
      key: "period_visit_count",
      label: "期間内来院回数",
      width: "w-28",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.period_visit_count ?? "—"}</span>,
    },
    {
      key: "total_visit_count",
      label: "累計来院回数",
      width: "w-28",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.total_visit_count}</span>,
    },
    {
      key: "annual_visit_count",
      label: "年間来院回数",
      width: "w-28",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.annual_visit_count}</span>,
    },
    {
      key: "last_visit_date",
      label: "最終来院日",
      width: "w-28",
      render: (owner) => <span className="font-mono">{formatDate(owner.last_visit_date)}</span>,
    },
    {
      key: "first_visit_date",
      label: "初診日",
      width: "w-28",
      render: (owner) => <span className="font-mono">{formatDate(owner.first_visit_date)}</span>,
    },
  ],
  last_visit: [
    OWNER_NAME_COLUMN,
    {
      key: "last_visit_date",
      label: "最終来院日",
      width: "w-28",
      render: (owner) => <span className="font-mono">{formatDate(owner.last_visit_date)}</span>,
    },
    {
      key: "days_since_last_visit",
      label: "経過日数",
      width: "w-24",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{formatDaysSince(owner.days_since_last_visit)}</span>,
    },
    {
      key: "last_visit_bucket",
      label: "分類",
      width: "w-32",
      render: (owner) => renderLastVisitBucketBadge(owner.last_visit_bucket ?? null),
    },
    {
      key: "total_visit_count",
      label: "累計来院回数",
      width: "w-28",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.total_visit_count}</span>,
    },
    {
      key: "annual_visit_count",
      label: "年間来院回数",
      width: "w-28",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{owner.annual_visit_count}</span>,
    },
    {
      key: "total_amount",
      label: "累計診療費",
      width: "w-32",
      textAlign: "right",
      render: (owner) => <span className="font-mono">{formatFee(owner.total_amount ?? owner.total_fee)}</span>,
    },
  ],
};

export function getAggregationOwnerColumns(activeTab: AggregationTab) {
  return TAB_SPECIFIC_COLUMNS[activeTab];
}
