import { CATEGORY_LABELS, DISPLAY_CATEGORIES } from "./constants";

export interface CategorySubtotal {
  label: string;
  total: number;
  /**
   * DEC-40: 未分類・要確認行のみ設定。
   * - number: 締め時点 snapshot の件数
   * - null: 旧 snapshot でフィールド欠落 → UI は「記録なし」
   * - undefined: 件数を表示しない部門
   */
  count?: number | null;
}

// category_breakdown JSONB の形:
//   {
//     categories: { [category]: { [paymentMethodName]: amount } },
//     unclassified_other_count?: number,  // DEC-40
//     ...
//   }
// 生成型は any (json.RawMessage) のため unknown として安全にパースし、
// DISPLAY_CATEGORIES でグルーピングした部門別小計を表示順で返す。
// 未知・欠損・不正データは握りつぶさず空配列 / 取りこぼし無しで扱う。
export function summarizeCategoryTotals(raw: unknown): CategorySubtotal[] {
  const categories = extractCategories(raw);
  if (!categories) return [];

  const unclassified = extractUnclassifiedOtherCount(raw);

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
    const isOther = group.keys.includes("other");
    if (isOther) {
      // 金額 > 0、または件数記録ありで > 0 のとき行を出す。
      // 件数フィールド欠落かつ金額 0 は出さない。金額 > 0 で件数欠落なら count=null（記録なし）。
      const hasCount = unclassified.recorded && unclassified.count > 0;
      if (total !== 0 || hasCount) {
        grouped.push({
          label: group.label,
          total,
          count: unclassified.recorded ? unclassified.count : null,
        });
      }
      continue;
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

/**
 * DEC-40: snapshot から unclassified_other_count を安全に取り出す。
 * recorded=false は旧データ（フィールド欠落）→ UI は「記録なし」（0 扱いにしない）。
 */
export function extractUnclassifiedOtherCount(raw: unknown): {
  recorded: boolean;
  count: number;
} {
  if (!raw || typeof raw !== "object") return { recorded: false, count: 0 };
  if (!Object.prototype.hasOwnProperty.call(raw, "unclassified_other_count")) {
    return { recorded: false, count: 0 };
  }
  const value = (raw as { unclassified_other_count: unknown }).unclassified_other_count;
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return { recorded: false, count: 0 };
  }
  return { recorded: true, count: value };
}

function extractCategories(raw: unknown): Record<string, unknown> | null {
  if (!raw || typeof raw !== "object") return null;
  const categories = (raw as { categories?: unknown }).categories;
  if (!categories || typeof categories !== "object") return null;
  return categories as Record<string, unknown>;
}
