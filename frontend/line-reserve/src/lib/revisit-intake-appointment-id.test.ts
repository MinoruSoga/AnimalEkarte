import { describe, expect, it } from "vitest";

import {
  isRevisitIntakeAppointmentIdUsable,
  toRevisitIntakeAppointmentId,
} from "./revisit-intake-appointment-id";

describe("revisit-intake appointment id guard", () => {
  it.each(["1", "9007199254740991", "9007199254740992", "9007199254740993", "9223372036854775807"])(
    "正規形の十進文字列 %s はそのまま使える（Number 変換しない）",
    (id) => {
      expect(isRevisitIntakeAppointmentIdUsable(id)).toBe(true);
      expect(toRevisitIntakeAppointmentId(id)).toBe(id);
    },
  );

  it("2^53 境界の隣接 ID は衝突しない", () => {
    expect(toRevisitIntakeAppointmentId("9007199254740992")).toBe("9007199254740992");
    expect(toRevisitIntakeAppointmentId("9007199254740993")).toBe("9007199254740993");
  });

  it.each([
    ["0", "0"],
    ["先頭ゼロ", "01"],
    ["負数", "-1"],
    ["符号", "+1"],
    ["小数", "1.0"],
    ["指数", "1e3"],
    ["空白", " 1"],
    ["空文字", ""],
    ["int64 超", "9223372036854775808"],
  ])("%s（%s）は使えず null", (_label, id) => {
    expect(isRevisitIntakeAppointmentIdUsable(id)).toBe(false);
    expect(toRevisitIntakeAppointmentId(id)).toBeNull();
  });
});
