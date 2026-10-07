import { describe, expect, it } from "vitest";
import type { MasterItem } from "@/types";
import {
  isCageRepresentative,
  isMergeDragSource,
  isMergeEligibleTarget,
  isMergedMember,
  mergeCageDragId,
  resolveRepresentativeName,
} from "./cage-merge-model";

const cage = (id: string, overrides: Partial<MasterItem> = {}): MasterItem => ({
  id,
  name: `ケージ${id}`,
  category: "犬舎",
  price: 0,
  status: "active",
  ...overrides,
});

const rep = cage("10");
const source = cage("11");
const member = cage("12", { mergedIntoCageId: "10" });
const inactive = cage("14", { status: "inactive" });
const cages = [rep, source, member, inactive];

const noOccupant = () => false;

describe("cage-merge-model (EMR-306)", () => {
  it("mergeCageDragId は merge-cage- prefix の source id を返す（patient drag id と衝突しない）", () => {
    expect(mergeCageDragId("10")).toBe("merge-cage-10");
  });

  it("isMergedMember / isCageRepresentative / resolveRepresentativeName", () => {
    expect(isMergedMember(member)).toBe(true);
    expect(isMergedMember(rep)).toBe(false);
    expect(isCageRepresentative(cages, "10")).toBe(true);
    expect(isCageRepresentative(cages, "11")).toBe(false);
    expect(resolveRepresentativeName(cages, "10")).toBe("ケージ10");
    // 代表が一覧外なら ID フォールバック
    expect(resolveRepresentativeName(cages, "99")).toBe("ID 99");
  });

  it("isMergeDragSource: member・代表・占有・無効ケージは source 不可", () => {
    expect(isMergeDragSource({ cages, cage: source, hasVisibleOccupant: false })).toBe(true);
    expect(isMergeDragSource({ cages, cage: member, hasVisibleOccupant: false })).toBe(false);
    expect(isMergeDragSource({ cages, cage: rep, hasVisibleOccupant: false })).toBe(false);
    expect(isMergeDragSource({ cages, cage: inactive, hasVisibleOccupant: false })).toBe(false);
    expect(isMergeDragSource({ cages, cage: source, hasVisibleOccupant: true })).toBe(false);
  });

  it("isMergeEligibleTarget: self/member/代表/占有/inactive は不可、standalone→standalone は可", () => {
    // valid: 11 → 15（standalone→standalone 連結）
    const standalone2 = cage("15");
    expect(
      isMergeEligibleTarget({
        cages: [...cages, standalone2],
        draggedCageId: "11",
        targetCageId: "15",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(true);
    // 代表 target（10）は不可: BE は既存グループへの member 追加を 409 拒否
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "10",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(false);
    // self
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "11",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(false);
    // member target（12 は member）
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "12",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(false);
    // occupied target（10 に occupant がいる）
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "10",
        hasVisibleOccupant: (id) => id === "10",
      }),
    ).toBe(false);
    // occupied dragged
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "10",
        hasVisibleOccupant: (id) => id === "11",
      }),
    ).toBe(false);
    // dragged が member（source 不適格）
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "12",
        targetCageId: "11",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(false);
    // inactive target
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "14",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(false);
    // 一覧外 id
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "999",
        hasVisibleOccupant: noOccupant,
      }),
    ).toBe(false);
  });
});
