import type { Hospitalization, InventoryItem, MedicalRecord, ReservationStatus, SortOrder } from "@/types";
import { RESERVATION_STATUS_LABELS } from "@/types";
import { C } from "@/lib/design-tokens";

export const getMedicalRecordStatusColor = (status: MedicalRecord["status"]) => {
  switch (status) {
    case "作成中":
      return "bg-blue-500/10 text-blue-700 border-blue-200";
    case "確定済":
      return "bg-gray-500/10 text-gray-700 border-gray-200";
    default:
      return "";
  }
};

export const getHospitalizationStatusColor = (status: Hospitalization["status"]) => {
  switch (status) {
    case "入院中":
      return "bg-blue-500/10 text-blue-700 border-blue-200";
    case "退院済":
      return "bg-gray-500/10 text-gray-700 border-gray-200";
    case "予約":
      return "bg-green-500/10 text-green-700 border-green-200";
    default:
      return "";
  }
};

export const getHospitalizationTypeColor = (type: Hospitalization["hospitalizationType"]) => {
  return type === "入院"
    ? "bg-red-500/10 text-red-700 border-red-200"
    : "bg-purple-500/10 text-purple-700 border-purple-200";
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

export const getReservationTypeColor = (type: string) => {
    switch (type) {
      case "treatment":
      case "診療":
          return "bg-blue-100 text-blue-700 border-blue-200";
      case "checkup":
      case "検診":
      case "検査":
          return "bg-green-100 text-green-700 border-green-200";
      case "surgery":
      case "手術":
          return "bg-red-100 text-red-700 border-red-200";
      case "trimming":
      case "トリミング":
          return "bg-orange-100 text-orange-700 border-orange-200";
      case "vaccine":
      case "ワクチン":
          return "bg-purple-100 text-purple-700 border-purple-200";
      case "入院":
      case "ホテル":
          return "bg-cyan-100 text-cyan-700 border-cyan-200";
      default: return "bg-gray-100 text-gray-700 border-gray-200";
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
        return "bg-yellow-500/10 text-yellow-700 border-yellow-200";
      case "検査中":
        return "bg-blue-500/10 text-blue-700 border-blue-200";
      case "完了":
        return "bg-green-500/10 text-green-700 border-green-200";
      default:
        return "";
    }
  };

export const getAccountingStatusColor = (status: string) => {
    switch (status) {
      case "未収":
        return "bg-red-500/10 text-red-700 border-red-200";
      case "回収済":
        return "bg-green-500/10 text-green-700 border-green-200";
      case "キャンセル":
        return "bg-gray-500/10 text-gray-700 border-gray-200";
      default:
        return "";
    }
  };

export const getTrimmingStatusColor = (status: string) => {
  switch (status) {
    case "完了":
      return "bg-[#E8F5E9] text-[#2E7D32] border-[#E8F5E9]"; // Green badge
    case "予約":
      return "bg-[#E3F2FD] text-[#1565C0] border-[#E3F2FD]"; // Blue badge
    case "進行中":
      return "bg-[#FFF3E0] text-[#E65100] border-[#FFF3E0]"; // Orange badge
    default:
      return "bg-gray-100 text-gray-800 border-gray-200";
  }
};

export const getPetStatusColor = (status: string) => {
  return status === "生存"
    ? "bg-[#DDEDEA] text-[#0F7B6C] border-[#DDEDEA] hover:bg-[#DDEDEA]"
    : "bg-[#EBECED] text-[#9B9A97] border-[#EBECED] hover:bg-[#EBECED]";
};

export const getMasterStatusColor = (status: string) => {
  return status === "active"
    ? "bg-green-50 text-green-700 border-green-200"
    : "bg-gray-100 text-gray-500 border-gray-200";
};

export const getInventoryStatusColor = (status: InventoryItem["status"]) => {
  switch (status) {
    case "sufficient":
      return "bg-green-500/10 text-green-700 border-green-200";
    case "low":
      return "bg-amber-500/10 text-amber-700 border-amber-200";
    case "out_of_stock":
      return "bg-red-500/10 text-red-700 border-red-200";
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
