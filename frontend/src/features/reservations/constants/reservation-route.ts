export const RESERVATION_ROUTE_VALUES = [
  "line",
  "phone",
  "reception",
  "exam_room",
] as const;

export type ReservationRoute = (typeof RESERVATION_ROUTE_VALUES)[number];

export const RESERVATION_ROUTE_LABELS: Record<ReservationRoute, string> = {
  line: "LINE",
  phone: "電話",
  reception: "受付",
  exam_room: "診察室",
};
