const NO_ADD_BUTTON_COLUMNS = new Set(["診療中", "会計待ち", "会計済"]);

/** `?date=` が受理する形式。暦日の実在性は別途フィールド往復一致で検査する。 */
const RECEPTION_DATE_PARAM_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;

/**
 * `?date=YYYY-MM-DD` を受付ボードの表示日へ解決する。
 * 未指定・空・形式不正・非存在の暦日（例: 2/30, 13月）は fallback（当日 JST）へ丸める。
 * `new Date(str)` は UTC 深夜罠（lib/jst-date.ts FE4-8 契約）のため使わず、
 * 数値コンストラクタ + フィールド往復一致で実在暦日だけを受理する。
 */
export function resolveReceptionDateParam(dateParam: string | null, fallback: string): string {
  if (!dateParam) return fallback;
  const matched = RECEPTION_DATE_PARAM_PATTERN.exec(dateParam);
  if (!matched) return fallback;
  const year = Number(matched[1]);
  const month = Number(matched[2]);
  const day = Number(matched[3]);
  const probe = new Date(year, month - 1, day);
  if (probe.getFullYear() !== year || probe.getMonth() !== month - 1 || probe.getDate() !== day) {
    return fallback;
  }
  return dateParam;
}

/** 「＋」ボタンが発行する予約作成クエリ種別。受付予約=通常予約、それ以外=当日受付 walk-in。 */
export type ReceptionColumnAddKind = "newReservation" | "reception";

export function receptionColumnAddKind(columnTitle: string): ReceptionColumnAddKind {
  return columnTitle === "受付予約" ? "newReservation" : "reception";
}

/**
 * カラムの「＋」ボタンを出すかどうか。EMR-243: 非本日表示では
 * checked_in（当日受付）を作る reception 系の導線だけを抑制し、
 * 受付予約列の通常予約作成は選択日のまま許可する。
 */
export function canAddColumnEntryOnDate(columnTitle: string, isToday: boolean): boolean {
  if (NO_ADD_BUTTON_COLUMNS.has(columnTitle)) return false;
  if (!isToday && receptionColumnAddKind(columnTitle) === "reception") return false;
  return true;
}
