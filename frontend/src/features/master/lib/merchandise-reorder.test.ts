import { describe, expect, it } from "vitest";
import { moveItemInOrder } from "./merchandise-reorder";

const ids = Array.from({ length: 101 }, (_, i) => String(i + 1));
// 境界をまたぐ移動（51 を前へ / 50 を後へ）の完全な期待順序: 1..49, 51, 50, 52..101
const expectedAfterBoundarySwap = [...ids.slice(0, 49), "51", "50", ...ids.slice(51)];

describe("moveItemInOrder (PR665 / EMR-401 C2-5)", () => {
  it("ページ境界をまたいで前へ移動できる（51件目 → 50件目の直前=1ページ目末尾）", () => {
    const next = moveItemInOrder(ids, "51", "up");
    expect(next).not.toBeNull();
    expect(next).toEqual(expectedAfterBoundarySwap);
  });

  it("ページ境界をまたいで後へ移動できる（50件目 → 51件目の直後=2ページ目先頭）", () => {
    const next = moveItemInOrder(ids, "50", "down");
    expect(next).not.toBeNull();
    expect(next).toEqual(expectedAfterBoundarySwap);
  });

  it("先頭で up・末尾で down・不明 ID は移動しない（null）", () => {
    expect(moveItemInOrder(ids, "1", "up")).toBeNull();
    expect(moveItemInOrder(ids, "101", "down")).toBeNull();
    expect(moveItemInOrder(ids, "999", "up")).toBeNull();
  });
});
