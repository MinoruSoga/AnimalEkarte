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
  const instant = new Date(Date.UTC(
    date.getFullYear(),
    date.getMonth(),
    date.getDate(),
    date.getHours() - 9,
    date.getMinutes(),
    date.getSeconds(),
    date.getMilliseconds(),
  ));
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
