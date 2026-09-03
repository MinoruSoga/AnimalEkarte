import type { ReactNode } from "react";

import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";
import { normalizeKana } from "@/lib/normalize-kana";

interface DataTableColumn {
  header: ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
}

export const DIAGNOSIS_TYPE_COLUMNS: DataTableColumn[] = [
  { header: "", className: "w-11 px-0" },
  { header: "カテゴリ名" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" },
  { header: "操作", className: "w-[80px]", align: "right" },
];

export const DIAGNOSIS_NAME_COLUMNS: DataTableColumn[] = [
  { header: "", className: "w-11 px-0" },
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
  const lower = normalizeKana(searchTerm).toLowerCase();
  return items.filter((category) => normalizeKana(category.name).toLowerCase().includes(lower));
}

export function filterDiagnosisNamesBySearch(
  items: DiagnosisName[],
  searchTerm: string,
): DiagnosisName[] {
  if (!searchTerm) return items;
  const lower = normalizeKana(searchTerm).toLowerCase();
  return items.filter((name) => normalizeKana(name.name).toLowerCase().includes(lower));
}

export function buildDiagnosisTypeNameMap(categories: DiagnosisType[] | undefined) {
  return new Map<string, string>(
    (categories ?? []).map((category) => [category.id, category.name]),
  );
}
