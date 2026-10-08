import { describe, expect, it } from "vitest";
import { resolveCageNameById } from "./cage-name";

describe("resolveCageNameById (EMR-306)", () => {
  // Cage（id: string）と MasterItem（id: string|number）の混在形をカバーする。
  const cages: { id: string | number; name: string }[] = [
    { id: "1", name: "A-1" },
    { id: "2", name: "A-2" },
    { id: 3, name: "A-3" },
  ];

  it("一覧にある ID はケージ名を返す（string/number id 両対応）", () => {
    expect(resolveCageNameById(cages, "2")).toBe("A-2");
    expect(resolveCageNameById(cages, "3")).toBe("A-3");
  });

  it("一覧に無い ID は `ID <id>` にフォールバックする", () => {
    expect(resolveCageNameById(cages, "99")).toBe("ID 99");
  });
});
