import { describe, expect, it } from "vitest";

import { calcAgeAt } from "./calc-age";

describe("calcAgeAt(asOf, birthDate)", () => {
  it("誕生日前日: 基準日が誕生日の前日の場合は誕生日を迎える前として計算する", () => {
    expect(calcAgeAt("2024-06-14", "2020-06-15")).toBe(3);
  });

  it("誕生日当日: 基準日が誕生日と同日の場合はその年の誕生日を迎えたものとして計算する", () => {
    expect(calcAgeAt("2024-06-15", "2020-06-15")).toBe(4);
  });

  it("うるう年2/29生まれ: 非うるう年の2/28時点では誕生日未到来として計算する", () => {
    expect(calcAgeAt("2023-02-28", "2020-02-29")).toBe(2);
  });

  it("うるう年2/29生まれ: 非うるう年の3/1時点では誕生日到来済みとして計算する", () => {
    expect(calcAgeAt("2023-03-01", "2020-02-29")).toBe(3);
  });

  it("未来日付: 生年月日が基準日より未来の場合は0でクランプする", () => {
    expect(calcAgeAt("2024-01-01", "2025-01-01")).toBe(0);
  });

  it("Date インスタンスを直接渡しても計算できる", () => {
    expect(calcAgeAt(new Date("2024-06-15T00:00:00Z"), new Date("2020-06-15T00:00:00Z"))).toBe(4);
  });
});
