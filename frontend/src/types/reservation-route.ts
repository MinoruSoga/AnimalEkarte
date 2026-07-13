// FE6-17: features/reservations/constants/reservation-route.ts から正本を移動。
// CODING_RULES §3.6 に従い型定義は src/types/ に置き、feature 側は re-export のみとする。
export const RESERVATION_ROUTE_VALUES = [
  "line",
  "phone",
  "reception",
  "exam_room",
  "record_shortcut",
] as const;

export type ReservationRoute = (typeof RESERVATION_ROUTE_VALUES)[number];

export const RESERVATION_ROUTE_LABELS: Record<ReservationRoute, string> = {
  line: "LINE",
  phone: "電話",
  reception: "受付",
  exam_room: "診察室",
  record_shortcut: "記録入力",
};
