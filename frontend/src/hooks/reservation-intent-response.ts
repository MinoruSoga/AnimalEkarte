import { transformReservation } from "@/lib/transforms/reservation";

/**
 * transformReservation の入力型（BE 予約レスポンス）。axios のレスポンス型付けは
 * 呼び出し側でこの型を使う。EMR-308 tentative 列（intended_date: "YYYY-MM-DD"、
 * had_intended_date、source_medical_record_id）は移行 DTO（@/types/reservation-legacy-wire、
 * backend/internal/reservation/reservation_response.go reservationResponse）が持つ（EMR-527）。
 */
export type ReservationApiResponse = Parameters<typeof transformReservation>[0];

/**
 * EMR-308: 共有 transformReservation に tentative フィールドを付与した変換。
 * - isTentative: generated ReservationStatus は "tentative" を含まないため String 比較で判別
 * - intendedDate: 次回予定の意図日（YYYY-MM-DD、JST 暦日）。tentative 由来の行のみ
 * - hadIntendedDate: 昇格元の tentative があったか（BE はキャンセル後も intended_date を保持）
 * - sourceMedicalRecordId: 作成元カルテ id（next-visit-intent 経路のみ）
 */
export const transformReservationWithIntent = (reservation: ReservationApiResponse) => ({
  ...transformReservation(reservation),
  /** Preserve the server revision verbatim; Date would truncate sub-millisecond precision. */
  ...(reservation.updated_at ? { updatedAt: reservation.updated_at } : {}),
  /** EMR-308: tentative（次回予定・未予約）行かどうか */
  isTentative: String(reservation.status) === "tentative",
  /** 次回予定の意図日（YYYY-MM-DD、JST 暦日）。tentative 由来の行のみ */
  intendedDate: reservation.intended_date ?? undefined,
  /** 昇格元の tentative があったか */
  hadIntendedDate: reservation.had_intended_date === true,
  /** 作成元カルテ id（next-visit-intent 経路のみ） */
  sourceMedicalRecordId:
    reservation.source_medical_record_id != null
      ? String(reservation.source_medical_record_id)
      : undefined,
});

/**
 * EMR-308 拡張フィールド（isTentative / intendedDate / hadIntendedDate /
 * sourceMedicalRecordId）を含む予約型。共有 Reservation の上位型。
 */
export type Reservation = ReturnType<typeof transformReservationWithIntent>;
