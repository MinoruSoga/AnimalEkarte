import { describe, expect, it } from "vitest";

import {
  canAddColumnEntryOnDate,
  receptionColumnAddKind,
  resolveReceptionDateParam,
} from "./reception-model";

/**
 * EMR-243: `?date=YYYY-MM-DD` バリデータと非本日の「＋」ボタン方針を固定する。
 * バリデータは未指定・空・形式不正・非存在暦日を fallback（当日 JST）へ丸める契約。
 */
describe("resolveReceptionDateParam", () => {
  const fallback = "2026-10-02";

  it("有効な YYYY-MM-DD はそのまま返す", () => {
    expect(resolveReceptionDateParam("2026-10-05", fallback)).toBe("2026-10-05");
    expect(resolveReceptionDateParam(fallback, "2026-01-01")).toBe("2026-10-02");
  });

  it("未指定・空文字は fallback に丸める", () => {
    expect(resolveReceptionDateParam(null, fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("", fallback)).toBe(fallback);
  });

  it("形式不正は fallback に丸める", () => {
    expect(resolveReceptionDateParam("2026/10/05", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("2026-1-5", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("not-a-date", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("2026-10-05junk", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("2026-10-05T00:00:00Z", fallback)).toBe(fallback);
  });

  it("形式上正しくても非存在の暦日は fallback に丸める", () => {
    expect(resolveReceptionDateParam("2026-02-30", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("2026-13-01", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("2026-00-10", fallback)).toBe(fallback);
    expect(resolveReceptionDateParam("2026-04-31", fallback)).toBe(fallback);
  });

  it("うるう年の 2/29 は実在日として受理する", () => {
    expect(resolveReceptionDateParam("2028-02-29", fallback)).toBe("2028-02-29");
    expect(resolveReceptionDateParam("2027-02-29", fallback)).toBe(fallback);
  });
});

describe("receptionColumnAddKind / canAddColumnEntryOnDate", () => {
  it("受付予約列は通常予約、それ以外は当日受付 walk-in を発行する", () => {
    expect(receptionColumnAddKind("受付予約")).toBe("newReservation");
    expect(receptionColumnAddKind("受付済")).toBe("reception");
  });

  it("本日表示では 受付予約/受付済 にボタンを出し、診療中系には出さない", () => {
    expect(canAddColumnEntryOnDate("受付予約", true)).toBe(true);
    expect(canAddColumnEntryOnDate("受付済", true)).toBe(true);
    expect(canAddColumnEntryOnDate("診療中", true)).toBe(false);
    expect(canAddColumnEntryOnDate("会計待ち", true)).toBe(false);
    expect(canAddColumnEntryOnDate("会計済", true)).toBe(false);
  });

  it("非本日では reception 系（当日受付）のボタンだけを抑制し、通常予約は残す", () => {
    expect(canAddColumnEntryOnDate("受付予約", false)).toBe(true);
    expect(canAddColumnEntryOnDate("受付済", false)).toBe(false);
    expect(canAddColumnEntryOnDate("診療中", false)).toBe(false);
  });
});
