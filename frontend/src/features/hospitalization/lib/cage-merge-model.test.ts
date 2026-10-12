import { describe, expect, it } from "vitest";
import type { MasterItem } from "@/types";
import {
  isCageRepresentative,
  isInMergedGroup,
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
    expect(isMergeDragSource({ cages, cage: source, hasOccupant: false })).toBe(true);
    expect(isMergeDragSource({ cages, cage: member, hasOccupant: false })).toBe(false);
    expect(isMergeDragSource({ cages, cage: rep, hasOccupant: false })).toBe(false);
    expect(isMergeDragSource({ cages, cage: inactive, hasOccupant: false })).toBe(false);
    expect(isMergeDragSource({ cages, cage: source, hasOccupant: true })).toBe(false);
  });

  it("isMergeEligibleTarget: self/member/代表/占有/inactive は不可、standalone→standalone は可", () => {
    // valid: 11 → 15（standalone→standalone 連結）
    const standalone2 = cage("15");
    expect(
      isMergeEligibleTarget({
        cages: [...cages, standalone2],
        draggedCageId: "11",
        targetCageId: "15",
        isOccupied: noOccupant,
      }),
    ).toBe(true);
    // 代表 target（10）は不可: BE は既存グループへの member 追加を 409 拒否
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "10",
        isOccupied: noOccupant,
      }),
    ).toBe(false);
    // self
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "11",
        isOccupied: noOccupant,
      }),
    ).toBe(false);
    // member target（12 は member）
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "12",
        isOccupied: noOccupant,
      }),
    ).toBe(false);
    // occupied target（10 に occupant がいる）
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "10",
        isOccupied: (id) => id === "10",
      }),
    ).toBe(false);
    // occupied dragged
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "10",
        isOccupied: (id) => id === "11",
      }),
    ).toBe(false);
    // dragged が member（source 不適格）
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "12",
        targetCageId: "11",
        isOccupied: noOccupant,
      }),
    ).toBe(false);
    // inactive target
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "14",
        isOccupied: noOccupant,
      }),
    ).toBe(false);
    // 一覧外 id
    expect(
      isMergeEligibleTarget({
        cages,
        draggedCageId: "11",
        targetCageId: "999",
        isOccupied: noOccupant,
      }),
    ).toBe(false);
  });

  // Grok review HIGH-1: 連結スワップ事前拒否の判定関数
  it("isInMergedGroup: メンバーと代表は true、独立・一覧外・空値は false", () => {
    expect(isInMergedGroup(cages, "12")).toBe(true); // member
    expect(isInMergedGroup(cages, "10")).toBe(true); // representative
    expect(isInMergedGroup(cages, "11")).toBe(false); // standalone
    expect(isInMergedGroup(cages, "14")).toBe(false); // inactive standalone
    expect(isInMergedGroup(cages, "999")).toBe(false); // 一覧外
    expect(isInMergedGroup(cages, null)).toBe(false);
    expect(isInMergedGroup(cages, undefined)).toBe(false);
    expect(isInMergedGroup(cages, "")).toBe(false);
    expect(isInMergedGroup([], "10")).toBe(false); // ケージ未取得時は検知不可（BE が最終判定）
  });
});
