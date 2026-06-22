import { CATEGORY_LABELS, DISPLAY_CATEGORIES } from "./constants";

export interface CategorySubtotal {
  label: string;
  total: number;
}

// category_breakdown JSONB の形:
//   { categories: { [category]: { [paymentMethodName]: amount } }, ... }
// 生成型は any (json.RawMessage) のため unknown として安全にパースし、
// DISPLAY_CATEGORIES でグルーピングした部門別小計を表示順で返す。
// 未知・欠損・不正データは握りつぶさず空配列 / 取りこぼし無しで扱う。
export function summarizeCategoryTotals(raw: unknown): CategorySubtotal[] {
  const categories = extractCategories(raw);
  if (!categories) return [];

  // 生カテゴリ → 支払方法合計
  const rawTotals = new Map<string, number>();
  for (const [category, methods] of Object.entries(categories)) {
    if (!methods || typeof methods !== "object") continue;
    let sum = 0;
    for (const amount of Object.values(methods as Record<string, unknown>)) {
      if (typeof amount === "number" && Number.isFinite(amount)) sum += amount;
    }
    rawTotals.set(category, (rawTotals.get(category) ?? 0) + sum);
  }

  const grouped: CategorySubtotal[] = [];
  const consumed = new Set<string>();

  // 表示区分順に集約（診療 = examination + test + procedure + medicine など）
  for (const group of DISPLAY_CATEGORIES) {
    let total = 0;
    for (const key of group.keys) {
      total += rawTotals.get(key) ?? 0;
      consumed.add(key);
    }
    if (total !== 0) grouped.push({ label: group.label, total });
  }

  // DISPLAY_CATEGORIES に存在しない未知カテゴリも取りこぼさない
  for (const [category, total] of rawTotals) {
    if (consumed.has(category) || total === 0) continue;
    grouped.push({ label: CATEGORY_LABELS[category] ?? category, total });
  }

  return grouped;
}

function extractCategories(raw: unknown): Record<string, unknown> | null {
  if (!raw || typeof raw !== "object") return null;
  const categories = (raw as { categories?: unknown }).categories;
  if (!categories || typeof categories !== "object") return null;
  return categories as Record<string, unknown>;
}
