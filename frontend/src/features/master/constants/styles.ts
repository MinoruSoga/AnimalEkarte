import { CircleDot } from "lucide-react";
import { C } from "@/lib/design-tokens";
import type { FilterProperty } from "@/components/shared/PropertyFilter/types";

export const MASTER_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`;

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

/**
 * 全マスタ共通: テーブル列幅（DataTable columns の className に使用）。
 * 「ステータス」「操作」列は複数ファイルで同一幅(100px/80px)が重複していたため一元化。
 * ステータス列は90pxだと「ステータス」ラベルが折り返すため100pxを標準値とする。
 * DESIGN.mdのspacingスケールは32pxまでのため、列幅(コンテンツ駆動の100-200px域)はスケール対象外の構造値として扱う。
 */
export const MASTER_TABLE_COL = {
  w80: "w-[80px]",
  w100: "w-[100px]",
  w120: "w-[120px]",
  w130: "w-[130px]",
  w140: "w-[140px]",
  w150: "w-[150px]",
  w180: "w-[180px]",
  w200: "w-[200px]",
} as const;
