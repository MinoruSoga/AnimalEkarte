import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { replacePetSubOwners, useReplacePetSubOwners } from "./replace-pet-sub-owners";

vi.mock("@/lib/axios", () => ({
  axios: {
    put: vi.fn(),
  },
}));

const mockedPut = vi.mocked(axios.put);

describe("replace pet sub-owners", () => {
  beforeEach(() => {
    mockedPut.mockReset();
  });

  it("version と sub_owners を含む body で PUT を1回呼ぶ", async () => {
    mockedPut.mockResolvedValueOnce({ data: undefined });
    const request = {
      version: 3,
      sub_owners: [{ owner_id: 12, relationship: "妻" }],
    };

    await expect(replacePetSubOwners("7", request)).resolves.toBeUndefined();
    expect(mockedPut).toHaveBeenCalledTimes(1);
    expect(mockedPut).toHaveBeenCalledWith("/v1/pets/7/sub-owners", request);
  });

  it("空配列を含む全解除 request をそのまま送る", async () => {
    mockedPut.mockResolvedValueOnce({ data: undefined });

    await replacePetSubOwners("7", { version: 4, sub_owners: [] });

    expect(mockedPut).toHaveBeenCalledWith("/v1/pets/7/sub-owners", {
      version: 4,
      sub_owners: [],
    });
  });

  it("成功後に pet・副飼主一覧・metadata の query key を invalidate する", async () => {
    mockedPut.mockResolvedValueOnce({ data: undefined });
    const queryClient = new QueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client: queryClient }, children);
    const { result } = renderHook(() => useReplacePetSubOwners(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        petId: "7",
        request: { version: 3, sub_owners: [] },
      });
    });

    expect(invalidateSpy).toHaveBeenCalledTimes(3);
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.pets.detail("7"),
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.petSubOwners.detail("7"),
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.petSubOwners.metadata("7"),
    });
  });
});
