import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import {
  useCreateCarePlanItem,
  useDeleteCarePlanItem,
  useGetCarePlanItems,
  useUpdateCarePlanItem,
} from "./care-plan-items";
import type {
  CarePlanItem,
  CreateCarePlanItemInput,
  UpdateCarePlanItemInput,
} from "./care-plan-items";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

const mockedGet = vi.mocked(axios.get);
const mockedPost = vi.mocked(axios.post);
const mockedPatch = vi.mocked(axios.patch);
const mockedDelete = vi.mocked(axios.delete);

const responseItem: CarePlanItem = {
  id: "1",
  hospitalization_id: "7",
  type: "instruction",
  name: "ケアプラン項目",
  description: "",
  timing: [],
  status: "active",
  notes: "",
  medicine_id: null,
  procedure_id: null,
  hospitalization_plan_id: null,
  unit_price: 0,
  category: "",
  sort_order: 0,
  created_at: "2026-07-28T00:00:00Z",
  updated_at: "2026-07-28T00:00:00Z",
};

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

function createWrapper(queryClient = createQueryClient()) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe("care plan item API", () => {
  beforeEach(() => {
    mockedGet.mockReset();
    mockedPost.mockReset();
    mockedPatch.mockReset();
    mockedDelete.mockReset();
  });

  it("一覧取得で正しい endpoint の結果を返す", async () => {
    mockedGet.mockResolvedValueOnce({ data: [responseItem] });
    const { result } = renderHook(() => useGetCarePlanItems("7"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockedGet).toHaveBeenCalledTimes(1);
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations/7/care-plan-items");
    expect(result.current.data).toEqual([responseItem]);
  });

  it("hospitalization ID が空なら一覧を取得しない", async () => {
    const { result } = renderHook(() => useGetCarePlanItems(""), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(mockedGet).not.toHaveBeenCalled();
  });

  it("削除後に対象入院のケアプラン一覧を無効化する", async () => {
    mockedDelete.mockResolvedValueOnce({ data: undefined });
    const queryClient = createQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);
    const { result } = renderHook(() => useDeleteCarePlanItem("7"), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync("9");
    });

    expect(mockedDelete).toHaveBeenCalledTimes(1);
    expect(mockedDelete).toHaveBeenCalledWith("/v1/hospitalizations/7/care-plan-items/9");
    expect(invalidateSpy).toHaveBeenCalledTimes(1);
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.hospitalizations.carePlanItems("7"),
    });
  });

  const createCases: Array<{
    name: string;
    input: CreateCarePlanItemInput;
    expectedBody: object;
  }> = [
    {
      name: "投薬の medicine_id を数値へ変換する",
      input: {
        type: "medicine",
        name: "内服薬",
        medicine_id: "101",
        procedure_id: null,
        hospitalization_plan_id: null,
      },
      expectedBody: {
        type: "medicine",
        name: "内服薬",
        medicine_id: 101,
        procedure_id: null,
        hospitalization_plan_id: null,
      },
    },
    {
      name: "処置・検査の procedure_id を数値へ変換する",
      input: {
        type: "treatment",
        name: "血液検査",
        medicine_id: null,
        procedure_id: "202",
        hospitalization_plan_id: null,
      },
      expectedBody: {
        type: "treatment",
        name: "血液検査",
        medicine_id: null,
        procedure_id: 202,
        hospitalization_plan_id: null,
      },
    },
    {
      name: "持ち物の hospitalization_plan_id を数値へ変換する",
      input: {
        type: "item",
        name: "持参フード",
        medicine_id: null,
        procedure_id: null,
        hospitalization_plan_id: "303",
      },
      expectedBody: {
        type: "item",
        name: "持参フード",
        medicine_id: null,
        procedure_id: null,
        hospitalization_plan_id: 303,
      },
    },
  ];

  it.each(createCases)("$name", async ({ input, expectedBody }) => {
    mockedPost.mockResolvedValueOnce({ data: responseItem });
    const { result } = renderHook(() => useCreateCarePlanItem("7"), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await result.current.mutateAsync(input);
    });

    expect(mockedPost).toHaveBeenCalledTimes(1);
    expect(mockedPost).toHaveBeenCalledWith("/v1/hospitalizations/7/care-plan-items", expectedBody);
  });

  const updateCases: Array<{
    name: string;
    input: UpdateCarePlanItemInput;
    expectedBody: object;
  }> = [
    {
      name: "更新時も medicine_id を数値へ変換する",
      input: {
        type: "medicine",
        medicine_id: "404",
        procedure_id: null,
        hospitalization_plan_id: null,
      },
      expectedBody: {
        type: "medicine",
        medicine_id: 404,
        procedure_id: null,
        hospitalization_plan_id: null,
      },
    },
    {
      name: "更新時も procedure_id を数値へ変換する",
      input: {
        type: "treatment",
        medicine_id: null,
        procedure_id: "505",
        hospitalization_plan_id: null,
      },
      expectedBody: {
        type: "treatment",
        medicine_id: null,
        procedure_id: 505,
        hospitalization_plan_id: null,
      },
    },
    {
      name: "更新時も hospitalization_plan_id を数値へ変換する",
      input: {
        type: "item",
        medicine_id: null,
        procedure_id: null,
        hospitalization_plan_id: "606",
      },
      expectedBody: {
        type: "item",
        medicine_id: null,
        procedure_id: null,
        hospitalization_plan_id: 606,
      },
    },
  ];

  it.each(updateCases)("$name", async ({ input, expectedBody }) => {
    mockedPatch.mockResolvedValueOnce({ data: responseItem });
    const { result } = renderHook(() => useUpdateCarePlanItem("7"), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await result.current.mutateAsync({
        itemId: "9",
        input,
      });
    });

    expect(mockedPatch).toHaveBeenCalledTimes(1);
    expect(mockedPatch).toHaveBeenCalledWith(
      "/v1/hospitalizations/7/care-plan-items/9",
      expectedBody,
    );
  });

  it("null は null のまま送る", async () => {
    mockedPost.mockResolvedValueOnce({ data: responseItem });
    const { result } = renderHook(() => useCreateCarePlanItem("7"), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await result.current.mutateAsync({
        type: "instruction",
        name: "安静",
        medicine_id: null,
        procedure_id: null,
        hospitalization_plan_id: null,
      });
    });

    expect(mockedPost).toHaveBeenCalledWith("/v1/hospitalizations/7/care-plan-items", {
      type: "instruction",
      name: "安静",
      medicine_id: null,
      procedure_id: null,
      hospitalization_plan_id: null,
    });
  });

  it("省略・undefined・空文字を 0 に変換しない", async () => {
    mockedPost.mockResolvedValueOnce({ data: responseItem });
    const { result } = renderHook(() => useCreateCarePlanItem("7"), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await result.current.mutateAsync({
        type: "instruction",
        name: "経過観察",
        medicine_id: undefined,
        procedure_id: "",
      });
    });

    expect(mockedPost).toHaveBeenCalledTimes(1);
    const body = mockedPost.mock.calls[0]?.[1];
    expect(body).not.toHaveProperty("medicine_id");
    expect(body).not.toHaveProperty("procedure_id");
    expect(body).not.toHaveProperty("hospitalization_plan_id");
  });

  it.each([
    ["空白", " "],
    ["非数値", "abc"],
    ["小数", "1.5"],
    ["ゼロ", "0"],
    ["負数", "-1"],
    ["safe integer 上限超過", String(Number.MAX_SAFE_INTEGER + 1)],
  ])("%sの参照 ID は送信しない", async (_name, medicineId) => {
    const { result } = renderHook(() => useCreateCarePlanItem("7"), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({
          type: "medicine",
          name: "不正な参照 ID",
          medicine_id: medicineId,
        }),
      ).rejects.toThrow("Care plan reference ID must be a positive safe integer");
    });

    expect(mockedPost).not.toHaveBeenCalled();
  });
});
