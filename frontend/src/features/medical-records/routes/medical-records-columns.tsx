// React/Framework
import { type ReactNode, useCallback, useMemo } from "react";

// Internal
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { LIST_TABLE_COL } from "@/components/shared/DataTable/list-table-col";

// Types
import type { MedicalRecordSortKey } from "../api/get-medical-records";

// B-1 follow-up: 列ソート server 化。BE が許可する4キー（medical_record_repository.go の
// medicalRecordSortColumns）のみ SortableHeader を表示する。種/主訴/関連/担当医は非対応（follow-up）。
const SORTABLE_COLUMN_LABELS: Record<MedicalRecordSortKey, string> = {
  date: "診療日",
  owner_name: "飼主名",
  pet_name: "ペット名",
  status: "ステータス",
};

interface MedicalRecordColumn {
  header: ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
}

interface UseMedicalRecordsColumnsArgs {
  showClinicColumn: boolean;
  directionForSort: (key: MedicalRecordSortKey) => "ascending" | "descending" | "none";
  onSortToggle: (key: MedicalRecordSortKey) => void;
}

export function useMedicalRecordsColumns({
  showClinicColumn,
  directionForSort,
  onSortToggle,
}: UseMedicalRecordsColumnsArgs): MedicalRecordColumn[] {
  // rendering-hoist-jsx 対象外: directionFor/onToggle が sortKey/sortOrder に依存するため
  // COLUMNS 自体を useMemo でメモ化し、ソート状態変化時のみ再生成する。
  const sortableHeader = useCallback((key: MedicalRecordSortKey) => (
    <SortableHeader
      label={SORTABLE_COLUMN_LABELS[key]}
      direction={directionForSort(key)}
      onToggle={() => onSortToggle(key)}
    />
  ), [directionForSort, onSortToggle]);

  return useMemo<MedicalRecordColumn[]>(() => [
    { header: sortableHeader("date"), className: "w-[120px]" },
    { header: sortableHeader("owner_name") },
    { header: sortableHeader("pet_name") },
    { header: "種", className: "w-[80px] hidden lg:table-cell" },
    // 主訴/担当医 are secondary at narrow widths; status + 操作 stay visible (BUG-458).
    { header: "主訴", className: "hidden md:table-cell" },
    { header: "関連", className: "w-[100px] hidden lg:table-cell" },
    { header: "担当医", className: "w-[100px] hidden md:table-cell" },
    { header: sortableHeader("status"), className: LIST_TABLE_COL.status },
    ...(showClinicColumn ? [{ header: "医院", className: "w-[120px] hidden lg:table-cell" }] : []),
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [showClinicColumn, sortableHeader]);
}
