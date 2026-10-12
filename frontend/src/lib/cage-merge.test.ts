import { describe, expect, it } from "vitest";
import { isCageRepresentative, cageMembersByRepresentative } from "./cage-merge";

// Cage（id: string, mergedIntoCageId: string | null）と
// MasterItem（id: string | number, mergedIntoCageId?: string | null）の
// 混在形をカバーする最小構造。
interface Item {
  id: string | number;
  name: string;
  mergedIntoCageId?: string | null;
}

const item = (id: string | number, mergedIntoCageId?: string | null): Item => ({
  id,
  name: `ケージ${id}`,
  mergedIntoCageId,
});

describe("cage-merge (EMR-306)", () => {
  const cages: Item[] = [
    item("1"), // 代表（"2", "3" が指す）
    item("2", "1"),
    item(3, "1"),
    item("4"), // 独立
    item("5", null), // null sentinel = 未連結
    item("6"), // mergedIntoCageId 未設定（undefined）= 未連結
  ];

  it("isCageRepresentative: 誰かの mergedIntoCageId が指す id は代表", () => {
    expect(isCageRepresentative(cages, "1")).toBe(true);
    expect(isCageRepresentative(cages, 1)).toBe(true); // number id も string 正規化で一致
    expect(isCageRepresentative(cages, "2")).toBe(false); // member 自身は代表でない
    expect(isCageRepresentative(cages, "4")).toBe(false); // standalone
    expect(isCageRepresentative(cages, "99")).toBe(false); // 一覧外
    expect(isCageRepresentative([], "1")).toBe(false); // 空一覧
  });

  it("cageMembersByRepresentative: 代表 id → メンバー一覧（入力順・型保持）", () => {
    const byRep = cageMembersByRepresentative(cages);
    expect(byRep.get("1")?.map((m) => m.id)).toEqual(["2", 3]);
    // メンバー不在の代表 id・独立ケージのキーは含まれない
    expect(byRep.has("2")).toBe(false);
    expect(byRep.has("4")).toBe(false);
    expect(byRep.size).toBe(1);
    // null/undefined sentinel は索引に入らない
    expect([...byRep.values()].flat().map((m) => m.id)).not.toContain("5");
    expect([...byRep.values()].flat().map((m) => m.id)).not.toContain("6");
  });

  it("cageMembersByRepresentative: 全件未連結なら空 Map", () => {
    expect(cageMembersByRepresentative([item("1"), item("2")]).size).toBe(0);
  });
});
