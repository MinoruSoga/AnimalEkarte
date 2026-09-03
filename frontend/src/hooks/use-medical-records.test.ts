import { describe, expect, it } from "vitest";

import { useGetMedicalRecords } from "./use-medical-records";
import { useGetMedicalRecords as featureHook } from "../features/medical-records/api/get-medical-records";

describe("use-medical-records (FE-RC-015 elevation)", () => {
  it("re-exports the feature list hook (same reference)", () => {
    expect(useGetMedicalRecords).toBe(featureHook);
  });
});
