import { describe, expect, it } from "vitest";
import { formatDiff } from "./cash-register-history-model";

describe("formatDiff", () => {
  it("正の差額はプラスを付けず ¥ 区切りで表示する", () => {
    expect(formatDiff(1234)).toBe("¥1,234");
  });

  it("負の差額は マイナス→¥→数字 の順で表示する", () => {
    expect(formatDiff(-1234)).toBe("-¥1,234");
  });

  it("0 は ¥0 と表示する", () => {
    expect(formatDiff(0)).toBe("¥0");
  });
});
