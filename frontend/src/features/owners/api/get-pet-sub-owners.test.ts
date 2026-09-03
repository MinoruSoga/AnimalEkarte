import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import {
  getPetSubOwnerMetadata,
  getPetSubOwners,
  getSubOwnerCandidates,
  useGetPetSubOwnerMetadata,
  useGetPetSubOwners,
  useGetSubOwnerCandidates,
} from "./get-pet-sub-owners";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

const mockedGet = vi.mocked(axios.get);

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

describe("pet sub-owner queries", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("GET /v1/pets/:id/sub-owners の応答DTOをそのまま返す", async () => {
    const response = {
      sub_owners: [
        {
          owner_id: 12,
          name: "山田 花子",
          name_kana: "ヤマダ ハナコ",
          relationship: "妻",
        },
      ],
    };
    mockedGet.mockResolvedValueOnce({ data: response });

    await expect(getPetSubOwners("7")).resolves.toEqual(response);
    expect(mockedGet).toHaveBeenCalledWith("/v1/pets/7/sub-owners");
  });

  it("ペットの主飼主IDとversionを専用DTOで取得する", async () => {
    mockedGet.mockResolvedValueOnce({
      data: { owner_id: 10, version: 3, name: "ポチ" },
    });

    await expect(getPetSubOwnerMetadata("7")).resolves.toEqual({
      owner_id: 10,
      version: 3,
    });
    expect(mockedGet).toHaveBeenCalledWith("/v1/pets/7");
  });

  it.each([
    { input: "  山田  ", expectedSearch: "山田" },
    { input: "鈴木", expectedSearch: "鈴木" },
  ])(
    "検索語 $input を使う1リクエストで飼主候補を最小DTOへ変換する",
    async ({ input, expectedSearch }) => {
      mockedGet.mockResolvedValueOnce({
        data: {
          data: [
            {
              id: 10,
              owner_name: "山田 太郎",
              owner_name_kana: "ヤマダ タロウ",
            },
          ],
          total: 1,
          page: 1,
          limit: 20,
        },
      });

      await expect(getSubOwnerCandidates(input)).resolves.toEqual([
        { ownerId: 10, name: "山田 太郎", nameKana: "ヤマダ タロウ" },
      ]);
      expect(mockedGet).toHaveBeenCalledTimes(1);
      expect(mockedGet).toHaveBeenCalledWith("/v1/owners", {
        params: { search: expectedSearch },
      });
    },
  );

  it.each(["", "   "])("空検索語 %j では飼主候補を取得しない", async (search) => {
    await expect(getSubOwnerCandidates(search)).resolves.toEqual([]);
    expect(mockedGet).not.toHaveBeenCalled();

    const wrapper = createWrapper();
    const candidates = renderHook(() => useGetSubOwnerCandidates(search), {
      wrapper,
    });

    await waitFor(() => expect(candidates.result.current.fetchStatus).toBe("idle"));
    expect(mockedGet).not.toHaveBeenCalled();
  });

  it("編集権限がない場合は非空検索語でも飼主候補を取得しない", async () => {
    const wrapper = createWrapper();
    const candidates = renderHook(() => useGetSubOwnerCandidates("山田", false), { wrapper });

    await waitFor(() => expect(candidates.result.current.fetchStatus).toBe("idle"));
    expect(mockedGet).not.toHaveBeenCalled();
  });

  it("候補query keyは検索語ごとに分離しenabledを含めない", () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    const wrapper = createWrapper(queryClient);
    const candidates = renderHook(({ search }) => useGetSubOwnerCandidates(search, false), {
      wrapper,
      initialProps: { search: "山田" },
    });

    candidates.rerender({ search: "鈴木" });

    expect(
      queryClient
        .getQueryCache()
        .getAll()
        .map((query) => query.queryKey),
    ).toEqual([
      ["owners", { scope: "sub-owner-options", search: "山田" }],
      ["owners", { scope: "sub-owner-options", search: "鈴木" }],
    ]);
    expect(mockedGet).not.toHaveBeenCalled();
  });

  it("sub-owner hook と metadata hook は別キャッシュで取得する", async () => {
    mockedGet.mockImplementation(async (url) => {
      if (url === "/v1/pets/7/sub-owners") {
        return { data: { sub_owners: [] } };
      }
      return { data: { owner_id: 10, version: 3 } };
    });

    const wrapper = createWrapper();
    const subOwners = renderHook(() => useGetPetSubOwners("7"), { wrapper });
    const metadata = renderHook(() => useGetPetSubOwnerMetadata("7"), { wrapper });

    await waitFor(() => expect(subOwners.result.current.isSuccess).toBe(true));
    await waitFor(() => expect(metadata.result.current.isSuccess).toBe(true));

    expect(subOwners.result.current.data).toEqual({ sub_owners: [] });
    expect(metadata.result.current.data).toEqual({ owner_id: 10, version: 3 });
    expect(mockedGet).toHaveBeenCalledTimes(2);
  });

  it.each(["temp-1", "temp-1710000000000", ""])(
    "BUG-022: 非永続 petId %j では sub-owners / metadata API を発行しない",
    async (petId) => {
      const wrapper = createWrapper();
      const subOwners = renderHook(() => useGetPetSubOwners(petId), { wrapper });
      const metadata = renderHook(() => useGetPetSubOwnerMetadata(petId), {
        wrapper,
      });

      await waitFor(() => expect(subOwners.result.current.fetchStatus).toBe("idle"));
      await waitFor(() => expect(metadata.result.current.fetchStatus).toBe("idle"));
      expect(subOwners.result.current.isFetching).toBe(false);
      expect(metadata.result.current.isFetching).toBe(false);
      expect(mockedGet).not.toHaveBeenCalled();
    },
  );
});
