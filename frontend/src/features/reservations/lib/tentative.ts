import { formatJSTWallDate, todayJSTISO } from "@/lib/jst-date";
import type { Reservation } from "../types";

/**
 * EMR-308: tentative（次回予定・未予約）予約の判定・期限計算ユーティリティ。
 * tentative 行は BE が status="tentative" で返すが generated 型には未反映のため、
 * transformReservation が isTentative / intendedDate に正規化済み。
 * 意図日は JST 暦日で比較する（start は意図日 00:00 JST の sentinel）。
 */

/** カレンダー/詳細で表示する明示ラベル（色だけに頼らない識別子） */
export const TENTATIVE_APPOINTMENT_LABEL = "次回予定・未予約";
/** 意図日が過ぎた tentative の明示ラベル（期限超過パネル/詳細バナー） */
export const EXPIRED_TENTATIVE_LABEL = "期限超過・未対応";

export function isTentativeReservation(reservation: Pick<Reservation, "isTentative">): boolean {
  return reservation.isTentative === true;
}

/**
 * tentative の意図日（YYYY-MM-DD）。intendedDate が無い行（レスポンス欠落等）は
 * sentinel start の JST 暦日でフォールバックする。
 */
export function tentativeIntendedDate(
  reservation: Pick<Reservation, "intendedDate" | "start">,
): string {
  return reservation.intendedDate ?? formatJSTWallDate(reservation.start);
}

/**
 * 意図日が JST 暦日で「今日より前」の tentative は期限超過。
 * 当日の tentative は期限超過ではない（当日こそ確定/紐付けの対象）。
 */
export function isExpiredTentativeReservation(
  reservation: Pick<Reservation, "isTentative" | "intendedDate" | "start">,
  today: string = todayJSTISO(),
): boolean {
  if (!isTentativeReservation(reservation)) return false;
  return tentativeIntendedDate(reservation) < today;
}
