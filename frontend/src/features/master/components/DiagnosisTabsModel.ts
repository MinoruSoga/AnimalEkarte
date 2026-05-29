import type { ReactNode } from "react";

import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";

interface DataTableColumn {
  header: ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
}

export const DIAGNOSIS_TYPE_COLUMNS: DataTableColumn[] = [
  { header: "", className: "w-[32px]" },
  { header: "カテゴリ名" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" },
  { header: "操作", className: "w-[80px]", align: "right" },
];

export const DIAGNOSIS_NAME_COLUMNS: DataTableColumn[] = [
  { header: "", className: "w-[32px]" },
  { header: "所属カテゴリ", className: "w-[160px]" },
  { header: "診断病名" },
  { header: "ステータス", className: "w-[100px]", align: "center" },
  { header: "操作", className: "w-[80px]", align: "right" },
];

export function filterDiagnosisTypesBySearch(
  items: DiagnosisType[],
  searchTerm: string,
): DiagnosisType[] {
  if (!searchTerm) return items;
  const lower = searchTerm.toLowerCase();
  return items.filter((category) => category.name.toLowerCase().includes(lower));
}

export function filterDiagnosisNamesBySearch(
  items: DiagnosisName[],
  searchTerm: string,
): DiagnosisName[] {
  if (!searchTerm) return items;
  const lower = searchTerm.toLowerCase();
  return items.filter((name) => name.name.toLowerCase().includes(lower));
}

export function buildDiagnosisTypeNameMap(categories: DiagnosisType[] | undefined) {
  return new Map<string, string>(
    (categories ?? []).map((category) => [category.id, category.name]),
  );
}
