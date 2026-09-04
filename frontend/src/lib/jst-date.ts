const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

function toInstant(value: Date | string): Date {
  return typeof value === "string" ? new Date(value) : value;
}

function toJSTParts(value: Date | string): Date {
  const instant = toInstant(value);
  return new Date(instant.getTime() + JST_OFFSET_MS);
}

export function formatJSTDate(value: Date | string): string {
  const jstDate = toJSTParts(value);
  return `${jstDate.getUTCFullYear()}-${padDatePart(jstDate.getUTCMonth() + 1)}-${padDatePart(jstDate.getUTCDate())}`;
}

export function formatJSTTime(value: Date | string): string {
  const jstDate = toJSTParts(value);
  return `${padDatePart(jstDate.getUTCHours())}:${padDatePart(jstDate.getUTCMinutes())}`;
}

export function todayJSTISO(): string {
  return formatJSTDate(new Date());
}

export function isPastJSTDate(date: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) return false;
  return date < todayJSTISO();
}

export function toJSTWallDate(value: Date | string): Date {
  const jstDate = toJSTParts(value);
  return new Date(
    jstDate.getUTCFullYear(),
    jstDate.getUTCMonth(),
    jstDate.getUTCDate(),
    jstDate.getUTCHours(),
    jstDate.getUTCMinutes(),
    jstDate.getUTCSeconds(),
    jstDate.getUTCMilliseconds(),
  );
}

export function buildJSTWallDateTime(date: string, time: string): Date {
  const [year, month, day] = date.split("-").map(Number);
  const [hour, minute] = time.split(":").map(Number);
  return new Date(year, month - 1, day, hour, minute, 0, 0);
}

export function jstWallDateToISOString(date: Date): string {
  const instant = new Date(
    Date.UTC(
      date.getFullYear(),
      date.getMonth(),
      date.getDate(),
      date.getHours() - 9,
      date.getMinutes(),
      date.getSeconds(),
      date.getMilliseconds(),
    ),
  );
  return instant.toISOString();
}

export function jstDateStartISOString(date: string): string {
  return `${date}T00:00:00+09:00`;
}

export function jstNowISOString(): string {
  return formatJSTOffsetDateTime(new Date());
}

function formatJSTOffsetDateTime(value: Date | string): string {
  const jstDate = toJSTParts(value);
  return `${jstDate.getUTCFullYear()}-${padDatePart(jstDate.getUTCMonth() + 1)}-${padDatePart(jstDate.getUTCDate())}T${padDatePart(jstDate.getUTCHours())}:${padDatePart(jstDate.getUTCMinutes())}:${padDatePart(jstDate.getUTCSeconds())}.${String(jstDate.getUTCMilliseconds()).padStart(3, "0")}+09:00`;
}

export function formatJSTDateTimeLocal(value: Date | string): string {
  const date = toJSTWallDate(value);
  return `${formatJSTWallDate(date)}T${formatJSTWallTime(date)}`;
}

export function jstDateTimeLocalToISOString(value: string): string {
  if (!value) return "";
  const [date, time = "00:00"] = value.split("T");
  const normalizedTime = time.length === 5 ? `${time}:00` : time;
  return `${date}T${normalizedTime}+09:00`;
}

/**
 * FE4-8 parse 契約: "YYYY-MM-DD" → Date の解釈は 2 契約が併存する。
 * この関数はローカル getter で整形するため、DatePicker 系（ローカル正午 parse。
 * @see components/shared/DatePicker/DatePickerModel.ts の parseLocalDate）から得た
 * Date を渡す想定。line-reserve/shared-liff 側の UTC 深夜 parse 由来の Date とは
 * 契約が異なるため相互交換は日付ズレを起こす（禁止）。
 */
export function formatJSTWallDate(date: Date): string {
  return `${date.getFullYear()}-${padDatePart(date.getMonth() + 1)}-${padDatePart(date.getDate())}`;
}

export function formatJSTWallTime(date: Date): string {
  return `${padDatePart(date.getHours())}:${padDatePart(date.getMinutes())}`;
}

/**
 * FE6-6: 2 つの日付文字列（JST 00:00 起点）の日数差を返す。
 * refDate が未指定/不正な場合は 0 を返す。負値は 0 に丸める。
 */
export function daysSince(isoDate: string, refDate: string): number {
  if (!refDate) return 0;
  const from = new Date(`${isoDate}T00:00:00+09:00`);
  const to = new Date(`${refDate}T00:00:00+09:00`);
  if (isNaN(from.getTime()) || isNaN(to.getTime())) return 0;
  return Math.max(0, Math.floor((to.getTime() - from.getTime()) / (1000 * 60 * 60 * 24)));
}

/**
 * FE6-6: 現在時刻の JST 年月を "YYYY-MM" 形式で返す。
 */
export function currentJSTYearMonth(): string {
  const jst = new Date(Date.now() + JST_OFFSET_MS);
  const y = jst.getUTCFullYear();
  const m = padDatePart(jst.getUTCMonth() + 1);
  return `${y}-${m}`;
}

/** JST 当月の開始日・終了日 (YYYY-MM-DD)。未納一覧などの期間デフォルト用。 */
export function currentJSTMonthDateRange(): { start: string; end: string } {
  const ym = currentJSTYearMonth();
  const [year, month] = ym.split("-").map(Number);
  const lastDay = new Date(Date.UTC(year, month, 0)).getUTCDate();
  return {
    start: `${ym}-01`,
    end: `${ym}-${padDatePart(lastDay)}`,
  };
}
