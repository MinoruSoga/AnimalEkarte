import { describe, expect, it } from "vitest";

import {
  useGetMedicalRecords,
  getMedicalRecords,
  transformToHistoryItem,
} from "./use-medical-records";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

describe("use-medical-records (FE-RC-015 elevation)", () => {
  it("exports the list hook and query function from hooks (no features import)", () => {
    expect(typeof useGetMedicalRecords).toBe("function");
    expect(typeof getMedicalRecords).toBe("function");
  });
});

const minimalResponse: MedicalRecordResponse = {
  id: 1,
  clinic_id: 1,
  record_no: "MR-001",
  date: "2026-03-25T00:00:00Z",
  status: "finalized",
  version: 1,
  visit_count: 0,
  created_at: "2026-03-25T00:00:00Z",
  updated_at: "2026-03-25T00:00:00Z",
};

describe("useGetPetMedicalHistory の transformToHistoryItem（EMR-182 前回複写）", () => {
  it("inquiry の複写可能値を copySource へマップする", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      inquiry: {
        id: 1,
        chief_complaint: "元気がない",
        notes: "安静と投薬",
        chief_complaint_type_id: 5,
      },
    });
    expect(result.copySource).toEqual({
      chiefComplaint: "元気がない",
      treatmentPolicy: "安静と投薬",
      chiefComplaintTypeId: 5,
    });
  });

  it("copySource は空文字・未設定の項目を含めない", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      inquiry: { id: 1, chief_complaint: "", notes: "", chief_complaint_type_id: 2 },
    });
    expect(result.copySource).toEqual({ chiefComplaintTypeId: 2 });
  });

  it("inquiry 未設定のとき copySource は undefined", () => {
    expect(
      transformToHistoryItem({ ...minimalResponse, inquiry: undefined }).copySource,
    ).toBeUndefined();
  });

  it("複写可能な値が無いとき copySource は undefined", () => {
    const result = transformToHistoryItem({
      ...minimalResponse,
      inquiry: { id: 1, chief_complaint: "", notes: "" },
    });
    expect(result.copySource).toBeUndefined();
  });
});
