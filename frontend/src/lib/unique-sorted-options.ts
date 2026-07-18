/**
 * js-combine-iterations: 配列からユニークな値を抽出し、ソート済み SelectItem 形式で返す。
 * `Array.from(new Set(items.map(accessor).filter(Boolean))).sort().map(...)` の 3-4 パスを
 * 1パス（Set 構築）+ sort + map に統合する。
 */
export function uniqueSortedOptions<T>(
  items: T[],
  accessor: (item: T) => string | null | undefined,
): { value: string; label: string }[] {
  const set = new Set<string>();
  for (const item of items) {
    const val = accessor(item);
    if (val) set.add(val);
  }
  return Array.from(set)
    .sort()
    .map((v) => ({ value: v, label: v }));
}
