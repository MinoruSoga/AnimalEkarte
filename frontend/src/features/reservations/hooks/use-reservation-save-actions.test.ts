import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createElement, type MutableRefObject, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";

import { setStoredClinicId } from "@/lib/current-clinic";
import type { ReservationFormData } from "../types";
import { useReservationSaveActions } from "./use-reservation-save-actions";

const receptionErrorHandlerMock = vi.fn();
vi.mock("@/lib/handle-api-error", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/lib/handle-api-error")>()),
  handleApiError: (...args: unknown[]) => receptionErrorHandlerMock(...args),
}));
const receptionCandidatesMock = vi.fn();
vi.mock("../api/get-reception-candidates", () => ({
  getReceptionCandidates: (...args: unknown[]) => receptionCandidatesMock(...args),
}));

const createMutateAsyncMock = vi.fn();
const createBatchMutateAsyncMock = vi.fn();
const updateMutateMock = vi.fn();

vi.mock("../api/create-reservation", () => ({
  useCreateReservation: () => ({ mutateAsync: createMutateAsyncMock }),
  useCreateReservationBatch: () => ({ mutateAsync: createBatchMutateAsyncMock }),
}));

vi.mock("../api/update-reservation", () => ({
  useUpdateReservation: () => ({ mutate: updateMutateMock }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const CONFLICT_MESSAGE = "既に予約が存在します";

function axiosError(status: number, data: Record<string, unknown>): AxiosError {
  const config = {
    headers: new AxiosHeaders(),
  } as InternalAxiosRequestConfig;
  return new AxiosError("request failed", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
    config,
    data,
    headers: new AxiosHeaders(),
    status,
    statusText: "Conflict",
  });
}

function conflict409(): AxiosError {
  return axiosError(409, { error: CONFLICT_MESSAGE, code: "reservation_time_conflict" });
}

function makeFormData(): ReservationFormData {
  return {
    start: new Date("2026-05-29T03:30:00.000Z"),
    end: new Date("2026-05-29T04:00:00.000Z"),
    visitType: "first",
    type: "1",
    doctor: "1",
    status: "confirmed",
  } as ReservationFormData;
}

const ALL_ALLOWED = { canCreate: true, canEdit: true, canDelete: true } as const;

function setup(
  checkOverlap = vi.fn(() => false),
  permissions = { canCreate: true, canEdit: true, canDelete: true },
) {
  const editingAppointmentRef: MutableRefObject<ReservationFormData | null> = {
    current: null,
  };
  const handleCloseForm = vi.fn();
  const navigateBackIfNeeded = vi.fn();
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const { result, rerender } = renderHook(
    () =>
      useReservationSaveActions({
        editingAppointmentRef,
        checkOverlap,
        handleCloseForm,
        navigateBackIfNeeded,
        createMutations: {
          createOwnerFn: vi.fn(),
          createPetFn: vi.fn(),
        },
        permissions,
      }),
    {
      wrapper: ({ children }: { children: ReactNode }) =>
        createElement(QueryClientProvider, { client: queryClient }, children),
    },
  );
  return { result, rerender, checkOverlap, handleCloseForm, navigateBackIfNeeded };
}

describe("useReservationSaveActions 409 conflict message (EMR-76)", () => {
  beforeEach(() => {
    createMutateAsyncMock.mockReset();
    createBatchMutateAsyncMock.mockReset();
    updateMutateMock.mockReset();
  });

  it("single-pet create returns the backend 既に予約が存在します on a 409 reservation_time_conflict", async () => {
    createMutateAsyncMock.mockRejectedValue(conflict409());
    const { result } = setup();
    await act(async () => {
      const message = await result.current.handleSave(makeFormData(), [
        { id: "10", ownerId: "20", name: "ポチ" },
      ]);
      expect(message).toBe(CONFLICT_MESSAGE);
    });
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(1);
  });

  it("batch create returns the same message and a retry still surfaces it (not a 500-style fallback)", async () => {
    createBatchMutateAsyncMock.mockRejectedValue(conflict409());
    const { result } = setup();
    const pets = [
      { id: "10", ownerId: "20", name: "ポチ" },
      { id: "11", ownerId: "20", name: "タマ" },
    ];
    await act(async () => {
      expect(await result.current.handleSave(makeFormData(), pets)).toBe(CONFLICT_MESSAGE);
    });
    // retry after a concurrent booking attempt surfaces the same conflict message
    await act(async () => {
      expect(await result.current.handleSave(makeFormData(), pets)).toBe(CONFLICT_MESSAGE);
    });
    expect(createBatchMutateAsyncMock).toHaveBeenCalledTimes(2);
  });

  it("update path resolves the backend conflict message instead of a generic update failure", async () => {
    updateMutateMock.mockImplementation((_payload, callbacks) => {
      callbacks.onError(conflict409());
    });
    const editingAppointmentRef: MutableRefObject<ReservationFormData | null> = {
      current: { ...makeFormData(), id: "r1" } as ReservationFormData,
    };
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { result } = renderHook(
      () =>
        useReservationSaveActions({
          editingAppointmentRef,
          checkOverlap: vi.fn(() => false),
          handleCloseForm: vi.fn(),
          navigateBackIfNeeded: vi.fn(),
          createMutations: { createOwnerFn: vi.fn(), createPetFn: vi.fn() },
          permissions: ALL_ALLOWED,
        }),
      {
        wrapper: ({ children }: { children: ReactNode }) =>
          createElement(QueryClientProvider, { client: queryClient }, children),
      },
    );
    await act(async () => {
      const message = await result.current.handleSave(makeFormData(), [
        { id: "10", ownerId: "20", name: "ポチ" },
      ]);
      expect(message).toBe(CONFLICT_MESSAGE);
    });
    expect(updateMutateMock).toHaveBeenCalledTimes(1);
  });

  it("FE overlap precheck returns the same 既に予約が存在します message as the API 409", async () => {
    const checkOverlap = vi.fn(() => true);
    const { result } = setup(checkOverlap);
    await act(async () => {
      const message = await result.current.handleSave(makeFormData(), [
        { id: "10", ownerId: "20", name: "ポチ" },
      ]);
      expect(message).toBe(CONFLICT_MESSAGE);
    });
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
});

describe("EMR-318 reception reuse", () => {
  beforeEach(() => {
    setStoredClinicId("1");
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-05-29T01:00:00Z"));
    receptionErrorHandlerMock.mockReset();
    receptionCandidatesMock.mockReset();
    createMutateAsyncMock.mockReset();
    updateMutateMock.mockReset();
  });
  afterEach(() => vi.useRealTimers());
  const reception = () => ({
    ...makeFormData(),
    reservationRoute: "reception" as const,
    status: "checked_in" as const,
  });
  const pets = [{ id: "10", ownerId: "20", name: "ポチ" }];
  const candidate = (id = 5, status = "confirmed") => ({
    id,
    status,
    clinic_id: 1,
    pet_id: 10,
    owner_id: 20,
    start_time: "2026-05-29T01:00:00Z",
    updated_at: "2026-05-29T00:00:00.123456789Z",
  });
  it("reuses confirmed reservation with status only and no create", async () => {
    receptionCandidatesMock.mockResolvedValue([candidate()]);
    updateMutateMock.mockImplementation((_payload, callbacks) => callbacks.onSuccess());
    const { result } = setup();
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toBeNull();
    });
    expect(updateMutateMock).toHaveBeenCalledWith(
      {
        id: "5",
        req: { status: "checked_in", expected_updated_at: "2026-05-29T00:00:00.123456789Z" },
      },
      expect.any(Object),
    );
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("does not repeat check-in or change its date for existing checked-in visit", async () => {
    receptionCandidatesMock.mockResolvedValue([
      { ...candidate(5, "checked_in"), checked_in_at: "2026-05-29T00:00:00Z" },
    ]);
    const { result } = setup();
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toBeNull();
    });
    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("blocks create when candidate lookup fails or multiple candidates exist", async () => {
    const { result } = setup();
    receptionCandidatesMock.mockRejectedValue(new Error("fetch failed"));
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toBeTruthy();
    });
    receptionCandidatesMock.mockResolvedValue([candidate(), candidate(6)]);
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toContain("複数");
    });
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("does not reuse a reservation without edit permission", async () => {
    receptionCandidatesMock.mockResolvedValue([candidate()]);
    const { result } = setup(undefined, { canCreate: true, canEdit: false, canDelete: false });
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toContain("権限");
    });
    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("blocks mutation if permission changes during lookup", async () => {
    const permissions = { canCreate: true, canEdit: true, canDelete: true };
    let resolveLookup: ((value: []) => void) | undefined;
    receptionCandidatesMock.mockImplementation(
      () =>
        new Promise<[]>((resolve) => {
          resolveLookup = resolve;
        }),
    );
    const { result, rerender } = setup(undefined, permissions);
    let pending: Promise<string | null> | undefined;
    act(() => {
      pending = result.current.handleSave(reception(), pets);
    });
    permissions.canCreate = false;
    act(() => rerender());
    await act(async () => {
      resolveLookup?.([]);
      expect(await pending).toContain("権限");
    });
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("handles missing selected clinic before lookup once", async () => {
    localStorage.removeItem("auth_current_clinic:v1");
    const { result } = setup();
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toBeTruthy();
    });
    expect(receptionErrorHandlerMock).toHaveBeenCalledTimes(1);
    expect(receptionCandidatesMock).not.toHaveBeenCalled();
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("blocks mutation if selected clinic changes during lookup", async () => {
    receptionCandidatesMock.mockImplementation(async () => {
      setStoredClinicId("2");
      return [];
    });
    const { result } = setup();
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toBeTruthy();
    });
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
    expect(updateMutateMock).not.toHaveBeenCalled();
  });
  it("blocks mutation when JST date rolls over during lookup", async () => {
    receptionCandidatesMock.mockImplementation(async () => {
      vi.setSystemTime(new Date("2026-05-29T15:00:00Z"));
      return [];
    });
    const { result } = setup();
    await act(async () => {
      expect(await result.current.handleSave(reception(), pets)).toContain("受付日");
    });
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
  });
  it("prevents a second in-flight save", async () => {
    receptionCandidatesMock.mockResolvedValue([]);
    let finish: (() => void) | undefined;
    createMutateAsyncMock.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const { result } = setup();
    await act(async () => {
      const first = result.current.handleSave(reception(), pets);
      for (let i = 0; i < 10; i += 1) await Promise.resolve();
      expect(await result.current.handleSave(reception(), pets)).toContain("処理中");
      finish?.();
      await first;
    });
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(1);
  });
});
