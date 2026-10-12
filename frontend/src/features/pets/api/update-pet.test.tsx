import type { ReactNode } from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import { useUpdatePet } from "./update-pet";

vi.mock("@/lib/axios", () => ({
  axios: { patch: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

const mockedPatch = vi.mocked(axios.patch);
const mockedHandleApiError = vi.mocked(handleApiError);

function createWrapper(queryClient: QueryClient) {
  return function QueryClientWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

const invalidatedKeysOf = (spy: ReturnType<typeof vi.fn>) =>
  spy.mock.calls.map(([arg]) => (arg as { queryKey: readonly unknown[] }).queryKey);

describe("useUpdatePet", () => {
  beforeEach(() => {
    mockedPatch.mockReset();
    mockedHandleApiError.mockReset();
    mockedPatch.mockResolvedValue({
      data: { id: 1, owner_id: 20, name: "ポチ" },
    });
  });

  const setup = () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        mutations: { retry: false },
        queries: { retry: false },
      },
    });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);
    const { result } = renderHook(() => useUpdatePet(), {
      wrapper: createWrapper(queryClient),
    });
    return { result, invalidateSpy };
  };

  it("通常更新(owner_id なし)は pets のみ無効化し予約/受付キーを広く invalidate しない", async () => {
    const { result, invalidateSpy } = setup();

    await act(async () => {
      await result.current.mutateAsync({ id: "1", req: { name: "ポチ改" } });
    });

    expect(invalidateSpy).toHaveBeenCalledTimes(2);
    expect(invalidatedKeysOf(invalidateSpy)).toEqual([
      queryKeys.pets.list(),
      queryKeys.pets.detail("1"),
    ]);
    expect(mockedHandleApiError).not.toHaveBeenCalled();
  });

  it("飼主変更(owner_id 指定)は両飼主の一覧・受付・予約キャッシュも無効化する", async () => {
    const { result, invalidateSpy } = setup();

    await act(async () => {
      await result.current.mutateAsync({ id: "1", req: { owner_id: 30 } });
    });

    expect(mockedPatch).toHaveBeenCalledWith("/v1/pets/1", { owner_id: 30 });
    expect(invalidateSpy).toHaveBeenCalledTimes(8);
    expect(invalidatedKeysOf(invalidateSpy)).toEqual([
      queryKeys.pets.list(),
      queryKeys.pets.detail("1"),
      queryKeys.reservations.all(),
      queryKeys.reservations.detailAll(),
      queryKeys.reception.all(),
      queryKeys.owners.all(),
      queryKeys.ownerReportPetsAll(),
      queryKeys.ownerSharedPets.all(),
    ]);
    expect(mockedHandleApiError).not.toHaveBeenCalled();
  });
});
