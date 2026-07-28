import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";
import {
  getOwnerSharedPets,
  useGetOwnerSharedPets,
} from "./get-owner-shared-pets";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

const mockedGet = vi.mocked(axios.get);

const RAW_RESPONSE = {
  shared_pets: [
    {
      id: 42,
      pet_number: "P-0042",
      name: "ハナ",
      status: "alive",
      gender: "female",
      animal_species: { name: "犬" },
      birth_date: "2021-02-03",
      color: "白",
      weight: 8.5,
      environment: "室内",
      remarks: "共有ペット",
      relationship: "妻",
    },
  ],
};

function createWrapper(
  queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  }),
) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("owner shared-pet query", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("GET /v1/owners/:id/shared-pets のraw DTOを1リクエストでそのまま返す", async () => {
    mockedGet.mockResolvedValueOnce({ data: RAW_RESPONSE });

    await expect(getOwnerSharedPets("7")).resolves.toEqual(RAW_RESPONSE);
    expect(mockedGet).toHaveBeenCalledTimes(1);
    expect(mockedGet).toHaveBeenCalledWith("/v1/owners/7/shared-pets");
  });

  it("非空owner IDでは専用query keyと静的キャッシュ設定で1回取得する", async () => {
    mockedGet.mockResolvedValueOnce({ data: RAW_RESPONSE });
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    const { result } = renderHook(() => useGetOwnerSharedPets("7"), {
      wrapper: createWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(RAW_RESPONSE);
    expect(mockedGet).toHaveBeenCalledTimes(1);
    const query = queryClient.getQueryCache().find({
      exact: true,
      queryKey: queryKeys.ownerSharedPets.detail("7"),
    });
    expect(query?.options.staleTime).toBe(QUERY_STALE_TIMES.STATIC);
    expect(query?.options.gcTime).toBe(QUERY_GC_TIMES.LONG);
  });

  it.each([undefined, ""] as const)(
    "owner IDが%jならリクエストしない",
    async (ownerId) => {
      const { result } = renderHook(() => useGetOwnerSharedPets(ownerId), {
        wrapper: createWrapper(),
      });

      await waitFor(() => expect(result.current.fetchStatus).toBe("idle"));
      expect(mockedGet).not.toHaveBeenCalled();
    },
  );
});
