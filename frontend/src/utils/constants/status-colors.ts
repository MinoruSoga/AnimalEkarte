/**
 * 予約ステータス・診察区分・ダッシュボードのカラー定数集約ファイル。
 * ReservationDetailModal / AppointmentCard / DashboardDetailModal のインライン定義を統合。
 */
import { C } from "@/lib/design-tokens";

// ──────────────────────────────────────────────
// 予約ステータスカラーマップ
// ──────────────────────────────────────────────

export const RESERVATION_STATUS_COLORS = {
  confirmed:       { label: "予約確定",    dot: "bg-emerald-500", bg: "bg-emerald-50",  text: "text-emerald-700" },
  pending:         { label: "仮予約",      dot: "bg-sky-500",     bg: "bg-sky-50",      text: "text-sky-700" },
  checked_in:      { label: "受付済",      dot: "bg-blue-500",    bg: "bg-blue-50",     text: "text-blue-700" },
  in_consultation: { label: "診療中",      dot: "bg-violet-500",  bg: "bg-violet-50",   text: "text-violet-700" },
  accounting:      { label: "会計待ち",    dot: "bg-amber-500",   bg: "bg-amber-50",    text: "text-amber-700" },
  completed:       { label: "完了",        dot: "bg-gray-400",    bg: "bg-gray-50",     text: "text-gray-600" },
  cancelled:       { label: "キャンセル",  dot: "bg-red-500",     bg: "bg-red-50",      text: "text-red-700" },
} as const;

export type ReservationStatus = keyof typeof RESERVATION_STATUS_COLORS;

/**
 * 予約ステータスのカラー情報を返す。
 * 未知のステータスは "pending" にフォールバック。
 */
export function getReservationStatusColor(status: string) {
  return (
    RESERVATION_STATUS_COLORS[status as ReservationStatus] ??
    RESERVATION_STATUS_COLORS.pending
  );
}

// ──────────────────────────────────────────────
// visitType（初診/再診）カラーマップ
// ──────────────────────────────────────────────

/** アクセントカラー（ReservationDetailModal のヘッダー帯・バッジ） */
export const VISIT_TYPE_COLORS = {
  初診: {
    border: "border-red-200",
    bg: "bg-red-50",
    text: "text-red-700",
    dot: "bg-red-500",
    // AppointmentCard バッジ用カラー
    badgeBg: "bg-[#D3E5EF]/60",
    badgeText: "text-[#183B56]/90",
    badgeBorder: "border-[#B8D4E3]/50",
  },
  再診: {
    border: "border-blue-200",
    bg: "bg-blue-50",
    text: "text-blue-700",
    dot: "bg-blue-500",
    // AppointmentCard バッジ用カラー
    badgeBg: "bg-[#F7F6F3]/60",
    badgeText: "text-[#37352F]/90",
    badgeBorder: "border-[rgba(55,53,47,0.09)]/50",
  },
} as const;

export type VisitType = keyof typeof VISIT_TYPE_COLORS;

/**
 * visitType（初診/再診 または first/return）のカラー情報を返す。
 * "first" は "初診" として扱い、それ以外は "再診" にフォールバック。
 */
export function getVisitTypeColor(visitType: string) {
  if (visitType === "初診" || visitType === "first") {
    return VISIT_TYPE_COLORS["初診"];
  }
  return VISIT_TYPE_COLORS["再診"];
}

// ──────────────────────────────────────────────
// ダッシュボード表示用 日本語ステータスカラーマップ
// ──────────────────────────────────────────────

/** DashboardDetailModal のステータスバッジカラー（日本語キー） */
export const DASHBOARD_STATUS_COLORS: Record<string, string> = {
  "受付予約": `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentLight}`,
  "受付済":   `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen}`,
  "診療中":   `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderStatusPurple}`,
  "会計待ち": `${C.bgWarning50} ${C.textWarningIcon} ${C.borderWarning20}`,
  "会計済":   `${C.bgActive} ${C.text} ${C.borderLight}`,
};

/** DASHBOARD_STATUS_COLORS 未定義ステータスのフォールバッククラス */
export const DASHBOARD_STATUS_COLOR_FALLBACK = "bg-gray-100 text-gray-600 border-gray-200";
