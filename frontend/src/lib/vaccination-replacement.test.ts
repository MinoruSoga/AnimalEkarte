import { describe, expect, it } from "vitest";

import {
  collectReplacedVaccinationIds,
  isExternalClinicVaccinationName,
} from "./vaccination-replacement";

function record(
  overrides: Partial<{
    id: string;
    petId: string;
    date: string;
    vaccineId: string;
    vaccineName: string;
  }>,
) {
  return {
    id: "v1",
    petId: "pet-1",
    date: "2026-01-01",
    vaccineId: "10",
    vaccineName: "混合ワクチン",
    ...overrides,
  };
}

describe("collectReplacedVaccinationIds (EMR-428)", () => {
  it("同じペット・同じ vaccineId で接種日がより後の記録がある古い記録は置き換わり", () => {
    const records = [
      record({ id: "old", date: "2026-01-01" }),
      record({ id: "new", date: "2026-02-01" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.has("old")).toBe(true);
    expect(replaced.has("new")).toBe(false);
  });

  it("別ワクチン（vaccineId が違う）の後続では置き換わらない", () => {
    const records = [
      record({ id: "old", date: "2026-01-01", vaccineId: "10" }),
      record({ id: "new", date: "2026-02-01", vaccineId: "11" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.size).toBe(0);
  });

  it("同日に2件あればどちらも置き換わらない", () => {
    const records = [
      record({ id: "a", date: "2026-01-01" }),
      record({ id: "b", date: "2026-01-01" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.size).toBe(0);
  });

  it("vaccineName が「他院」で始まる記録は記録ごとに判定する（置き換わらない）", () => {
    const records = [
      record({ id: "old", date: "2026-01-01", vaccineName: "他院で接種したワクチン" }),
      record({ id: "new", date: "2026-02-01" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.has("old")).toBe(false);
  });

  it("「他院」で始まる記録は、後続があっても同じワクチンとみなさず他を置き換えない", () => {
    const records = [
      record({ id: "normal", date: "2026-01-01" }),
      record({ id: "external", date: "2026-02-01", vaccineName: "他院（名称不明）" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.has("normal")).toBe(false);
    expect(replaced.has("external")).toBe(false);
  });

  it('vaccineId が "0" の記録は記録ごとに判定する（置き換わらない）', () => {
    const records = [
      record({ id: "old", date: "2026-01-01", vaccineId: "0" }),
      record({ id: "new", date: "2026-02-01", vaccineId: "0" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.size).toBe(0);
  });

  it("別ペットの後続では置き換わらない", () => {
    const records = [
      record({ id: "old", petId: "pet-1", date: "2026-01-01" }),
      record({ id: "new", petId: "pet-2", date: "2026-02-01" }),
    ];

    const replaced = collectReplacedVaccinationIds(records);

    expect(replaced.size).toBe(0);
  });
});

describe("isExternalClinicVaccinationName (EMR-428 / EMR-427 同規則)", () => {
  it("前方一致「他院」で判定する", () => {
    expect(isExternalClinicVaccinationName("他院で接種")).toBe(true);
    expect(isExternalClinicVaccinationName("他院")).toBe(true);
    expect(isExternalClinicVaccinationName(" previously 他院")).toBe(false);
    expect(isExternalClinicVaccinationName("混合ワクチン")).toBe(false);
  });
});
