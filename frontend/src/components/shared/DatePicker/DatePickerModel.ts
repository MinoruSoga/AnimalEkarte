// DatePicker の非コンポーネントロジック (日付パース/フォーマット・カレンダークラス定数)。
// コンポーネントファイル (DatePickerParts.tsx) から分離して
// react-refresh/only-export-components 違反を解消する。
import { C } from "@/lib/design-tokens";
import { formatJSTWallDate } from "@/lib/jst-date";
import { formatJapaneseDate } from "@/lib/format/date";

/**
 * FE4-8 parse 契約: "YYYY-MM-DD" → Date の解釈は 2 契約が併存する
 * （line-reserve/shared-liff = UTC 深夜 parse ⇔ この DatePicker 系 = ローカル正午 parse）。
 * 相互交換は日付ズレを起こすため禁止（本関数の戻り値を line-reserve 側の
 * UTC 深夜前提コードへ渡さない・その逆も禁止）。
 */
export function parseLocalDate(iso: string): Date | undefined {
  if (!iso) return undefined;
  const [year, month, day] = iso.split("-").map(Number);
  if (!year || !month || !day) return undefined;
  return new Date(year, month - 1, day, 12, 0, 0);
}

/** FE4-8: lib/jst-date.ts の formatJSTWallDate（ローカル getter・同一契約）へ委譲。 */
export function formatIso(date: Date): string {
  return formatJSTWallDate(date);
}

/** FE4-8: utils/format/date.ts の formatJapaneseDate（ローカル getter・同一契約）へ委譲。 */
export function formatDisplay(date: Date): string {
  return formatJapaneseDate(date);
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
  selected: `${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.focusBgBrand} ${C.textOnBrand}`,
  today: `${C.bgPage} ${C.text}`,
  nav: "hidden",
  month_caption: "hidden",
};

export const RANGE_CALENDAR_CLASSES = {
  selected: `${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.focusBgBrand} ${C.textOnBrand}`,
  today: `${C.bgPage} ${C.text}`,
  nav: "hidden",
};

export const TRIGGER_BASE = `flex h-11 min-w-0 w-full items-center justify-between rounded-md border ${C.borderMedium} bg-white px-3 text-sm ${C.text} transition-colors ${C.hoverBgPage} focus-within:outline-none focus-within:ring-1 focus-within:ring-ring ${C.focusRingMedium}`;
