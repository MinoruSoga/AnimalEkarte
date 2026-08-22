import { describe, expect, it } from "vitest";
import {
  OCCUPATION_FILTER_ALL,
  OCCUPATION_FILTER_UNSET,
  filterStaffsByOccupation,
} from "./get-staffs";
import type { ShiftStaff } from "../types";

const STAFFS: ShiftStaff[] = [
  { id: "1", name: "A", occupationId: "10", occupationName: "医師" },
  { id: "2", name: "B", occupationId: "20", occupationName: "看護師" },
  { id: "3", name: "C", occupationId: null, occupationName: null },
];

describe("filterStaffsByOccupation", () => {
  it("all は全員", () => {
    expect(filterStaffsByOccupation(STAFFS, OCCUPATION_FILTER_ALL)).toHaveLength(3);
  });

  it("occupation id で絞る", () => {
    expect(filterStaffsByOccupation(STAFFS, "10").map((s) => s.id)).toEqual(["1"]);
  });

  it("unset は職種なし", () => {
    expect(filterStaffsByOccupation(STAFFS, OCCUPATION_FILTER_UNSET).map((s) => s.id)).toEqual([
      "3",
    ]);
  });
});
