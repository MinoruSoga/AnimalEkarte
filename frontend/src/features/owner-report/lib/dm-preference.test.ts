import { describe, it, expect } from "vitest";

import { formatDMPreference } from "./dm-preference";

describe("formatDMPreference", () => {
  it("true は『必要』を返す", () => {
    expect(formatDMPreference(true)).toBe("必要");
  });

  it("false は『不要』を返す", () => {
    expect(formatDMPreference(false)).toBe("不要");
  });

  it("undefined（未設定）は『未設定』を返し、false（不要）と混同しない", () => {
    // 既存レコードは値を持たない。未設定を「不要」と誤表示しないことを保証する。
    expect(formatDMPreference(undefined)).toBe("未設定");
    expect(formatDMPreference(undefined)).not.toBe(formatDMPreference(false));
  });
});
