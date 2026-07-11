import { describe, it, expect } from "vitest";
import * as gen from "@/types/generated/models";
import type { ShiftType } from "./index";

// FE4-3: tygo 生成定数は typeof で string に退化するため、
// 手書き literal union の値集合が生成定数のランタイム値と完全一致することを機械固定する。
describe("shifts union drift", () => {
  it("ShiftType の値集合が ShiftType* 生成定数と一致する", () => {
    const values: ShiftType[] = ["full", "morning", "afternoon", "off", "paid_leave"];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.ShiftTypeFull,
        gen.ShiftTypeMorning,
        gen.ShiftTypeAfternoon,
        gen.ShiftTypeOff,
        gen.ShiftTypePaidLeave,
      ]),
    );
  });
});
