import { C } from "@/lib/design-tokens";

export function parseHighlightDate(dateParam: string | null): { year: number; month: number } | null {
  if (!dateParam) return null;
  const matched = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateParam);
  if (!matched) return null;
  return { year: Number(matched[1]), month: Number(matched[2]) };
}

export function diffClass(diff: number): string {
  return diff === 0 ? C.textStatusGreen : C.danger;
}

export function formatDiff(diff: number): string {
  return `${diff >= 0 ? "+" : ""}${diff.toLocaleString()}`;
}
