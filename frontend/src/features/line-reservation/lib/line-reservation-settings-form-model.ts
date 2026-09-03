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
