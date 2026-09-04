// TrimmingListTable のフィルタ定義 (非コンポーネント)。コンポーネントファイルから分離して
// react-refresh/only-export-components 違反を解消する。
import { Calendar, CircleDot, PawPrint, User } from "lucide-react";
import {
  CONDITIONS_NO_EMPTY,
  CONDITIONS_WITH_EMPTY,
} from "@/components/shared/PropertyFilter/types";
import type { FilterProperty } from "@/components/shared/PropertyFilter/types";

const TRIMMING_STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "予約", label: "予約" },
      { value: "進行中", label: "進行中" },
      { value: "完了", label: "完了" },
    ],
  },
];

export function buildTrimmingDynamicFilterProperties(
  speciesOptions: { value: string; label: string }[],
  staffOptions: { value: string; label: string }[],
): FilterProperty[] {
  return [
    ...TRIMMING_STATIC_FILTER_PROPERTIES,
    {
      key: "species",
      label: "種",
      type: "select" as const,
      icon: PawPrint,
      conditions: CONDITIONS_NO_EMPTY,
      options: speciesOptions,
    },
    {
      key: "staff",
      label: "担当",
      type: "select" as const,
      icon: User,
      conditions: CONDITIONS_WITH_EMPTY,
      options: staffOptions,
    },
  ];
}
