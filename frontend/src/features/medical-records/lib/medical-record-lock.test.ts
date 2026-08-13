import { describe, expect, it } from "vitest";
import { isMedicalRecordFinalizedStatus } from "./medical-record-lock";

describe("isMedicalRecordFinalizedStatus (BUG-035)", () => {
  it("locks only the FE domain label 確定済", () => {
    expect(isMedicalRecordFinalizedStatus("確定済")).toBe(true);
  });

  it("does not lock draft or unknown / wire values", () => {
    expect(isMedicalRecordFinalizedStatus("作成中")).toBe(false);
    expect(isMedicalRecordFinalizedStatus("finalized")).toBe(false);
    expect(isMedicalRecordFinalizedStatus("確定済み")).toBe(false);
    expect(isMedicalRecordFinalizedStatus(undefined)).toBe(false);
    expect(isMedicalRecordFinalizedStatus(null)).toBe(false);
    expect(isMedicalRecordFinalizedStatus("")).toBe(false);
  });
});
