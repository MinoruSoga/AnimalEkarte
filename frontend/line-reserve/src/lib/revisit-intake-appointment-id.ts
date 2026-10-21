// EMR-519: 予約 ID は BE から正の int64 範囲の十進文字列で返る。Number へ変換すると
// 2^53 超で precision を失うため、正規形（先頭ゼロ・符号・小数点なし）と
// 桁数 + 辞書順の上限比較だけで検査し、文字列のまま扱う。

export const REVISIT_INTAKE_UNAVAILABLE_MESSAGE = "この予約の問診は現在開けません。";

const APPOINTMENT_ID_INT64_MAX = "9223372036854775807";
const CANONICAL_DECIMAL_ID = /^[1-9][0-9]*$/;

/** 再診問診 API の appointment_id (decimal string) に使える、正規形の正の int64 十進文字列か。 */
export function isRevisitIntakeAppointmentIdUsable(id: string): boolean {
  if (!CANONICAL_DECIMAL_ID.test(id)) return false;
  return (
    id.length < APPOINTMENT_ID_INT64_MAX.length ||
    (id.length === APPOINTMENT_ID_INT64_MAX.length && id <= APPOINTMENT_ID_INT64_MAX)
  );
}

/** 使えない ID は null を返し、呼び出し側は request を発行しない。 */
export function toRevisitIntakeAppointmentId(id: string): string | null {
  return isRevisitIntakeAppointmentIdUsable(id) ? id : null;
}
