import type { Hospitalization, InventoryItem, MedicalRecord, ReservationStatus, SortOrder } from "@/types";
import { RESERVATION_STATUS_LABELS } from "@/types";
import { C } from "@/lib/design-tokens";
import type React from "react";

// ── Notion カラーパレット定数（インライン参照用） ──────────────────────────
const N = {
  // Blue（アクセント）
  blueBg:    "bg-[#D3E5EF]",
  blueText:  "text-[#183B56]",
  blueBorder:"border-[#B8D4E3]",
  // Green（ステータス緑）
  greenBg:    "bg-[#DDEDEA]",
  greenText:  "text-[#0F7B6C]",
  greenBorder:"border-[#DDEDEA]",
  // Purple（ステータス紫）
  purpleBg:    "bg-[#EEE0F7]",
  purpleText:  "text-[#6940A5]",
  purpleBorder:"border-[#6940A5]/20",
  // Orange（割引/会計）
  orangeBg:    "bg-[#FAEBDD]",
  orangeText:  "text-[#D9730D]",
  orangeBorder:"border-[#D9730D]/20",
  // Yellow（注意/保留）
  yellowBg:    "bg-[#FDECC8]",
  yellowText:  "text-[#C29243]",
  yellowBorder:"border-[#F2DBA7]",
  // Red（エラー/危険）
  redBg:    "bg-[#FFDEDE]",
  redText:  "text-[#EB5757]",
  redBorder:"border-[#EB5757]/20",
  // Gray（ニュートラル）
  grayBg:    "bg-[#EAE9E5]",
  grayText:  "text-[#37352F]/60",
  grayBorder:"border-[rgba(55,53,47,0.09)]",
} as const;

const badge = (bg: string, text: string, border: string) =>
  `${bg} ${text} ${border}`;

// ────────────────────────────────────────────────────────────────

export const getMedicalRecordStatusColor = (status: MedicalRecord["status"]) => {
  switch (status) {
    case "作成中":
      return badge(N.blueBg, N.blueText, N.blueBorder);
    case "確定済":
      return badge(N.grayBg, N.grayText, N.grayBorder);
    default:
      return "";
  }
};

export const getHospitalizationStatusColor = (status: Hospitalization["status"]) => {
  switch (status) {
    case "入院中":
      return badge(N.blueBg, N.blueText, N.blueBorder);
    case "退院済":
      return badge(N.grayBg, N.grayText, N.grayBorder);
    case "予約":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    case "一時帰宅":
      return badge(N.yellowBg, N.yellowText, N.yellowBorder);
    default:
      return "";
  }
};

export const getHospitalizationTypeColor = (type: Hospitalization["hospitalizationType"]) => {
  return type === "入院"
    ? badge(N.purpleBg, N.purpleText, N.purpleBorder)
    : badge(N.blueBg, N.blueText, N.blueBorder);
};

interface DashboardColumnColorSet {
  bg: string;
  dot: string;
  text: string;
  text60: string;
  hoverBgPage: string;
  hoverText: string;
}

const DASHBOARD_COLUMN_COLOR_MAP: Record<string, DashboardColumnColorSet> = {
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

export const getDashboardColumnColor = (title: string): DashboardColumnColorSet => {
  return DASHBOARD_COLUMN_COLOR_MAP[title] ?? {
    bg: C.bgPage,
    dot: C.bgStatusGrayMedium,
    text: C.text,
    text60: C.text60,
    hoverBgPage: C.hoverBgPage,
    hoverText: C.hoverText,
  };
};

interface ServiceTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

export const getReservationTypeColor = (type: string, dynamicColorMap?: Map<string, ServiceTypeColor>) => {
  if (dynamicColorMap) {
    const mapped = dynamicColorMap.get(type);
    if (mapped) return mapped.style;
  }

  // Fallback to Notion palette (className based)
  switch (type) {
    case "treatment":
    case "診療":
      return badge(N.blueBg, N.blueText, N.blueBorder);
    case "checkup":
    case "検診":
    case "検査":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    case "surgery":
    case "手術":
      return badge(N.redBg, N.redText, N.redBorder);
    case "trimming":
    case "トリミング":
      return badge(N.orangeBg, N.orangeText, N.orangeBorder);
    case "vaccine":
    case "ワクチン":
      return badge(N.purpleBg, N.purpleText, N.purpleBorder);
    case "入院":
    case "ホテル":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    default:
      return badge(N.grayBg, N.grayText, N.grayBorder);
  }
};

export const getReservationTypeName = (type: string) => {
  switch (type) {
    case "treatment": return "診療";
    case "checkup": return "検診";
    case "surgery": return "手術";
    case "trimming": return "トリミング";
    case "vaccine": return "ワクチン";
    default: return type || "その他";
  }
};

export const getExaminationStatusColor = (status: string) => {
  switch (status) {
    case "依頼中":
      return badge(N.yellowBg, N.yellowText, N.yellowBorder);
    case "検査中":
      return badge(N.blueBg, N.blueText, N.blueBorder);
    case "完了":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    default:
      return "";
  }
};

export const getAccountingStatusColor = (status: string) => {
  switch (status) {
    case "会計待ち":
      return badge(N.orangeBg, N.orangeText, N.orangeBorder);
    case "会計済":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    case "キャンセル":
      return badge(N.grayBg, N.grayText, N.grayBorder);
    default:
      return "";
  }
};

export const getTrimmingStatusColor = (status: string) => {
  switch (status) {
    case "完了":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    case "予約":
      return badge(N.blueBg, N.blueText, N.blueBorder);
    case "進行中":
      return badge(N.orangeBg, N.orangeText, N.orangeBorder);
    default:
      return badge(N.grayBg, N.grayText, N.grayBorder);
  }
};

export const getPetStatusColor = (status: string) => {
  return status === "生存"
    ? `${N.greenBg} ${N.greenText} ${N.greenBorder} hover:bg-[#DDEDEA]`
    : `${N.grayBg} text-[#9B9A97] ${N.grayBorder} hover:bg-[#EAE9E5]`;
};

export const getMasterStatusColor = (status: string) => {
  return status === "active"
    ? badge(N.greenBg, N.greenText, N.greenBorder)
    : badge(N.grayBg, N.grayText, N.grayBorder);
};

export const getInventoryStatusColor = (status: InventoryItem["status"]) => {
  switch (status) {
    case "sufficient":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    case "low":
      return badge(N.yellowBg, N.yellowText, N.yellowBorder);
    case "out_of_stock":
      return badge(N.redBg, N.redText, N.redBorder);
    default:
      return "";
  }
};

export const getInventoryStatusLabel = (status: InventoryItem["status"]) => {
  switch (status) {
    case "sufficient":
      return "十分";
    case "low":
      return "残少";
    case "out_of_stock":
      return "在庫切れ";
    default:
      return "";
  }
};

export const getCalendarViewLabel = (view: string): string => {
  switch (view) {
    case "month":
      return "月表示";
    case "week":
      return "週表示";
    default:
      return view;
  }
};

export const getReservationStatusLabel = (status: ReservationStatus): string => {
  return RESERVATION_STATUS_LABELS[status] ?? status;
};

export function getSortOrderLabel(order: SortOrder): string {
  return order === "desc" ? "新→古" : "古→新";
}

export const getEstimateStatusColor = (status: string) => {
  switch (status) {
    case "draft":
      return badge(N.grayBg, N.grayText, N.grayBorder);
    case "sent":
      return badge(N.blueBg, N.blueText, N.blueBorder);
    case "accepted":
      return badge(N.greenBg, N.greenText, N.greenBorder);
    case "rejected":
      return badge(N.redBg, N.redText, N.redBorder);
    case "expired":
      return badge(N.orangeBg, N.orangeText, N.orangeBorder);
    default:
      return badge(N.grayBg, N.grayText, N.grayBorder);
  }
};
