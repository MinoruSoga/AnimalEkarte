import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { MedicalRecord } from "../api/transforms";
import { useApplyMedicalRecord } from "./use-apply-medical-record";

function recordFixture(overrides: Partial<MedicalRecord> = {}): MedicalRecord {
  return {
    id: "record-a",
    chiefComplaint: "",
    chiefComplaintTypeId: null,
    plan: "",
    assessment: "",
    notes: "",
    visitType: "再診",
    nextVisitRecommendedDate: "",
    version: 1,
    ...overrides,
  } as MedicalRecord;
}

function buildArgs(existingRecord: MedicalRecord | undefined) {
  return {
    existingRecord,
    setChiefComplaint: vi.fn(),
    setChiefComplaintTypeId: vi.fn(),
    setTreatmentPolicy: vi.fn(),
    setPlan: vi.fn(),
    setAssessment: vi.fn(),
    setVisitType: vi.fn(),
    setNextVisitDate: vi.fn(),
  };
}

describe("useApplyMedicalRecord chief_complaint_type blank/clear/reload/switch", () => {
  it("BUG-406: non-null type hydrates into state", async () => {
    const args = buildArgs(recordFixture({ chiefComplaintTypeId: 5 }));
    renderHook(() => useApplyMedicalRecord(args));

    await waitFor(() => {
      expect(args.setChiefComplaintTypeId).toHaveBeenCalledWith(5);
    });
  });

  it("reload after intentional clear: server null clears local type (C3)", async () => {
    const setChiefComplaintTypeId = vi.fn();
    const { rerender } = renderHook(
      ({ existingRecord }: { existingRecord: MedicalRecord }) =>
        useApplyMedicalRecord({
          ...buildArgs(existingRecord),
          setChiefComplaintTypeId,
        }),
      {
        initialProps: {
          existingRecord: recordFixture({ id: "record-a", chiefComplaintTypeId: 5 }),
        },
      },
    );

    await waitFor(() => {
      expect(setChiefComplaintTypeId).toHaveBeenCalledWith(5);
    });
    setChiefComplaintTypeId.mockClear();

    rerender({
      existingRecord: recordFixture({ id: "record-a", chiefComplaintTypeId: null }),
    });

    await waitFor(() => {
      expect(setChiefComplaintTypeId).toHaveBeenCalledWith(null);
    });
  });

  it("chart switch from typed record to unset record clears type", async () => {
    const setChiefComplaintTypeId = vi.fn();
    const { rerender } = renderHook(
      ({ existingRecord }: { existingRecord: MedicalRecord }) =>
        useApplyMedicalRecord({
          ...buildArgs(existingRecord),
          setChiefComplaintTypeId,
        }),
      {
        initialProps: {
          existingRecord: recordFixture({ id: "record-a", chiefComplaintTypeId: 9 }),
        },
      },
    );

    await waitFor(() => {
      expect(setChiefComplaintTypeId).toHaveBeenCalledWith(9);
    });
    setChiefComplaintTypeId.mockClear();

    rerender({
      existingRecord: recordFixture({ id: "record-b", chiefComplaintTypeId: null }),
    });

    await waitFor(() => {
      expect(setChiefComplaintTypeId).toHaveBeenCalledWith(null);
    });
  });
});
