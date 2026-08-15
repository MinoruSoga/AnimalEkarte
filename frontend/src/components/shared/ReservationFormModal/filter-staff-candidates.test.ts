import { describe, expect, it } from "vitest";
import {
  filterStaffCandidatesByCapability,
  type ReservationStaffCapabilityLike,
} from "./filter-staff-candidates";

const staff = (id: number) => ({ id, name: `S${id}` });

describe("filterStaffCandidatesByCapability", () => {
  it("returns all candidates when no reservation type is selected", () => {
    const candidates = [staff(1), staff(2)];
    expect(
      filterStaffCandidatesByCapability(candidates, null, undefined),
    ).toEqual(candidates);
  });

  it("fail-closed: empty when capability metadata is pending", () => {
    const candidates = [staff(1), staff(2)];
    expect(
      filterStaffCandidatesByCapability(candidates, "5", undefined),
    ).toEqual([]);
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
    const result = filterStaffCandidatesByCapability(
      [staff(10), staff(11)],
      "5",
      map,
    );
    expect(result.map((s) => s.id)).toEqual([11]);
  });

  it("fail-closed: staff missing from map is excluded", () => {
    const map = new Map<string, ReservationStaffCapabilityLike>([
      ["11", { id: 11, capable_courses: [{ id: 5 }] }],
    ]);
    const result = filterStaffCandidatesByCapability(
      [staff(10), staff(11)],
      "5",
      map,
    );
    expect(result.map((s) => s.id)).toEqual([11]);
  });

  it("empty capable_courses excludes the staff for every type", () => {
    const map = new Map<string, ReservationStaffCapabilityLike>([
      ["10", { id: 10, capable_courses: [] }],
    ]);
    expect(
      filterStaffCandidatesByCapability([staff(10)], "5", map),
    ).toEqual([]);
  });
});
