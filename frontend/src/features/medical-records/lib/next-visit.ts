import {
  formatJSTWallDate,
  formatJSTWallTime,
  jstWallDateToISOString,
  todayJSTISO,
} from "@/lib/jst-date";
import { addYearsIso } from "@/components/shared/ReservationFormModal/next-visit-intent-date";
import type { Reservation } from "@/lib/transforms/reservation";

// EMR-308: 意図日ルール（過去日 NG・上限2年）は共有 NextVisitResetDialog も使うため
// components/shared/ReservationFormModal/next-visit-intent-date.ts が正本。
// ここは feature 内既存 import を壊さない re-export。
export { validateNextVisitIntentDate } from "@/components/shared/ReservationFormModal/next-visit-intent-date";

/**
 * EMR-308: 'tentative' は generated ReservationStatus union に永遠に入らない
 * （Go 側で ReservationStatusTentative を const enum block 外の var にすることで
 * tygo が union に含めない設計）。したがってこのローカル拡張 union は暫定措置
 * ではなく恒久型として維持する。
 */
type NextVisitReservationStatus = Reservation["status"] | "tentative";

/**
 * 次回予定表示の導出に必要な最小形状。
 * transformReservation の結果（Reservation[]）はキャストなしで代入できる。
 * tentative 行の意図日は intendedDate を優先する（reservations feature の
 * tentativeIntendedDate と同一規約）。欠落時のみ sentinel start の
 * JST 暦日にフォールバックする。
 */
export interface NextVisitSourceReservation {
  /** 確定モーダル（POST confirm-tentative）の対象行識別子 */
  id: string;
  status: NextVisitReservationStatus;
  start: Date;
  /** BE 応答の intended_date（YYYY-MM-DD）。載っていれば sentinel 導出より優先。 */
  intendedDate?: string;
}

export interface NextVisitState {
  /** "appointment": 実予約あり / "tentative": 次回予定のみ / "unset": 未設定 */
  kind: "appointment" | "tentative" | "unset";
  /** ボタン面に出す文字列。appointment は "YYYY-MM-DD HH:mm"、tentative は "YYYY-MM-DD"、unset は ""。 */
  label: string;
  /** Popover 日付入力の値。tentative が存在すればその意図日、無ければ ""。 */
  intentDate: string;
  /** アクティブな tentative 行の id。確定導線の対象。なければ null。 */
  tentativeId: string | null;
}

const REAL_APPOINTMENT_STATUSES: readonly NextVisitReservationStatus[] = ["pending", "confirmed"];

/** The next actual start that changes the display priority; tentative dates do not expire. */
export function nextVisitRefreshAt(
  reservations: readonly NextVisitSourceReservation[],
  now: Date,
): number | null {
  let next: Date | null = null;
  for (const reservation of reservations) {
    if (
      REAL_APPOINTMENT_STATUSES.includes(reservation.status) &&
      reservation.start.getTime() > now.getTime() &&
      (next === null || reservation.start.getTime() < next.getTime())
    ) {
      next = reservation.start;
    }
  }
  return next === null ? null : Date.parse(jstWallDateToISOString(next));
}

/**
 * ヘッダー「次回予定」の表示優先度（EMR-308）:
 *   1. 実予約（pending/confirmed かつ start_time が未来）— 最も近いものを日時表示
 *   2. tentative（アクティブな次回予定）— 意図日を日付表示
 *   3. 未設定
 * now は JST wall-clock Date（toJSTWallDate(new Date())）を渡す。
 * reservation.start も JST wall-clock Date なので getTime() 同士の比較で JST の前後関係になる。
 */
export function deriveNextVisitState(
  reservations: readonly NextVisitSourceReservation[],
  now: Date,
): NextVisitState {
  let nextAppointment: NextVisitSourceReservation | null = null;
  let tentative: NextVisitSourceReservation | null = null;
  for (const reservation of reservations) {
    if (REAL_APPOINTMENT_STATUSES.includes(reservation.status)) {
      if (
        reservation.start.getTime() > now.getTime() &&
        (nextAppointment === null || reservation.start.getTime() < nextAppointment.start.getTime())
      ) {
        nextAppointment = reservation;
      }
    } else if (reservation.status === "tentative" && tentative === null) {
      tentative = reservation;
    }
  }

  const intentDate = tentative
    ? (tentative.intendedDate ?? formatJSTWallDate(tentative.start))
    : "";
  const tentativeId = tentative?.id ?? null;

  if (nextAppointment) {
    return {
      kind: "appointment",
      label: `${formatJSTWallDate(nextAppointment.start)} ${formatJSTWallTime(nextAppointment.start)}`,
      intentDate,
      tentativeId,
    };
  }
  if (tentative) {
    return { kind: "tentative", label: intentDate, intentDate, tentativeId };
  }
  return { kind: "unset", label: "", intentDate: "", tentativeId: null };
}

/**
 * ヘッダーがフェッチする予約レンジの上限日（YYYY-MM-DD）。
 * 意図日の上限（今日から2年）と揃える。startDate は呼び出し側で todayJSTISO() を使う。
 */
export function nextVisitFetchEndDate(todayJst = todayJSTISO()): string {
  return addYearsIso(todayJst, 2);
}
