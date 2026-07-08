// LineReservationSettingsForm/Sections 共有の非コンポーネントヘルパー。
// コンポーネントファイルから分離して react-refresh/only-export-components 違反を解消する。

export function asJsonb<T>(value: unknown, fallback: T): T {
  if (value == null) return fallback;
  if (typeof value === "string") {
    try {
      return JSON.parse(value) as T;
    } catch {
      return fallback;
    }
  }
  return value as T;
}

export function toDisplayTime(t: string): string {
  if (t.length === 4) return `${t.slice(0, 2)}:${t.slice(2)}`;
  return t;
}

export function toStorageTime(t: string): string {
  return t.replace(":", "");
}
