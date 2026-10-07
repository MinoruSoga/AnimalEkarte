import { describe, it, expect } from "vitest";
import { findCageRepresentative } from "./cage-side-panel-model";
import type { Cage } from "../api/cages";

function cage(id: string, mergedIntoCageId: number | null = null): Cage {
  return {
    id,
    name: `${id}番ケージ`,
    cageType: "general",
    cageSize: "medium",
    price: 1000,
    isActive: true,
    description: "",
    sortOrder: Number(id),
    mergedIntoCageId,
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-01T00:00:00Z",
  };
}

// EMR-306: 連結メンバー行・サイドパネルの代表ケージ名解決
describe("findCageRepresentative (EMR-306)", () => {
  it("mergedIntoCageId が指すケージを一覧から返す", () => {
    const rep = cage("1");
    const member = cage("2", 1);
    expect(findCageRepresentative([rep, member], member)).toBe(rep);
  });

  it("cage が null、または未連結（mergedIntoCageId=null）なら null", () => {
    const rep = cage("1");
    expect(findCageRepresentative([rep], null)).toBeNull();
    expect(findCageRepresentative([rep], rep)).toBeNull();
  });

  it("代表ケージが一覧に存在しない場合は null", () => {
    const member = cage("2", 99);
    expect(findCageRepresentative([member], member)).toBeNull();
  });
});
