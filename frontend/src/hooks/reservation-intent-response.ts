import { transformReservation } from "@/lib/transforms/reservation";

/**
 * EMR-308: tentative（次回予定・未予約）関連の BE レスポンスフィールド。
 * generated 型は型再生成までこれらを持たないため、transform 入力側で widen する。
 * BE 契約: backend/internal/reservation/reservation_response.go
 * （intended_date: "YYYY-MM-DD"、had_intended_date、source_medical_record_id）。
 *
 * src/hooks 配置の理由: src/lib/transforms は本ユニットの所有パス外。
 * 共有 transform の入力型は Parameters<typeof transformReservation>[0] で導出し、
 * @/types/generated/models の新規 import（TASK-444-S1 禁止）を避ける。
 */
export interface ReservationIntentResponseFields {
  intended_date?: string | null;
  had_intended_date?: boolean;
  source_medical_record_id?: number | null;
}

/**
 * transformReservation の入力型（BE 予約レスポンス + EMR-308 tentative 拡張）。
 * axios のレスポンス型付けは呼び出し側でこの型を使う。
 * ReservationIntentResponseFields は全プロパティ optional のため、素の BE 型も代入可能。
 */
export type ReservationApiResponse = Parameters<typeof transformReservation>[0] &
  ReservationIntentResponseFields;

/**
 * EMR-308: 共有 transformReservation に tentative フィールドを付与した変換。
 * - isTentative: generated ReservationStatus は "tentative" を含まないため String 比較で判別
 * - intendedDate: 次回予定の意図日（YYYY-MM-DD、JST 暦日）。tentative 由来の行のみ
 * - hadIntendedDate: 昇格元の tentative があったか（BE はキャンセル後も intended_date を保持）
 * - sourceMedicalRecordId: 作成元カルテ id（next-visit-intent 経路のみ）
 */
export const transformReservationWithIntent = (reservation: ReservationApiResponse) => ({
  ...transformReservation(reservation),
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
