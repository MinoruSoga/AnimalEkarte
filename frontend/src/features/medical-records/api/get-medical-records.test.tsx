import { describe, it, expect } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { useGetMedicalRecords } from "./get-medical-records";
import type { BackendMedicalRecord } from "./types";

function makeBackendRecord(id: number): BackendMedicalRecord {
  return {
    id,
    clinic_id: 1,
    pet_id: 10,
    owner_id: 20,
    doctor_id: 3,
    record_no: `MR-${id}`,
    date: "2026-03-25T00:00:00Z",
    status: "draft",
    version: 1,
    visit_count: 0,
    created_at: "2026-03-25T00:00:00Z",
    updated_at: "2026-03-25T00:00:00Z",
  };
}

const createWrapper = createTestWrapper;

// BUG-B1 回帰防止: page/limit が常に送信され、旧DB由来を含む全件へページングできること
describe("useGetMedicalRecords", () => {
  it("filters 未指定でも page=1・limit=20 を既定送信する", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/medical-records", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({ data: [makeBackendRecord(1)], total: 425524, page: 1, limit: 20 });
      }),
    );

    const { result } = renderHook(() => useGetMedicalRecords(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("page")).toBe("1");
    expect(capturedUrl?.searchParams.get("limit")).toBe("20");
    // 旧DB由来 425,524件級の total をそのまま透過すること（B-1 の 20件打ち切りに回帰しない）
    expect(result.current.data?.total).toBe(425524);
    expect(result.current.data?.data).toHaveLength(1);
  });

  it("page 2 を指定すると page=2 で再取得する（旧DB帯レコードへの到達を保証）", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/medical-records", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({ data: [makeBackendRecord(21)], total: 425524, page: 2, limit: 20 });
      }),
    );

    const { result } = renderHook(() => useGetMedicalRecords({ page: 2, limit: 20 }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("page")).toBe("2");
    expect(result.current.data?.page).toBe(2);
  });

  it("search/status/doctor_id/animal_species_id/date range/clinic_ids をクエリに含める", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/medical-records", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20 });
      }),
    );

    const { result } = renderHook(
      () =>
        useGetMedicalRecords({
          search: "田中",
          status: "finalized",
          doctorId: "3",
          animalSpeciesId: "1",
          startDate: "2026-01-01",
          endDate: "2026-01-31",
          clinicIds: ["1", "2"],
        }),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("search")).toBe("田中");
    expect(capturedUrl?.searchParams.get("status")).toBe("finalized");
    expect(capturedUrl?.searchParams.get("doctor_id")).toBe("3");
    expect(capturedUrl?.searchParams.get("animal_species_id")).toBe("1");
    expect(capturedUrl?.searchParams.get("start_date")).toBe("2026-01-01");
    expect(capturedUrl?.searchParams.get("end_date")).toBe("2026-01-31");
    expect(capturedUrl?.searchParams.get("clinic_ids")).toBe("1,2");
  });

  it("clinicIds が1件以下のときは clinic_ids を送信しない（単一医院スコープ）", async () => {
    let capturedUrl: URL | undefined;
    server.use(
      http.get("/api/v1/medical-records", ({ request }) => {
        capturedUrl = new URL(request.url);
        return HttpResponse.json({ data: [], total: 0, page: 1, limit: 20 });
      }),
    );

    const { result } = renderHook(() => useGetMedicalRecords({ clinicIds: ["1"] }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(capturedUrl?.searchParams.get("clinic_ids")).toBeNull();
  });
});
