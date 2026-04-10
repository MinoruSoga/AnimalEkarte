import type { Hospitalization, InventoryItem, MedicalRecord, ReservationStatus, SortOrder } from "@/types";
import { RESERVATION_STATUS_LABELS } from "@/types";
import { C, BADGE } from "@/lib/design-tokens";
import type React from "react";

// ────────────────────────────────────────────────────────────────
// ⚠️ このファイルは旧来の色定義を BADGE トークンにブリッジするためのものです。
// 新規コンポーネントでは可能な限り BADGE.xxx を直接参照してください。
// ────────────────────────────────────────────────────────────────

export const getMedicalRecordStatusColor = (status: MedicalRecord["status"]) => {
  switch (status) {
    case "作成中": return BADGE.blue;
    case "確定済": return BADGE.gray;
    default: return "";
  }
};

export const getHospitalizationStatusColor = (status: Hospitalization["status"]) => {
  switch (status) {
    case "入院中": return BADGE.blue;
    case "退院済": return BADGE.gray;
    case "予約":   return BADGE.green;
    case "一時帰宅": return BADGE.yellow;
    default: return "";
  }
};

export const getHospitalizationTypeColor = (type: Hospitalization["hospitalizationType"]) => {
  return type === "入院" ? BADGE.purple : BADGE.blue;
};

interface ReceptionColumnColorSet {
  bg: string;
  dot: string;
  text: string;
  text60: string;
  hoverBgPage: string;
  hoverText: string;
}

const RECEPTION_COLUMN_COLOR_MAP: Record<string, ReceptionColumnColorSet> = {
  "受付予約": {
    bg: C.bgPage,
    dot: C.bgStatusGrayMedium,
    text: C.text,
    text60: C.text60,
    hoverBgPage: C.hoverBgPage,
    hoverText: C.hoverText,
  },
  "受付済": {
    bg: C.bgAccentLight50,
    dot: C.bgAccent,
    text: C.textAccentDark,
    text60: C.textAccentDark60,
    hoverBgPage: C.hoverBgAccentBadge40,
    hoverText: C.hoverTextAccentDark,
  },
  "診療中": {
    bg: C.bgStatusPurple60,
    dot: C.bgStatusPurpleDot,
    text: C.textStatusPurple,
    text60: C.textStatusPurple60,
    hoverBgPage: C.hoverBgPurpleLight40,
    hoverText: C.hoverTextStatusPurple,
  },
  "会計待ち": {
    bg: C.bgDiscountLight70,
    dot: C.bgDiscount,
    text: C.textDiscount,
    text60: C.textDiscount70,
    hoverBgPage: C.hoverBgOrangeBadge40,
    hoverText: C.hoverTextDiscount,
  },
  "会計済": {
    bg: C.bgStatusGreen60,
    dot: C.bgStatusGreenDot,
    text: C.textStatusGreen,
    text60: C.textStatusGreen60,
    hoverBgPage: C.hoverBgGreenBadge40,
    hoverText: C.hoverTextStatusGreen,
  },
};

export const getReceptionColumnColor = (title: string): ReceptionColumnColorSet => {
  return RECEPTION_COLUMN_COLOR_MAP[title] ?? {
    bg: C.bgPage,
    dot: C.bgStatusGrayMedium,
    text: C.text,
    text60: C.text60,
    hoverBgPage: C.hoverBgPage,
    hoverText: C.hoverText,
  };
};

interface ReservationCategoryColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

export const getReservationTypeColor = (type: string, dynamicColorMap?: Map<string, ReservationCategoryColor>) => {
  if (dynamicColorMap) {
    const mapped = dynamicColorMap.get(type);
    if (mapped) return mapped.style;
  }

  switch (type) {
    case "treatment":
    case "診療":   return BADGE.blue;
    case "checkup":
    case "検診":
    case "検査":   return BADGE.green;
    case "surgery":
    case "手術":   return BADGE.red;
    case "trimming":
    case "トリミング": return BADGE.orange;
    case "vaccine":
    case "ワクチン": return BADGE.purple;
    case "入院":
    case "ホテル": return BADGE.green;
    default:      return BADGE.gray;
  }
};

export const getReservationTypeName = (type: string) => {
  switch (type) {
    case "treatment": return "診療";
    case "checkup":   return "検診";
    case "surgery":   return "手術";
    case "trimming":  return "トリミング";
    case "vaccine":   return "ワクチン";
    default:          return type || "その他";
  }
};

export const getExaminationStatusColor = (status: string) => {
  switch (status) {
    case "依頼中": return BADGE.yellow;
    case "検査中": return BADGE.blue;
    case "完了":   return BADGE.green;
    default:      return "";
  }
};

export const getAccountingStatusColor = (status: string) => {
  switch (status) {
    case "会計待ち": return BADGE.orange;
    case "会計済":   return BADGE.green;
    case "キャンセル": return BADGE.gray;
    default:      return "";
  }
};

export const getTrimmingStatusColor = (status: string) => {
  switch (status) {
    case "完了":   return BADGE.green;
    case "予約":   return BADGE.blue;
    case "進行中": return BADGE.orange;
    default:      return BADGE.gray;
  }
};

export const getPetStatusColor = (status: string) => {
  return status === "生存" ? BADGE.greenHover : BADGE.grayHover;
};

export const getMasterStatusColor = (status: string) => {
  return status === "active" ? BADGE.green : BADGE.gray;
};

export const getInventoryStatusColor = (status: InventoryItem["status"]) => {
  switch (status) {
    case "sufficient":   return BADGE.green;
    case "low":          return BADGE.yellow;
    case "out_of_stock": return BADGE.red;
    default:            return "";
  }
};

export const getInventoryStatusLabel = (status: InventoryItem["status"]) => {
  switch (status) {
    case "sufficient":   return "十分";
    case "low":          return "残少";
    case "out_of_stock": return "在庫切れ";
    default:            return "";
  }
};

export const getCalendarViewLabel = (view: string): string => {
  return view === "month" ? "月表示" : view === "week" ? "週表示" : view;
};

export const getReservationStatusLabel = (status: ReservationStatus): string => {
  return RESERVATION_STATUS_LABELS[status] ?? status;
};

export function getSortOrderLabel(order: SortOrder): string {
  return order === "desc" ? "新→古" : "古→新";
}

export const getEstimateStatusColor = (status: string) => {
  switch (status) {
    case "draft":    return BADGE.gray;
    case "sent":     return BADGE.blue;
    case "accepted": return BADGE.green;
    case "rejected": return BADGE.red;
    case "expired":  return BADGE.orange;
    default:        return BADGE.gray;
  }
};
