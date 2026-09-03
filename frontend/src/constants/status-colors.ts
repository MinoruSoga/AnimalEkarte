/**
 * 予約ステータス・診察区分・ダッシュボードのカラー定数集約ファイル。
 * ReservationDetailModal / AppointmentCard / ReceptionDetailModal のインライン定義を統合。
 */
import { C } from "@/lib/design-tokens";
import { RESERVATION_STATUS_LABELS } from "@/types";
import type { ReservationStatus } from "@/types";

// ──────────────────────────────────────────────
// 予約ステータスカラーマップ
// ──────────────────────────────────────────────

export const RESERVATION_STATUS_COLORS = {
  confirmed: {
    label: RESERVATION_STATUS_LABELS.confirmed,
    dot: C.bgStatusEmeraldDot,
    bg: C.bgStatusEmerald,
    text: C.textStatusEmerald,
  },
  pending: {
    label: RESERVATION_STATUS_LABELS.pending,
    dot: C.bgStatusSkyDot,
    bg: C.bgStatusSky,
    text: C.textStatusSky,
  },
  checked_in: {
    label: RESERVATION_STATUS_LABELS.checked_in,
    dot: C.bgStatusBlueDot,
    bg: C.bgStatusBlueLight,
    text: C.textStatusBlue,
  },
  in_consultation: {
    label: RESERVATION_STATUS_LABELS.in_consultation,
    dot: C.bgStatusPurpleDot,
    bg: C.bgStatusPurple,
    text: C.textStatusPurple,
  },
  accounting: {
    label: RESERVATION_STATUS_LABELS.accounting,
    dot: C.bgStatusAmberDot,
    bg: C.bgStatusAmber,
    text: C.textStatusAmber,
  },
  completed: {
    label: RESERVATION_STATUS_LABELS.completed,
    dot: C.bgStatusGrayMedium,
    bg: C.bgStatusGray,
    text: C.textStatusGray,
  },
  cancelled: {
    label: RESERVATION_STATUS_LABELS.cancelled,
    dot: C.bgStatusRedDot,
    bg: C.bgRedLight,
    text: C.textNotionRed,
  },
  no_show: {
    label: RESERVATION_STATUS_LABELS.no_show,
    dot: C.bgStatusRedDot,
    bg: C.bgRedLight,
    text: C.textNotionRed,
  },
} as const;

/**
 * 予約ステータスのカラー情報を返す。
 * 未知のステータスは "pending" にフォールバック。
 */
export function getReservationStatusColor(status: string) {
  return (
    RESERVATION_STATUS_COLORS[status as ReservationStatus] ?? RESERVATION_STATUS_COLORS.pending
  );
}

// ──────────────────────────────────────────────
// visitType（初診/再診）カラーマップ
// ──────────────────────────────────────────────

/** アクセントカラー（ReservationDetailModal のヘッダー帯・バッジ） */
const VISIT_TYPE_COLORS = {
  初診: {
    border: C.borderRedBadge,
    bg: C.bgRedLight,
    text: C.textNotionRed,
    dot: C.bgStatusRedDot,
    // AppointmentCard バッジ用カラー
    badgeBg: C.bgAccentLight60,
    badgeText: C.textAccentDark90,
    badgeBorder: C.borderAccentBadge50,
  },
  再診: {
    border: C.borderAccentBadge,
    bg: C.bgAccentLight,
    text: C.textAccentDark,
    dot: C.bgStatusBlueDot,
    // AppointmentCard バッジ用カラー
    badgeBg: C.bgPage60,
    badgeText: C.text90,
    badgeBorder: C.borderLight50,
  },
} as const;

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

/** ReceptionDetailModal のステータスバッジカラー（日本語キー） */
export const RECEPTION_STATUS_COLORS: Record<string, string> = {
  受付予約: `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentLight}`,
  受付済: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen}`,
  診療中: `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderStatusPurple}`,
  会計待ち: `${C.bgWarning50} ${C.textWarningIcon} ${C.borderWarning20}`,
  会計済: `${C.bgActive} ${C.text} ${C.borderLight}`,
};

/** RECEPTION_STATUS_COLORS 未定義ステータスのフォールバッククラス */
export const RECEPTION_STATUS_COLOR_FALLBACK = `${C.bgStatusGray} ${C.textStatusGray} ${C.borderStatusGray}`;
