// LineReservationSettingsForm/Sections 共有の非コンポーネントヘルパー。
// コンポーネントファイルから分離して react-refresh/only-export-components 違反を解消する。

import type {
  BusinessHours,
  BreakHour,
  BusinessHoursByWeekday,
} from "../components/LineReservationSettingsFormSections";

// ── JSONB 型ガード ──
// バックエンドの JSONB カラムは tygo が `string /* []byte */` と宣言するが、
// 実運用では json.RawMessage 相当の生オブジェクト/配列として届く。
// 文字列で届いた場合(旧クライアント互換等)は JSON.parse も許容し、
// いずれの経路でも型ガードを通してから返すことで無検証キャストを排除する。

function isUnknownArray(value: unknown): value is unknown[] {
  return Array.isArray(value);
}

export function parseClosedWeekdays(value: unknown): number[] | null {
  let parsed: unknown = value;
  if (typeof parsed === "string") {
    try {
      parsed = JSON.parse(parsed);
    } catch {
      return null;
    }
  }
  if (parsed == null) return [];
  if (!isUnknownArray(parsed)) return null;

  const weekdays: number[] = [];
  for (const weekday of parsed) {
    if (typeof weekday === "number" && Number.isInteger(weekday) && weekday >= 0 && weekday <= 6) {
      weekdays.push(weekday);
      continue;
    }
    if (typeof weekday === "string" && /^[0-6]$/.test(weekday)) {
      weekdays.push(Number(weekday));
      continue;
    }
    return null;
  }
  return weekdays;
}

// parseClosedWeekdays の結果同士の等値比較（順序非依存・重複も多度数込みで比較）。
// null(不正値) は null 同士のみ等値。
export function sameClosedWeekdays(a: number[] | null, b: number[] | null): boolean {
  if (a === null || b === null) return a === b;
  if (a.length !== b.length) return false;
  const remaining = new Map<number, number>();
  for (const weekday of b) {
    remaining.set(weekday, (remaining.get(weekday) ?? 0) + 1);
  }
  for (const weekday of a) {
    const count = remaining.get(weekday) ?? 0;
    if (count === 0) return false;
    remaining.set(weekday, count - 1);
  }
  return true;
}

export function isStringArray(v: unknown): v is string[] {
  return Array.isArray(v) && v.every((item) => typeof item === "string");
}

export function isBusinessHours(v: unknown): v is BusinessHours {
  if (typeof v !== "object" || v === null || Array.isArray(v)) return false;
  const record = v as Record<string, unknown>;
  return typeof record.start === "string" && typeof record.end === "string";
}

export function isBreakHourArray(v: unknown): v is BreakHour[] {
  return Array.isArray(v) && v.every(isBusinessHours);
}

export function isBusinessHoursByWeekday(v: unknown): v is BusinessHoursByWeekday {
  if (typeof v !== "object" || v === null || Array.isArray(v)) return false;
  return Object.values(v as Record<string, unknown>).every(isBusinessHours);
}

export function asJsonb<T>(value: unknown, fallback: T, isT: (v: unknown) => v is T): T {
  if (value == null) return fallback;
  if (typeof value === "string") {
    let parsed: unknown;
    try {
      parsed = JSON.parse(value);
    } catch {
      return fallback;
    }
    return isT(parsed) ? parsed : fallback;
  }
  return isT(value) ? value : fallback;
}

export function toDisplayTime(t: string): string {
  if (t.length === 4) return `${t.slice(0, 2)}:${t.slice(2)}`;
  return t;
}

export function toStorageTime(t: string): string {
  return t.replace(":", "");
}
