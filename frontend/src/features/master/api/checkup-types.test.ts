import { createElement, type ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";

import { server } from "@/testing/mocks/node";
import { queryKeys } from "@/lib/query-keys";
import {
  useCreateCheckupType,
  useDeleteCheckupType,
  useGetAllCheckupTypes,
  useReorderCheckupTypes,
  useUpdateCheckupType,
} from "./checkup-types";

afterEach(() => {
  server.resetHandlers();
});

// EMR-456: 健診種別の保険設定（is_non_insurance）の wire 接続を固定する。
// 生成型 CheckupType は model の json:"-" のため is_non_insurance を持たず、
// 実 wire は domain-owned response（checkup_type_response.go）である点を前提にする。
function makeRawCheckupType(overrides: Record<string, unknown> = {}) {
  return {
    id: 7,
    clinic_id: 1,
    name: "年次健診",
    price: 1200,
    is_active: true,
    description: "",
    interval: "",
    target_age: "",
    sort_order: 1,
    is_non_insurance: false,
    version: 1,
    created_at: "2026-10-20T00:00:00Z",
    updated_at: "2026-10-20T00:00:00Z",
    ...overrides,
  };
}

function createQueryWrapper(queryClient: QueryClient) {
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
}

describe("useGetAllCheckupTypes (EMR-456)", () => {
  it("maps the wire is_non_insurance true/false into insuranceSetting and keeps version", async () => {
    server.use(
      http.get("/api/v1/masters/checkup-types", () =>
        HttpResponse.json([
          makeRawCheckupType({ id: 1, name: "年次健診", is_non_insurance: true }),
          makeRawCheckupType({ id: 2, name: "健康診断", is_non_insurance: false }),
        ]),
      ),
    );

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useGetAllCheckupTypes(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      expect.objectContaining({
        id: "1",
        masterKind: "checkup",
        insuranceSetting: true,
        version: 1,
      }),
      expect.objectContaining({
        id: "2",
        masterKind: "checkup",
        insuranceSetting: false,
        version: 1,
      }),
    ]);
  });

  it("keeps legacy null rows readable without inventing a stored value", async () => {
    server.use(
      http.get("/api/v1/masters/checkup-types", () =>
        HttpResponse.json([makeRawCheckupType({ is_non_insurance: null })]),
      ),
    );

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useGetAllCheckupTypes(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    // 旧欠落は未設定のまま。専有API単独の旧共有transformはaliasを省略し、
    // EMR-454との統合後はnullを返す。どちらもbooleanへ補完しない。
    const row = result.current.data?.[0];
    expect(row?.insuranceSetting).toBeNull();
    const compatibilitySetting = row && "isNonInsurance" in row ? row.isNonInsurance : undefined;
    expect(compatibilitySetting ?? null).toBeNull();
  });

  it("treats a response without the field (older backend) as unknown, not false", async () => {
    const raw: Record<string, unknown> = makeRawCheckupType();
    delete raw.is_non_insurance;
    delete raw.version;
    server.use(http.get("/api/v1/masters/checkup-types", () => HttpResponse.json([raw])));

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useGetAllCheckupTypes(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]).toMatchObject({ insuranceSetting: null, version: null });
  });
});

describe("checkup type mutations (EMR-456)", () => {
  it("sends the explicit create value and returns the persisted insurance setting", async () => {
    let receivedBody: unknown;
    server.use(
      http.post("/api/v1/masters/checkup-types", async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(makeRawCheckupType({ is_non_insurance: true }), { status: 201 });
      }),
    );

    const queryClient = createTestQueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateCheckupType(), {
      wrapper: createQueryWrapper(queryClient),
    });

    let created: Awaited<ReturnType<typeof result.current.mutateAsync>> | undefined;
    await act(async () => {
      created = await result.current.mutateAsync({ name: "年次健診", is_non_insurance: true });
    });

    expect(receivedBody).toMatchObject({ name: "年次健診", is_non_insurance: true });
    expect(created?.insuranceSetting).toBe(true);
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.category("checkup-types"),
    });
  });

  it("sends the re-edited value on update and refreshes the checkup list", async () => {
    let receivedBody: unknown;
    server.use(
      http.patch("/api/v1/masters/checkup-types/7", async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(makeRawCheckupType({ is_non_insurance: false }));
      }),
    );

    const queryClient = createTestQueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateCheckupType(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ id: "7", req: { is_non_insurance: false, version: 3 } });
    });

    expect(receivedBody).toEqual({ is_non_insurance: false, version: 3 });
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: queryKeys.masters.category("checkup-types"),
    });
  });

  it("omits is_non_insurance on create when unset so the API applies the new-entry default", async () => {
    let receivedBody: unknown;
    server.use(
      http.post("/api/v1/masters/checkup-types", async ({ request }) => {
        receivedBody = await request.json();
        return HttpResponse.json(makeRawCheckupType({ is_non_insurance: true }), { status: 201 });
      }),
    );

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useCreateCheckupType(), {
      wrapper: createQueryWrapper(queryClient),
    });

    let created: Awaited<ReturnType<typeof result.current.mutateAsync>> | undefined;
    await act(async () => {
      created = await result.current.mutateAsync({ name: "年次健診" });
    });

    expect(receivedBody).toEqual({ name: "年次健診" });
    expect(created?.insuranceSetting).toBe(true);
  });

  it("surfaces a 409 version conflict as a failed mutation without refreshing as success", async () => {
    server.use(
      http.patch("/api/v1/masters/checkup-types/7", () =>
        HttpResponse.json({ code: "CONFLICT", message: "stale" }, { status: 409 }),
      ),
    );

    const queryClient = createTestQueryClient();
    const invalidate = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateCheckupType(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({ id: "7", req: { is_non_insurance: true, version: 1 } }),
      ).rejects.toBeTruthy();
    });

    expect(invalidate).not.toHaveBeenCalled();
  });

  it("keeps delete and reorder on the fixed endpoints", async () => {
    let reorderBody: unknown;
    server.use(
      http.delete("/api/v1/masters/checkup-types/7", () => new HttpResponse(null, { status: 204 })),
      http.patch("/api/v1/masters/checkup-types/reorder", async ({ request }) => {
        reorderBody = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );

    const queryClient = createTestQueryClient();
    const deleteHook = renderHook(() => useDeleteCheckupType(), {
      wrapper: createQueryWrapper(queryClient),
    });
    const reorderHook = renderHook(() => useReorderCheckupTypes(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => {
      await deleteHook.result.current.mutateAsync("7");
      await reorderHook.result.current.mutateAsync({ ids: [2, 1] });
    });

    expect(reorderBody).toMatchObject({ ids: [2, 1] });
  });
});
