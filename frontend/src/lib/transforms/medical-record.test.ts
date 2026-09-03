import { describe, expect, it } from "vitest";

import {
  toBackendMedicalRecordStatus,
  transformMedicalRecord,
} from "./medical-record";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

const minimal: MedicalRecordResponse = {
  id: 1,
  clinic_id: 1,
  pet_id: 10,
  owner_id: 20,
  doctor_id: 3,
  record_no: "MR-001",
  date: "2026-03-25T00:00:00Z",
  status: "draft",
  version: 1,
  visit_count: 0,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("transformMedicalRecord (lib)", () => {
  it("maps draft/finalized status and deceased pet", () => {
    expect(transformMedicalRecord(minimal).status).toBe("作成中");
    expect(transformMedicalRecord({ ...minimal, status: "finalized" }).status).toBe("確定済");
    expect(
      transformMedicalRecord({
        ...minimal,
        pet: { id: 10, name: "ポチ", pet_number: "P-1", status: "deceased" },
      }).petIsDeceased,
    ).toBe(true);
  });

  it("maps visit_type labels", () => {
    expect(transformMedicalRecord({ ...minimal, visit_type: "first" }).visitType).toBe("初診");
    expect(transformMedicalRecord({ ...minimal, visit_type: "revisit" }).visitType).toBe("再診");
  });

  it("toBackendMedicalRecordStatus reverses UI labels", () => {
    expect(toBackendMedicalRecordStatus("作成中")).toBe("draft");
    expect(toBackendMedicalRecordStatus("確定済")).toBe("finalized");
    expect(toBackendMedicalRecordStatus("不明")).toBeUndefined();
  });
});
