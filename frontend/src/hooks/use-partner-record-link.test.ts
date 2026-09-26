import { renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { usePartnerRecordLink } from "./use-partner-record-link";

const mocks = vi.hoisted(() => ({
  reservationsData: [] as Array<Record<string, unknown>>,
  reservationsLoading: false,
  medicalRecordsData: [] as Array<Record<string, unknown>>,
  medicalRecordsLoading: false,
  useGetReservations: vi.fn(),
  useGetMedicalRecords: vi.fn(),
}));

vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: mocks.useGetReservations,
}));

vi.mock("@/hooks/use-medical-records", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-medical-records")>();
  return {
    ...actual,
    useGetMedicalRecords: mocks.useGetMedicalRecords,
  };
});

const BASE_INPUT = {
  petId: "10",
  visitDate: "2026-09-26",
} as const;

beforeEach(() => {
  mocks.reservationsData = [];
  mocks.reservationsLoading = false;
  mocks.medicalRecordsData = [];
  mocks.medicalRecordsLoading = false;
  mocks.useGetReservations.mockReset();
  mocks.useGetReservations.mockImplementation(() => ({
    data: mocks.reservationsData,
    isLoading: mocks.reservationsLoading,
  }));
  mocks.useGetMedicalRecords.mockReset();
  mocks.useGetMedicalRecords.mockImplementation(() => ({
    data: {
      data: mocks.medicalRecordsData,
      total: mocks.medicalRecordsData.length,
      page: 1,
      limit: 1,
    },
    isLoading: mocks.medicalRecordsLoading,
  }));
});

describe("usePartnerRecordLink — kind=trimming（カルテ→トリミング）", () => {
  it("同日・同一ペットの未完了 trimming appointment があれば open で appointmentId を付ける", () => {
    mocks.reservationsData = [
      { id: "55", category: "trimming", status: "in_consultation" },
      { id: "60", category: "general", status: "in_consultation" },
    ];

    const { result } = renderHook(() => usePartnerRecordLink({ kind: "trimming", ...BASE_INPUT }));

    expect(result.current.isLoading).toBe(false);
    expect(result.current.target).toEqual({
      mode: "open",
      href: "/trimming/new?visitDate=2026-09-26&petId=10&appointmentId=55",
      appointmentId: "55",
    });
  });

  it("trimming appointment が無ければ create で appointmentId 無しの record_shortcut 経路へ", () => {
    mocks.reservationsData = [{ id: "60", category: "general", status: "in_consultation" }];

    const { result } = renderHook(() => usePartnerRecordLink({ kind: "trimming", ...BASE_INPUT }));

    expect(result.current.target).toEqual({
      mode: "create",
      href: "/trimming/new?visitDate=2026-09-26&petId=10",
      appointmentId: undefined,
    });
  });

  it("完了/キャンセル済みの trimming appointment は相方なしとして create を返す", () => {
    mocks.reservationsData = [
      { id: "55", category: "trimming", status: "completed" },
      { id: "56", category: "trimming", status: "cancelled" },
      { id: "57", category: "trimming", status: "no_show" },
    ];

    const { result } = renderHook(() => usePartnerRecordLink({ kind: "trimming", ...BASE_INPUT }));

    expect(result.current.target?.mode).toBe("create");
    expect(result.current.target?.appointmentId).toBeUndefined();
  });

  it("reservations 解決中は isLoading=true で target=null", () => {
    mocks.reservationsLoading = true;

    const { result } = renderHook(() => usePartnerRecordLink({ kind: "trimming", ...BASE_INPUT }));

    expect(result.current.isLoading).toBe(true);
    expect(result.current.target).toBeNull();
  });
});

describe("usePartnerRecordLink — kind=medical-record（トリミング→カルテ）", () => {
  it("同日・同一ペットのカルテがあれば open で詳細ページへ", () => {
    mocks.medicalRecordsData = [{ id: "42" }];

    const { result } = renderHook(() =>
      usePartnerRecordLink({ kind: "medical-record", ...BASE_INPUT }),
    );

    expect(result.current.isLoading).toBe(false);
    expect(result.current.target).toEqual({
      mode: "open",
      href: "/medical-records/42",
      appointmentId: undefined,
    });
  });

  it("同日カルテが無ければ create で auto-create 経路（/medical-records/new）へ", () => {
    const { result } = renderHook(() =>
      usePartnerRecordLink({ kind: "medical-record", ...BASE_INPUT }),
    );

    expect(result.current.target).toEqual({
      mode: "create",
      href: "/medical-records/new?visitDate=2026-09-26&petId=10",
      appointmentId: undefined,
    });
  });

  it("petId/visitDate を query に反映する", () => {
    renderHook(() => usePartnerRecordLink({ kind: "medical-record", ...BASE_INPUT }));

    expect(mocks.useGetMedicalRecords).toHaveBeenCalledWith(
      { petId: "10", startDate: "2026-09-26", endDate: "2026-09-26", limit: 1 },
      { enabled: true },
    );
  });
});

describe("usePartnerRecordLink — ゲート", () => {
  it("enabled=false なら query を走らせず target=null", () => {
    const { result } = renderHook(() =>
      usePartnerRecordLink({ kind: "trimming", ...BASE_INPUT, enabled: false }),
    );

    expect(result.current.isLoading).toBe(false);
    expect(result.current.target).toBeNull();
    expect(mocks.useGetReservations).toHaveBeenCalledWith(
      expect.objectContaining({ enabled: false }),
    );
  });

  it("petId 未指定では query を走らせず target=null", () => {
    const { result } = renderHook(() =>
      usePartnerRecordLink({ kind: "medical-record", petId: undefined, visitDate: "2026-09-26" }),
    );

    expect(result.current.target).toBeNull();
    expect(mocks.useGetMedicalRecords).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ enabled: false }),
    );
  });

  it("visitDate が YYYY-MM-DD 以外では query を走らせず target=null", () => {
    const { result } = renderHook(() =>
      usePartnerRecordLink({ kind: "trimming", petId: "10", visitDate: "2026/09/26" }),
    );

    expect(result.current.target).toBeNull();
    expect(mocks.useGetReservations).toHaveBeenCalledWith(
      expect.objectContaining({ enabled: false }),
    );
  });

  it("kind と反対側の query は走らない（medical-record では reservations 非実行）", () => {
    renderHook(() => usePartnerRecordLink({ kind: "medical-record", ...BASE_INPUT }));

    expect(mocks.useGetMedicalRecords).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({ enabled: true }),
    );
    expect(mocks.useGetReservations).toHaveBeenCalledWith(
      expect.objectContaining({ enabled: false }),
    );
  });
});
