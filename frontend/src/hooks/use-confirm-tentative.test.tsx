import type { ReactNode } from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import { useConfirmTentative } from "./use-confirm-tentative";

/**
 * EMR-308: POST /v1/reservations/:id/confirm-tentative — tentative 行を同一行のまま
 * 実予約へ昇格する共有 mutation。URL・body 変換・invalidate 範囲を固定する。
 */

vi.mock("@/lib/axios", () => ({
  axios: { post: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

const mockedPost = vi.mocked(axios.post);
const mockedHandleApiError = vi.mocked(handleApiError);

function createWrapper(queryClient: QueryClient) {
  return function QueryClientWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

const BACKEND_RESERVATION = {
  id: 42,
  start_time: "2026-10-05T01:30:00Z",
  end_time: "2026-10-05T02:30:00Z",
  status: "confirmed",
  intended_date: "2026-10-05",
  had_intended_date: true,
};

describe("useConfirmTentative (EMR-308)", () => {
  beforeEach(() => {
    mockedPost.mockReset();
    mockedHandleApiError.mockReset();
    mockedPost.mockResolvedValue({ data: BACKEND_RESERVATION });
  });

  it("POST /v1/reservations/:id/confirm-tentative に formData 由来の body を送る", async () => {
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
    });
    const { result } = renderHook(() => useConfirmTentative(), {
      wrapper: createWrapper(queryClient),
    });

    let reservation: unknown;
    await act(async () => {
      reservation = await result.current.mutateAsync({
        reservationId: "42",
        formData: {
          start: new Date(2026, 9, 5, 10, 30),
          end: new Date(2026, 9, 5, 11, 30),
          type: "7",
          doctor: "33",
          isDesignated: true,
          status: "confirmed",
          notes: "次回は検査",
        },
      });
    });

    expect(mockedPost).toHaveBeenCalledTimes(1);
    const [url, body] = mockedPost.mock.calls[0];
    expect(url).toBe("/v1/reservations/42/confirm-tentative");
    expect(body).toMatchObject({
      reservation_type_id: 7,
      doctor_id: 33,
      is_designated: true,
      status: "confirmed",
      notes: "次回は検査",
    });
    // JST wall-clock Date → ISO（Z形式。10:30 JST = 01:30 UTC）
    expect(body).toMatchObject({
      start_time: "2026-10-05T01:30:00.000Z",
      end_time: "2026-10-05T02:30:00.000Z",
    });
    // 応答は transformReservation 済み（isTentative=false / hadIntendedDate=true）
    expect(reservation).toMatchObject({ id: "42", hadIntendedDate: true });
  });

  it("status は pending/confirmed のみ送信する（他値は confirmed に倒す）", async () => {
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const { result } = renderHook(() => useConfirmTentative(), {
      wrapper: createWrapper(queryClient),
    });
    const form = { start: new Date(2026, 9, 5, 10, 0), end: new Date(2026, 9, 5, 11, 0) };

    await act(async () => {
      await result.current.mutateAsync({
        reservationId: "42",
        formData: { ...form, status: "pending" },
      });
    });
    expect(mockedPost.mock.calls[0][1]).toMatchObject({ status: "pending" });

    await act(async () => {
      await result.current.mutateAsync({
        reservationId: "42",
        formData: { ...form, status: "checked_in" },
      });
    });
    expect(mockedPost.mock.calls[1][1]).toMatchObject({ status: "confirmed" });
  });

  it("成功時に reservations / reception / detail を invalidate する", async () => {
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries").mockResolvedValue(undefined);
    const { result } = renderHook(() => useConfirmTentative(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({
        reservationId: "42",
        formData: { start: new Date(2026, 9, 5, 10, 0), end: new Date(2026, 9, 5, 11, 0) },
      });
    });

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: queryKeys.reservations.all() });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: queryKeys.reception.all() });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: queryKeys.reservations.detail("42"),
    });
  });

  it("start/end が欠けると API を呼ばず fail-closed で失敗する", async () => {
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const { result } = renderHook(() => useConfirmTentative(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({ reservationId: "42", formData: {} }),
      ).rejects.toThrow("confirm-tentative: start/end is required");
    });

    expect(mockedPost).not.toHaveBeenCalled();
  });

  it("API 失敗時は handleApiError へ委譲する", async () => {
    const error = new Error("boom");
    mockedPost.mockRejectedValueOnce(error);
    const queryClient = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const { result } = renderHook(() => useConfirmTentative(), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({
          reservationId: "42",
          formData: { start: new Date(2026, 9, 5, 10, 0), end: new Date(2026, 9, 5, 11, 0) },
        }),
      ).rejects.toThrow("boom");
    });

    expect(mockedHandleApiError).toHaveBeenCalledWith(error, "次回予定の確定");
  });
});
