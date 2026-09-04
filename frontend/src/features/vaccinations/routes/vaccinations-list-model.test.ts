import { describe, expect, it } from "vitest";

import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import type { VaccinationRecord } from "@/types";
import {
  buildVaccinationListQueryOptions,
  orderVaccinationListRows,
  vaccinationCreateHref,
  vaccinationListDetailHref,
} from "./vaccinations-list-model";

describe("vaccinationListDetailHref", () => {
  it("カルテに紐づく接種は予防接種タブ付きカルテへ行く", () => {
    expect(vaccinationListDetailHref({ id: "vac-1", medicalRecordId: "mr-1" })).toBe(
      "/medical-records/mr-1?tab=%E4%BA%88%E9%98%B2%E6%8E%A5%E7%A8%AE&vaccinationId=vac-1",
    );
  });

  it("未紐付けの接種は接種詳細画面へ行く", () => {
    expect(vaccinationListDetailHref({ id: "vac-1" })).toBe("/vaccinations/vac-1");
  });
});

describe("vaccinationCreateHref", () => {
  it("新規接種は独立フォーム /vaccinations/new?petId= へ行く（BUG-501）", () => {
    expect(vaccinationCreateHref("pet-1")).toBe("/vaccinations/new?petId=pet-1");
  });
});

describe("buildVaccinationListQueryOptions (BUG-502)", () => {
  const today = "2026-08-29";

  it("default window caps endDate at today so 2029 seed dates leave the first page", () => {
    const q = buildVaccinationListQueryOptions({ today });
    expect(q).toEqual({
      page: 1,
      limit: HISTORY_FETCH_LIMIT,
      startDate: undefined,
      endDate: today,
      search: undefined,
    });
    expect(q.endDate).toBe(today);
    expect(q.limit).toBe(100);
  });

  it("PACO (or any pet) search is preserved for server-side scoping", () => {
    const q = buildVaccinationListQueryOptions({ today, search: "  PACO  " });
    expect(q.search).toBe("PACO");
    expect(q.endDate).toBe(today);
    expect(q.limit).toBe(HISTORY_FETCH_LIMIT);
  });

  it("explicit date-range to wins over the today default", () => {
    const q = buildVaccinationListQueryOptions({
      today,
      dateRange: { from: "2026-01-01", to: "2029-12-31" },
    });
    expect(q.startDate).toBe("2026-01-01");
    expect(q.endDate).toBe("2029-12-31");
  });

  it("inverted date range fails closed to today-capped default window", () => {
    const q = buildVaccinationListQueryOptions({
      today,
      dateRange: { from: "2027-01-01", to: "2026-01-01" },
    });
    expect(q.startDate).toBeUndefined();
    expect(q.endDate).toBe(today);
  });
});

describe("orderVaccinationListRows (BUG-502)", () => {
  function row(
    partial: Partial<VaccinationRecord> & Pick<VaccinationRecord, "id">,
  ): VaccinationRecord {
    return {
      petId: "p1",
      ownerName: "owner",
      petName: "PACO",
      vaccineId: "v1",
      vaccineName: "混合",
      doctor: "doc",
      date: "2026-08-29",
      nextDate: "2026-09-10",
      ...partial,
    };
  }

  it("keeps near-term next_date ahead of far-future 2029 seed peers", () => {
    const nearTerm = row({ id: "1000000000", nextDate: "2026-09-10", date: "2026-08-29" });
    const seedFuture = row({
      id: "seed-2029",
      nextDate: "2029-12-01",
      date: "2029-12-01",
      petName: "PACO",
    });
    const mid = row({ id: "1000000001", nextDate: "2027-08-28", date: "2026-08-28" });

    const ordered = orderVaccinationListRows([seedFuture, mid, nearTerm]);
    expect(ordered.map((r) => r.id)).toEqual(["1000000000", "1000000001", "seed-2029"]);
  });

  it("sorts missing next_date after dated rows", () => {
    const withNext = row({ id: "a", nextDate: "2026-09-10" });
    const without = row({ id: "b", nextDate: undefined });
    expect(orderVaccinationListRows([without, withNext]).map((r) => r.id)).toEqual(["a", "b"]);
  });
});
