import { describe, it, expect, afterEach, vi } from "vitest";

import { formatPetAge } from "./pet-age";

describe("formatPetAge", () => {
  it("誕生日と基準日から N歳Mヶ月 を算出する", () => {
    // 2015/04/14 生まれ、基準日 2024/10/14 → ちょうど 9歳6ヶ月
    expect(formatPetAge("2015-04-14", new Date(2024, 9, 14))).toBe("9歳6ヶ月");
  });

  it("基準日が誕生日より前なら月を 1 繰り下げる", () => {
    // 同月だが日が手前 → 5ヶ月
    expect(formatPetAge("2015-04-14", new Date(2024, 9, 10))).toBe("9歳5ヶ月");
  });

  it("月が負になる場合は年を繰り下げて 12 を足す", () => {
    // 2015/11/14 生まれ、基準日 2024/01/14 → 8歳2ヶ月
    expect(formatPetAge("2015-11-14", new Date(2024, 0, 14))).toBe("8歳2ヶ月");
  });

  it("ISO の time 部分が付いていても日付部分で算出する", () => {
    expect(formatPetAge("2015-04-14T00:00:00Z", new Date(2024, 9, 14))).toBe("9歳6ヶ月");
  });

  it("当日（誕生日と基準日が一致）は 0歳0ヶ月 を返す", () => {
    // 当日生まれのペットも捏造せず正しく算出する。
    expect(formatPetAge("2024-09-10", new Date(2024, 8, 10))).toBe("0歳0ヶ月");
  });

  it("不正な日付文字列は null を返す（捏造しない）", () => {
    expect(formatPetAge("", new Date(2024, 9, 14))).toBeNull();
    expect(formatPetAge("not-a-date", new Date(2024, 9, 14))).toBeNull();
    expect(formatPetAge("2015-13-01", new Date(2024, 9, 14))).toBeNull(); // 月 > 12
    expect(formatPetAge("2015-04-32", new Date(2024, 9, 14))).toBeNull(); // 日 > 31
  });

  it("未来日付は null を返す", () => {
    expect(formatPetAge("2030-01-01", new Date(2024, 0, 1))).toBeNull();
  });

  describe("基準日省略時は JST 壁時計で算出する", () => {
    afterEach(() => {
      vi.useRealTimers();
    });

    it("月境界で UTC とずれない（UTC 深夜 = JST 翌月）", () => {
      // 2026-06-30T21:00:00Z = 2026-07-01 06:00 JST。
      // ローカル(UTC)基準だと 6 月扱いで 1 ヶ月手前にずれるが、JST 壁時計なら 7 月で算出する。
      vi.useFakeTimers();
      vi.setSystemTime(new Date("2026-06-30T21:00:00Z"));
      // 2025-07-01 生まれ → JST 2026-07-01 でちょうど 1歳0ヶ月。
      expect(formatPetAge("2025-07-01")).toBe("1歳0ヶ月");
    });
  });
});
