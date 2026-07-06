// DatePicker の非コンポーネントロジック (日付パース/フォーマット・カレンダークラス定数)。
// コンポーネントファイル (DatePickerParts.tsx) から分離して
// react-refresh/only-export-components 違反を解消する。
import { C } from "@/lib/design-tokens";

export function parseLocalDate(iso: string): Date | undefined {
  if (!iso) return undefined;
  const [year, month, day] = iso.split("-").map(Number);
  if (!year || !month || !day) return undefined;
  return new Date(year, month - 1, day, 12, 0, 0);
}

export function formatIso(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function formatDisplay(date: Date): string {
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  const weekdays = ["日", "月", "火", "水", "木", "金", "土"];
  const weekday = weekdays[date.getDay()];
  return `${year}年${month}月${day}日（${weekday}）`;
}

export function formatShort(date: Date): string {
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${year}/${month}/${day}`;
}

export function formatSlash(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}/${month}/${day}`;
}

export function parseRangeValue(value: string): { from?: Date; to?: Date } {
  if (!value) return {};
  const [fromStr, toStr] = value.split("~");
  return {
    from: parseLocalDate(fromStr),
    to: toStr ? parseLocalDate(toStr) : undefined,
  };
}

export function parseDateInput(input: string): Date | null {
  const trimmed = input.trim();
  if (!trimmed) return null;

  let year: number;
  let month: number;
  let day: number;

  const compact = trimmed.match(/^(\d{4})(\d{2})(\d{2})$/);
  if (compact) {
    year = Number(compact[1]);
    month = Number(compact[2]);
    day = Number(compact[3]);
  } else {
    const separated = trimmed.match(/^(\d{4})[/-](\d{1,2})[/-](\d{1,2})$/);
    if (!separated) return null;
    year = Number(separated[1]);
    month = Number(separated[2]);
    day = Number(separated[3]);
  }

  if (month < 1 || month > 12 || day < 1 || day > 31) return null;
  const date = new Date(year, month - 1, day, 12, 0, 0);
  if (date.getMonth() !== month - 1 || date.getDate() !== day) return null;
  return date;
}

export const SINGLE_CALENDAR_CLASSES = {
  selected:
    `${C.bgPrimary} ${C.textWhite} hover:${C.bgPrimary} hover:text-white focus:${C.bgPrimary} focus:text-white`,
  today: `${C.bgPage} ${C.text}`,
  nav: "hidden",
  month_caption: "hidden",
};

export const RANGE_CALENDAR_CLASSES = {
  selected:
    `${C.bgPrimary} ${C.textWhite} hover:${C.bgPrimary} hover:text-white focus:${C.bgPrimary} focus:text-white`,
  today: `${C.bgPage} ${C.text}`,
  nav: "hidden",
};

export const TRIGGER_BASE =
  `flex h-11 w-full items-center justify-between rounded-md border ${C.borderMedium} bg-white px-3 text-sm ${C.text} transition-colors ${C.hoverBgPage} focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring`;
