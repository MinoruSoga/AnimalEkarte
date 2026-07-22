import { todayJSTISO } from "@/lib/jst-date";

export function isEstimateExpired(
  validUntil: string | null | undefined,
  today = todayJSTISO(),
): boolean {
  if (!validUntil) return false;
  return validUntil.slice(0, 10) < today;
}
