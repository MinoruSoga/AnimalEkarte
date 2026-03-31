import { CircleDot } from "lucide-react";
import { C } from "@/lib/design-tokens";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";

export const MASTER_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

/** 全マスタ共通: 有効/無効 ステータスフィルタプロパティ */
export const MASTER_STATUS_FILTER: FilterProperty = {
  key: "status",
  label: "ステータス",
  type: "select",
  icon: CircleDot,
  options: [
    { value: "active", label: "有効" },
    { value: "inactive", label: "無効" },
  ],
};
