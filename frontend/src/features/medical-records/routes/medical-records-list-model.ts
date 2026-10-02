import {
  Activity,
  Calendar,
  CircleDot,
  FileText,
  Package,
  PawPrint,
  Pill,
  Stethoscope,
  User,
} from "lucide-react";
import type { FilterCondition, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { C, STYLE } from "@/lib/design-tokens";

interface MedicalRecordsFilterMaster {
  staffs: { id: string; name: string; isActive: boolean }[] | undefined;
  activeSpecies: { id: number; name: string }[];
  isSpeciesError: boolean;
  isSpeciesLoading: boolean;
  medicines: { id: string; name: string; isActive: boolean }[] | undefined;
  procedures: { id: string; name: string; isActive: boolean }[] | undefined;
  consultations: { id: string; name: string; isActive: boolean }[] | undefined;
  inventories: { id: string; name: string }[] | undefined;
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
    {
      key: "doctor",
      label: "担当医",
      type: "select" as const,
      icon: User,
      conditions: SERVER_EQUALITY_ONLY,
      options: doctorOptions,
    },
    {
      key: "species",
      label: "種",
      type: "select" as const,
      icon: PawPrint,
      conditions: SERVER_EQUALITY_ONLY,
      options: speciesOptions,
    },
    {
      key: "medicine",
      label: "薬剤",
      type: "select" as const,
      icon: Pill,
      conditions: SERVER_EQUALITY_ONLY,
      options: (input.medicines ?? [])
        .filter((m) => m.isActive)
        .map((m) => ({ value: m.id, label: m.name })),
    },
    {
      key: "procedure",
      label: "処置",
      type: "select" as const,
      icon: Activity,
      conditions: SERVER_EQUALITY_ONLY,
      options: (input.procedures ?? [])
        .filter((p) => p.isActive)
        .map((p) => ({ value: p.id, label: p.name })),
    },
    {
      key: "consultation",
      label: "診察",
      type: "select" as const,
      icon: Stethoscope,
      conditions: SERVER_EQUALITY_ONLY,
      options: (input.consultations ?? [])
        .filter((c) => c.isActive)
        .map((c) => ({ value: c.id, label: c.name })),
    },
    {
      key: "inventory",
      label: "在庫品",
      type: "select" as const,
      icon: Package,
      conditions: SERVER_EQUALITY_ONLY,
      options: (input.inventories ?? []).map((i) => ({ value: i.id, label: i.name })),
    },
  ];
}

export const PAGE_SIZE = 20;

const SERVER_EQUALITY_ONLY: FilterCondition[] = ["is"];

const STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "診療日",
    type: "date-range",
    icon: Calendar,
  },
  // EMR-245: 表示列（飼主名・ペット名・主訴）の部分一致フィルタ。
  // type:"text" は contains 固定（BE は owner_name / pet_name / chief_complaint
  // の ILIKE 部分一致のみ受け付ける）。横断検索の search とは別パラメータで送る。
  {
    key: "owner_name",
    label: "飼主名",
    type: "text",
    icon: User,
  },
  {
    key: "pet_name",
    label: "ペット名",
    type: "text",
    icon: PawPrint,
  },
  {
    key: "chief_complaint",
    label: "主訴",
    type: "text",
    icon: FileText,
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
