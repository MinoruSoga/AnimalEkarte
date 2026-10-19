import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw";
import { QueryClient } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";

import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";

import {
  fetchMedicineDoseParamsOnce,
  medicineDoseParamsQueryKey,
  useMedicineDoseParams,
} from "./medicine-dose-lookup";
import { getAllMedicinesMaster, useGetAllMedicinesMaster } from "@/hooks/use-treatment-master";
import { QueryClientProvider } from "@tanstack/react-query";
import { createElement } from "react";
import { useTreatmentDoseGate } from "../hooks/use-treatment-dose-gate";
import type { Treatment } from "../types";

function clientWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return {
    queryClient,
    wrapper: ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children),
  };
}

describe("PR700: 複写先クリニック別マスタ/dose-params スコープ", () => {
  it("row dose lookup forwards doseContext.clinicId", async () => {
    const seen: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines/42/dose-params", ({ request }) => {
        seen.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );
    const { queryClient, wrapper } = clientWrapper();
    renderHook(
      () =>
        useTreatmentDoseGate(
          { item_type: "medicine", medicine_id: "42", quantity: 1 } as Treatment,
          { medicines: undefined, weightKg: null, petSpecies: "犬", clinicId: "3" },
          "1",
          false,
        ),
      { wrapper },
    );
    await waitFor(() =>
      expect(queryClient.getQueryData(medicineDoseParamsQueryKey("42", "3"))).toEqual([]),
    );
    expect(seen).toEqual(["3"]);
    expect(queryClient.getQueryData(medicineDoseParamsQueryKey("42"))).toBeUndefined();
  });
  it("warm destination medicines stay isolated from conflicting global cache", async () => {
    let requests = 0;
    server.use(
      http.get("*/v1/masters/medicines", () => {
        requests++;
        return HttpResponse.json([]);
      }),
    );
    const { queryClient, wrapper } = clientWrapper();
    const global = [{ id: "1", price: 10, isActive: true }];
    const scoped = [{ id: "42", price: 3303, isActive: true }];
    queryClient.setQueryData(queryKeys.masters.category("medicines"), global);
    queryClient.setQueryData(queryKeys.masters.categoryForClinic("medicines", "3"), scoped);
    const { result } = renderHook(() => useGetAllMedicinesMaster({ clinicId: "3" }), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(scoped);
    expect(queryClient.getQueryData(queryKeys.masters.category("medicines"))).toEqual(global);
    expect(requests).toBe(0);
  });

  it("disabled medicines do not fetch, and empty dose clinic keeps the legacy key", async () => {
    let requests = 0;
    server.use(
      http.get("*/v1/masters/medicines", () => {
        requests++;
        return HttpResponse.json([]);
      }),
    );
    const { wrapper } = clientWrapper();
    const { result } = renderHook(
      () => useGetAllMedicinesMaster({ enabled: false, clinicId: "3" }),
      { wrapper },
    );
    expect(result.current.fetchStatus).toBe("idle");
    expect(requests).toBe(0);
    expect(medicineDoseParamsQueryKey("42", "")).toEqual(medicineDoseParamsQueryKey("42"));
  });
  it("medicines は truthy clinicId で scoped key + X-Clinic-ID を送る", async () => {
    const seen: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines", ({ request }) => {
        seen.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([{ id: 42, name: "薬A", price: 3303, is_active: true }]);
      }),
    );
    const { queryClient, wrapper } = clientWrapper();
    const { result } = renderHook(() => useGetAllMedicinesMaster({ clinicId: "3" }), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(seen).toEqual(["3"]);
    expect(
      queryClient.getQueryData(queryKeys.masters.categoryForClinic("medicines", "3")),
    ).toBeDefined();
    expect(queryClient.getQueryData(queryKeys.masters.category("medicines"))).toBeUndefined();
  });

  it("clinicId 省略・空文字はグローバル key + 暗黙ヘッダを維持する", async () => {
    const seen: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines", ({ request }) => {
        seen.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );
    const { queryClient, wrapper } = clientWrapper();
    for (const clinicId of [undefined, ""]) {
      const { result } = renderHook(() => useGetAllMedicinesMaster({ clinicId }), { wrapper });
      await waitFor(() => expect(result.current.isSuccess).toBe(true));
    }
    expect(seen.length).toBeGreaterThanOrEqual(1);
    expect(seen.every((h) => h !== "3")).toBe(true);
    expect(queryClient.getQueryData(queryKeys.masters.category("medicines"))).toBeDefined();
    expect(
      queryClient.getQueryData(queryKeys.masters.categoryForClinic("medicines", "3")),
    ).toBeUndefined();
  });

  it("getAllMedicinesMaster は truthy clinicId でヘッダを送る（fetcher 直テスト）", async () => {
    const seen: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines", ({ request }) => {
        seen.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );
    await getAllMedicinesMaster("9");
    await getAllMedicinesMaster();
    expect(seen).toEqual(["9", null]);
  });

  it("dose-params は truthy clinicId で scoped key + X-Clinic-ID を送り、既定は旧キー", async () => {
    const seen: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines/42/dose-params", ({ request }) => {
        seen.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );
    const { queryClient, wrapper } = clientWrapper();
    const scoped = renderHook(() => useMedicineDoseParams("42", "3"), { wrapper });
    await waitFor(() => expect(scoped.result.current.isSuccess).toBe(true));
    const legacy = renderHook(() => useMedicineDoseParams("42"), { wrapper });
    await waitFor(() => expect(legacy.result.current.isSuccess).toBe(true));
    expect(seen).toEqual(["3", null]);
    expect(medicineDoseParamsQueryKey("42", "3")).not.toEqual(medicineDoseParamsQueryKey("42"));
    expect(queryClient.getQueryData(medicineDoseParamsQueryKey("42", "3"))).toBeDefined();
    expect(queryClient.getQueryData(medicineDoseParamsQueryKey("42"))).toBeDefined();
  });

  it("fetchMedicineDoseParamsOnce も clinicId でヘッダを送る", async () => {
    const seen: Array<string | null> = [];
    server.use(
      http.get("*/v1/masters/medicines/7/dose-params", ({ request }) => {
        seen.push(request.headers.get("X-Clinic-ID"));
        return HttpResponse.json([]);
      }),
    );
    await fetchMedicineDoseParamsOnce("7", "5");
    await fetchMedicineDoseParamsOnce("7");
    expect(seen).toEqual(["5", null]);
  });
});
