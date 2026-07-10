import { describe, it, expect } from "vitest";
import { calcAge } from "./PetDeceasedBanner";

describe("PetDeceasedBanner calcAge(deceasedAt, birthDate)", () => {
  it("誕生日前日: 死亡日が誕生日の前日の場合は誕生日を迎える前として計算する", () => {
    expect(calcAge("2024-06-14", "2020-06-15")).toBe(3);
  });

  it("誕生日当日: 死亡日が誕生日と同日の場合はその年の誕生日を迎えたものとして計算する", () => {
    expect(calcAge("2024-06-15", "2020-06-15")).toBe(4);
  });

  it("うるう年2/29生まれ: 非うるう年の2/28時点では誕生日未到来として計算する", () => {
    expect(calcAge("2023-02-28", "2020-02-29")).toBe(2);
  });

  it("うるう年2/29生まれ: 非うるう年の3/1時点では誕生日到来済みとして計算する", () => {
    expect(calcAge("2023-03-01", "2020-02-29")).toBe(3);
  });

  it("未来日付: 誕生日が死亡日より未来の場合は0でクランプする", () => {
    expect(calcAge("2024-01-01", "2025-01-01")).toBe(0);
  });
});
