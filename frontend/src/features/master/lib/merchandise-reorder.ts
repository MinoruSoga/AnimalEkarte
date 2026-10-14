export type MerchandiseMoveDirection = "up" | "down";

/**
 * PR665 / EMR-401 C2-5: 50 行ページングを維持したまま、ページをまたぐ並べ替えを
 * 明示的な「前へ / 後へ」操作で行うための純関数。全件順序の ID 配列に対して
 * 1 つ前/後へ移動した新配列を返す。移動不要（先頭で up・末尾で down・不明 ID）は
 * null を返し、呼び出し側は何もしない。
 */
export function moveItemInOrder(
  ids: readonly string[],
  itemId: string,
  direction: MerchandiseMoveDirection,
): string[] | null {
  const from = ids.indexOf(itemId);
  if (from === -1) return null;
  const to = direction === "up" ? from - 1 : from + 1;
  if (to < 0 || to >= ids.length) return null;
  const next = [...ids];
  next.splice(from, 1);
  next.splice(to, 0, itemId);
  return next;
}
