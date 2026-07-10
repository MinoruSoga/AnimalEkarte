import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { calcAge } from "./PetDeceasedRecordButton";

describe("PetDeceasedRecordButton calcAge(birthDate)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("誕生日前日: 現在日が誕生日の前日の場合は誕生日を迎える前として計算する", () => {
    vi.setSystemTime(new Date("2024-06-14"));
    expect(calcAge("2020-06-15")).toBe("3歳");
  });

  it("誕生日当日: 現在日が誕生日と同日の場合はその年の誕生日を迎えたものとして計算する", () => {
    vi.setSystemTime(new Date("2024-06-15"));
    expect(calcAge("2020-06-15")).toBe("4歳");
  });

  it("うるう年2/29生まれ: 非うるう年の2/28時点では誕生日未到来として計算する", () => {
    vi.setSystemTime(new Date("2023-02-28"));
    expect(calcAge("2020-02-29")).toBe("2歳");
  });

  it("うるう年2/29生まれ: 非うるう年の3/1時点では誕生日到来済みとして計算する", () => {
    vi.setSystemTime(new Date("2023-03-01"));
    expect(calcAge("2020-02-29")).toBe("3歳");
  });

  it("未来日付: 誕生日が現在日より未来の場合は0歳でクランプする", () => {
    vi.setSystemTime(new Date("2024-01-01"));
    expect(calcAge("2025-01-01")).toBe("0歳");
  });
});
