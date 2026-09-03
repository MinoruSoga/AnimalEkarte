import { describe, expect, it } from "vitest";

import {
  isTargetExamGroup,
  orderExamGroupsForTarget,
} from "../lib/medical-record-examination-model";

describe("isTargetExamGroup", () => {
  it("examId が一致すれば対象", () => {
    expect(isTargetExamGroup(1013826, "1013826")).toBe(true);
  });

  it("examId が無ければ対象ではない", () => {
    expect(isTargetExamGroup(1013826, null)).toBe(false);
  });
});

describe("orderExamGroupsForTarget", () => {
  it("対象検査を先頭に出す", () => {
    expect(
      orderExamGroupsForTarget([{ id: 1 }, { id: 2 }, { id: 3 }], "2").map((g) => g.id),
    ).toEqual([2, 1, 3]);
  });

  it("examId が無ければ順序を変えない", () => {
    expect(orderExamGroupsForTarget([{ id: 1 }, { id: 2 }], null).map((g) => g.id)).toEqual([1, 2]);
  });
});
