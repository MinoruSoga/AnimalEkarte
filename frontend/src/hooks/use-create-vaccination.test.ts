import { createElement, type ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { server } from "@/testing/mocks/node";

import { useCreateVaccination, type CreateVaccinationRequest } from "./use-create-vaccination";

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

function backendVaccination(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    clinic_id: 1,
    vaccine_id: 7,
    date: "2026-09-21T00:00:00+09:00",
    pet_id: 1,
    medical_record_id: 99,
    lot1: "LOT-A",
    lot2: "",
    lot3: "",
    lot4: "",
    remarks: "",
    supplemental: "",
    next_date: "2026-10-19T00:00:00+09:00",
    created_at: "2026-09-21T00:00:00+09:00",
    updated_at: "2026-09-21T00:00:00+09:00",
    ...overrides,
  };
}

describe("useCreateVaccination (SLACK-VACCINE-MULTI)", () => {
  beforeEach(() => {
    vi.mocked(handleApiError).mockClear();
  });

  it("単件 POST /api/v1/vaccinations のみ呼び、amount を送らず batch 経路を使わない", async () => {
    const paths: string[] = [];
    const bodies: Record<string, unknown>[] = [];
    let batchHits = 0;

    server.use(
      http.post("/api/v1/vaccinations", async ({ request }) => {
        paths.push(new URL(request.url).pathname);
        bodies.push((await request.json()) as Record<string, unknown>);
        return HttpResponse.json(backendVaccination(), { status: 201 });
      }),
      http.post("/api/v1/vaccinations/batch", () => {
        batchHits += 1;
        return HttpResponse.json({ error: "batch not implemented" }, { status: 404 });
      }),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateVaccination(), {
      wrapper: createWrapper(queryClient),
    });

    const req: CreateVaccinationRequest = {
      pet_id: 1,
      medical_record_id: 99,
      vaccine_id: 7,
      date: "2026-09-21",
      lot1: "LOT-A",
      next_date: "2026-10-19",
      next_schedule_type: "4weeks",
    };

    await act(async () => {
      await result.current.mutateAsync(req);
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(paths).toEqual(["/api/v1/vaccinations"]);
    expect(batchHits).toBe(0);
    expect(bodies).toHaveLength(1);
    expect(bodies[0]).toEqual(req);
    expect(bodies[0]).not.toHaveProperty("amount");
    expect(result.current.data?.date).toBe("2026-09-21");
    expect(result.current.data?.nextDate).toBe("2026-10-19");
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: queryKeys.vaccinations.all() });
  });

  it("登録失敗（4xx）では handleApiError を呼び成功扱いにしない", async () => {
    server.use(
      http.post("/api/v1/vaccinations", () =>
        HttpResponse.json({ error: "接種日は今日以前の日付を入力してください" }, { status: 400 }),
      ),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCreateVaccination(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({
          vaccine_id: 7,
          date: "2099-01-01",
          pet_id: 1,
        }),
      ).rejects.toBeTruthy();
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(handleApiError).toHaveBeenCalledWith(expect.anything(), "ワクチン接種登録");
    expect(result.current.isSuccess).toBe(false);
  });

  it("同日 2–3 件は順次の単件 POST で保存し、各 body の vaccine/lot/next_date が混ざらない", async () => {
    const bodies: Record<string, unknown>[] = [];
    let seq = 0;

    server.use(
      http.post("/api/v1/vaccinations", async ({ request }) => {
        seq += 1;
        const body = (await request.json()) as Record<string, unknown>;
        bodies.push(body);
        return HttpResponse.json(
          backendVaccination({
            id: seq,
            vaccine_id: body.vaccine_id,
            date: "2026-09-21T00:00:00+09:00",
            lot1: body.lot1 ?? "",
            next_date:
              typeof body.next_date === "string" ? `${body.next_date}T00:00:00+09:00` : null,
          }),
          { status: 201 },
        );
      }),
    );

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(() => useCreateVaccination(), {
      wrapper: createWrapper(queryClient),
    });

    const shots: CreateVaccinationRequest[] = [
      {
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 1,
        date: "2026-09-21",
        lot1: "LOT-A",
        next_date: "2026-10-19",
      },
      {
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 2,
        date: "2026-09-21",
        lot1: "LOT-B",
        next_date: "2026-10-26",
      },
      {
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 3,
        date: "2026-09-21",
        lot1: "LOT-C",
        next_date: "2026-11-02",
      },
    ];

    const saved: Array<
      Awaited<ReturnType<ReturnType<typeof useCreateVaccination>["mutateAsync"]>>
    > = [];
    for (const shot of shots) {
      // Sequential single POSTs only — no batch body / parallel fan-in.
      const row = await act(async () => result.current.mutateAsync(shot));
      saved.push(row);
    }

    expect(bodies).toHaveLength(3);
    expect(bodies.map((b) => b.vaccine_id)).toEqual([1, 2, 3]);
    expect(bodies.map((b) => b.lot1)).toEqual(["LOT-A", "LOT-B", "LOT-C"]);
    expect(bodies.map((b) => b.next_date)).toEqual(["2026-10-19", "2026-10-26", "2026-11-02"]);
    expect(bodies.every((b) => b.date === "2026-09-21")).toBe(true);
    expect(bodies.every((b) => !("amount" in b))).toBe(true);

    expect(saved.map((r) => r.id)).toEqual(["1", "2", "3"]);
    expect(saved.map((r) => r.vaccineId)).toEqual(["1", "2", "3"]);
    expect(saved.map((r) => r.lot1)).toEqual(["LOT-A", "LOT-B", "LOT-C"]);
    expect(saved.map((r) => r.nextDate)).toEqual(["2026-10-19", "2026-10-26", "2026-11-02"]);
    // First row fields remain as posted; later shots do not rewrite earlier bodies.
    expect(bodies[0]).toMatchObject(shots[0]);
  });
});
