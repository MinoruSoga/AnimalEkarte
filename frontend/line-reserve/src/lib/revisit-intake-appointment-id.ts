// EMR-474 / D4: 予約 ID は現行 DTO では JSON number。Number.MAX_SAFE_INTEGER を超えた値は
// parse 時点で丸められており、String() しても元の ID に戻らない（別予約を指す恐れがある）。
// BE が ID を文字列で返すまで（EMR-519）、安全整数でない ID から再診問診の request を出さない。

export const REVISIT_INTAKE_UNAVAILABLE_MESSAGE = "この予約の問診は現在開けません。";

/** 再診問診 API の appointment_id (decimal string) に使える、正の安全整数の予約 ID か。 */
export function isRevisitIntakeAppointmentIdUsable(id: number): boolean {
  return Number.isSafeInteger(id) && id > 0;
}

/** 使えない ID は null を返し、呼び出し側は request を発行しない。 */
export function toRevisitIntakeAppointmentId(id: number): string | null {
  return isRevisitIntakeAppointmentIdUsable(id) ? String(id) : null;
}
