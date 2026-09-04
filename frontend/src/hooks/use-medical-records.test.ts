import { describe, expect, it } from "vitest";

import { useGetMedicalRecords, getMedicalRecords } from "./use-medical-records";

describe("use-medical-records (FE-RC-015 elevation)", () => {
  it("exports the list hook and query function from hooks (no features import)", () => {
    expect(typeof useGetMedicalRecords).toBe("function");
    expect(typeof getMedicalRecords).toBe("function");
  });
});
