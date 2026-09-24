import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { MutableRefObject } from "react";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";

import type { ReservationFormData } from "../types";
import { useReservationSaveActions } from "./use-reservation-save-actions";

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

function setup(checkOverlap = vi.fn(() => false)) {
  const editingAppointmentRef: MutableRefObject<ReservationFormData | null> = {
    current: null,
  };
  const handleCloseForm = vi.fn();
  const navigateBackIfNeeded = vi.fn();
  const { result } = renderHook(() =>
    useReservationSaveActions({
      editingAppointmentRef,
      checkOverlap,
      handleCloseForm,
      navigateBackIfNeeded,
      createMutations: {
        createOwnerFn: vi.fn(),
        createPetFn: vi.fn(),
      },
      permissions: ALL_ALLOWED,
    }),
  );
  return { result, checkOverlap, handleCloseForm, navigateBackIfNeeded };
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
    const { result } = renderHook(() =>
      useReservationSaveActions({
        editingAppointmentRef,
        checkOverlap: vi.fn(() => false),
        handleCloseForm: vi.fn(),
        navigateBackIfNeeded: vi.fn(),
        createMutations: { createOwnerFn: vi.fn(), createPetFn: vi.fn() },
        permissions: ALL_ALLOWED,
      }),
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
