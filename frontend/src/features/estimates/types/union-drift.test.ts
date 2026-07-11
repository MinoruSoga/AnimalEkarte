import { describe, it, expect } from "vitest";
import * as gen from "@/types/generated/models";
import type { EstimateStatus } from "./index";

// FE4-2: tygo 生成定数は typeof で string に退化するため、
// 手書き literal union の値集合が生成定数のランタイム値と完全一致することを機械固定する。
describe("estimates union drift", () => {
  it("EstimateStatus の値集合が EstimateStatus* 生成定数と一致する", () => {
    const values: EstimateStatus[] = ["draft", "sent", "approved", "rejected"];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.EstimateStatusDraft,
        gen.EstimateStatusSent,
        gen.EstimateStatusApproved,
        gen.EstimateStatusRejected,
      ]),
    );
  });
});
