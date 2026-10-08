import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { queryKeys } from "@/lib/query-keys";
import { CURRENT_CLINIC_STORAGE_KEY } from "@/lib/current-clinic";
import { server } from "@/testing/mocks/node";

import { useRecordPetDeath } from "./use-record-pet-death";
import { useRevokePetDeath } from "./use-revoke-pet-death";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

// PR#614 指摘: 死亡登録/取消は wire ペットサマリの死亡 sentinel（petStatus/petDeceasedAt）を
// 埋め込む全キャッシュを無効化する。pets.* だけだと staleTime(REALTIME=2分)内の stale 応答が
// 誕生月バッジ・カルテ作成導線を残す。
const SENTINEL_BEARING_KEYS = [
  queryKeys.reservations.list({}),
  queryKeys.reservations.detail("r1"),
  queryKeys.reception.byDate("2026-10-15"),
  queryKeys.hospitalizations.list({}),
  queryKeys.hospitalizations.detail("h1"),
  queryKeys.ownerSharedPets.detail("o1"),
  queryKeys.ownerReportPets("o1"),
] as const;

function setup() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  for (const key of SENTINEL_BEARING_KEYS) {
    queryClient.setQueryData(key, { seeded: true });
  }
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, wrapper };
}

const invalidated = (qc: QueryClient, key: readonly unknown[]) =>
  qc.getQueryState(key)?.isInvalidated === true;

describe("useRecordPetDeath / useRevokePetDeath — sentinel 埋め込みキャッシュの無効化", () => {
  beforeEach(() => {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "1");
  });

  afterEach(() => {
    localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
  });

  it("死亡登録成功で reservations/reception/hospitalizations/shared-pets 系を無効化する", async () => {
    server.use(http.patch("/api/v1/clinics/1/pets/p1/death", () => HttpResponse.json({})));
    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useRecordPetDeath(), { wrapper });

    await result.current.mutateAsync({ petId: "p1", deceasedAt: "2026-10-15" });

    await waitFor(() => {
      for (const key of SENTINEL_BEARING_KEYS) {
        expect(invalidated(queryClient, key), `key=${JSON.stringify(key)}`).toBe(true);
      }
    });
  });

  it("死亡解除成功でも同じ sentinel 埋め込みキャッシュを無効化する", async () => {
    server.use(http.delete("/api/v1/clinics/1/pets/p1/death", () => HttpResponse.json({})));
    const { queryClient, wrapper } = setup();
    const { result } = renderHook(() => useRevokePetDeath(), { wrapper });

    await result.current.mutateAsync("p1");

    await waitFor(() => {
      for (const key of SENTINEL_BEARING_KEYS) {
        expect(invalidated(queryClient, key), `key=${JSON.stringify(key)}`).toBe(true);
      }
    });
  });

  it("pets.detail/list は従来どおり無効化する", async () => {
    server.use(http.patch("/api/v1/clinics/1/pets/p1/death", () => HttpResponse.json({})));
    const { queryClient, wrapper } = setup();
    queryClient.setQueryData(queryKeys.pets.detail("p1"), { seeded: true });
    queryClient.setQueryData(queryKeys.pets.list(), { seeded: true });
    const { result } = renderHook(() => useRecordPetDeath(), { wrapper });

    await result.current.mutateAsync({ petId: "p1", deceasedAt: "2026-10-15" });

    await waitFor(() => {
      expect(invalidated(queryClient, queryKeys.pets.detail("p1"))).toBe(true);
      expect(invalidated(queryClient, queryKeys.pets.list())).toBe(true);
    });
  });
});
