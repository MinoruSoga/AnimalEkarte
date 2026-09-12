import { describe, expect, it } from "vitest";
import {
  filterStaffCandidatesByCapability,
  resolveStaffSelectionEligibility,
  STAFF_ORPHAN_REASON_MESSAGE,
  type ReservationStaffCapabilityLike,
} from "./filter-staff-candidates";

const staff = (id: number) => ({ id, name: `S${id}` });

describe("filterStaffCandidatesByCapability", () => {
  it("returns all candidates when no reservation type is selected", () => {
    const candidates = [staff(1), staff(2)];
    expect(filterStaffCandidatesByCapability(candidates, null, undefined)).toEqual(candidates);
  });

  it("fail-closed: empty when capability metadata is pending", () => {
    const candidates = [staff(1), staff(2)];
    expect(filterStaffCandidatesByCapability(candidates, "5", undefined)).toEqual([]);
  });

  it("keeps only staff with affirmative capable_courses for the type", () => {
    const map = new Map<string, ReservationStaffCapabilityLike>([
      [
        "10",
        {
          id: 10,
          capable_courses: [],
        },
      ],
      [
        "11",
        {
          id: 11,
          capable_courses: [{ id: 5, name: "トリミング" }],
        },
      ],
    ]);
    const result = filterStaffCandidatesByCapability([staff(10), staff(11)], "5", map);
    expect(result.map((s) => s.id)).toEqual([11]);
  });

  it("fail-closed: staff missing from map is excluded", () => {
    const map = new Map<string, ReservationStaffCapabilityLike>([
      ["11", { id: 11, capable_courses: [{ id: 5 }] }],
    ]);
    const result = filterStaffCandidatesByCapability([staff(10), staff(11)], "5", map);
    expect(result.map((s) => s.id)).toEqual([11]);
  });

  it("empty capable_courses excludes the staff for every type", () => {
    const map = new Map<string, ReservationStaffCapabilityLike>([
      ["10", { id: 10, capable_courses: [] }],
    ]);
    expect(filterStaffCandidatesByCapability([staff(10)], "5", map)).toEqual([]);
  });
});

describe("resolveStaffSelectionEligibility", () => {
  const names = new Map([
    ["10", "三井"],
    ["11", "鈴木"],
  ]);

  it("keeps eligible selection without orphan reason", () => {
    const result = resolveStaffSelectionEligibility({
      doctorId: "11",
      eligibleOptionIds: new Set(["11"]),
      nameById: names,
      candidatesSettled: true,
      hasQueryError: false,
    });
    expect(result).toEqual({
      isConfirmedOrphan: false,
      displayLabel: "鈴木",
      reasonMessage: null,
    });
  });

  it("marks confirmed orphan only after candidates settle successfully", () => {
    const pending = resolveStaffSelectionEligibility({
      doctorId: "10",
      eligibleOptionIds: new Set(),
      nameById: names,
      candidatesSettled: false,
      hasQueryError: false,
    });
    expect(pending.isConfirmedOrphan).toBe(false);
    expect(pending.displayLabel).toBe("三井");
    expect(pending.reasonMessage).toBeNull();

    const errored = resolveStaffSelectionEligibility({
      doctorId: "10",
      eligibleOptionIds: new Set(),
      nameById: names,
      candidatesSettled: false,
      hasQueryError: true,
    });
    expect(errored.isConfirmedOrphan).toBe(false);
    expect(errored.reasonMessage).toBeNull();

    const orphan = resolveStaffSelectionEligibility({
      doctorId: "10",
      eligibleOptionIds: new Set(["11"]),
      nameById: names,
      candidatesSettled: true,
      hasQueryError: false,
    });
    expect(orphan).toEqual({
      isConfirmedOrphan: true,
      displayLabel: "三井",
      reasonMessage: STAFF_ORPHAN_REASON_MESSAGE,
    });
  });
});
