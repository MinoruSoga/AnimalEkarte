import { Calendar, CircleDot, PawPrint, User } from "lucide-react";
import type { FilterCondition, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { C, STYLE } from "@/lib/design-tokens";

interface MedicalRecordsFilterMaster {
  staffs: { id: string; name: string; isActive: boolean }[] | undefined;
  activeSpecies: { id: number; name: string }[];
  isSpeciesError: boolean;
  isSpeciesLoading: boolean;
}

export function buildMedicalRecordsFilterProperties(
  input: MedicalRecordsFilterMaster,
): FilterProperty[] {
  const doctorOptions = (input.staffs ?? [])
    .filter((s) => s.isActive)
    .map((s) => ({ value: s.id, label: s.name }));
  const speciesOptions =
    input.isSpeciesError || input.isSpeciesLoading
      ? []
      : input.activeSpecies.map((s) => ({ value: String(s.id), label: s.name }));
  return [
    ...STATIC_FILTER_PROPERTIES,
    { key: "doctor", label: "担当医", type: "select" as const, icon: User, conditions: SERVER_EQUALITY_ONLY, options: doctorOptions },
    { key: "species", label: "種", type: "select" as const, icon: PawPrint, conditions: SERVER_EQUALITY_ONLY, options: speciesOptions },
  ];
}

export const PAGE_SIZE = 20;

export const SERVER_EQUALITY_ONLY: FilterCondition[] = ["is"];

export const STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "診療日",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    conditions: SERVER_EQUALITY_ONLY,
    options: [
      { value: "作成中", label: "作成中" },
      { value: "確定済", label: "確定済" },
    ],
  },
];

export const CLINIC_TOGGLE_RESET_PARAMS = ["page"] as const;

export const MEDICAL_RECORDS_HEADER_ROW = `border-b ${C.borderLight} ${C.bgPage} h-11`;
export const MEDICAL_RECORDS_HEADER_CELL = `${STYLE.sectionLabel} h-11`;
