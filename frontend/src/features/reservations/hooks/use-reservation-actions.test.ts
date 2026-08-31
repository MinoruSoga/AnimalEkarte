import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { MutableRefObject } from "react";

import type { Reservation, ReservationFormData, ReservationStatus } from "../types";
import {
  buildReservationUpdateRequest,
  isDestructiveReservationStatus,
  useReservationActions,
  type StatusConfirmTarget,
} from "./use-reservation-actions";

const updateMutateMock = vi.fn();
const createMutateAsyncMock = vi.fn();

vi.mock("../api/create-reservation", () => ({
  useCreateReservation: () => ({ mutateAsync: createMutateAsyncMock }),
}));

vi.mock("../api/delete-reservation", () => ({
  useDeleteReservation: () => ({ mutate: vi.fn() }),
}));

vi.mock("../api/update-reservation", () => ({
  useUpdateReservation: () => ({ mutate: updateMutateMock }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function makeReservation(overrides: Partial<Reservation> = {}): Reservation {
  return {
    id: "r1",
    start: new Date("2026-05-29T03:30:00.000Z"),
    end: new Date("2026-05-29T04:00:00.000Z"),
    ownerName: "山田",
    petName: "ポチ",
    petType: "犬",
    visitType: "first",
    type: "一般診察",
    category: "general",
    reservationTypeId: "1",
    doctor: "1",
    doctorId: "1",
    isDesignated: false,
    status: "confirmed",
    notes: "memo",
    petId: "10",
    ownerId: "20",
    source: "manual",
    reservationRoute: "reception",
    ...overrides,
  };
}

interface SetupOptions {
  appointments?: Reservation[];
  createOwnerFn?: ReturnType<typeof vi.fn>;
  createPetFn?: ReturnType<typeof vi.fn>;
}

function setup(options: SetupOptions = {}) {
  const editingAppointmentRef: MutableRefObject<ReservationFormData | null> = {
    current: null,
  };
  const setDeleteConfirmOpen = vi.fn();
  const setDeleteTarget = vi.fn();
  const setStatusConfirmOpen = vi.fn();
  const setStatusConfirmTarget = vi.fn();
  const setDetailAppointment = vi.fn();
  let statusConfirmTarget: StatusConfirmTarget | null = null;
  setStatusConfirmTarget.mockImplementation((value) => {
    statusConfirmTarget =
      typeof value === "function" ? value(statusConfirmTarget) : value;
  });

  const handleCloseForm = vi.fn();
  const createOwnerFn = options.createOwnerFn ?? vi.fn();
  const createPetFn = options.createPetFn ?? vi.fn();
  const { result, rerender } = renderHook(
    (props: { statusConfirmTarget: StatusConfirmTarget | null; appointments: Reservation[] }) =>
      useReservationActions({
        appointments: props.appointments,
        editingAppointmentRef,
        deleteTarget: null,
        setDeleteConfirmOpen,
        setDeleteTarget,
        statusConfirmTarget: props.statusConfirmTarget,
        setStatusConfirmOpen,
        setStatusConfirmTarget,
        setDetailAppointment,
        handleCloseForm,
        handleCloseDetail: vi.fn(),
        locationFrom: null,
        navigate: vi.fn(),
        createMutations: {
          createOwnerFn,
          createPetFn,
        },
      }),
    {
      initialProps: {
        statusConfirmTarget: null as StatusConfirmTarget | null,
        appointments: options.appointments ?? [],
      },
    },
  );

  return {
    result,
    rerender,
    setStatusConfirmOpen,
    setStatusConfirmTarget,
    handleCloseForm,
    getStatusConfirmTarget: () => statusConfirmTarget,
  };
}

describe("isDestructiveReservationStatus", () => {
  it.each([
    ["cancelled", true],
    ["no_show", true],
    ["confirmed", false],
    ["checked_in", false],
    ["completed", false],
  ] as const)("%s → %s", (status, expected) => {
    expect(isDestructiveReservationStatus(status)).toBe(expected);
  });
});

describe("useReservationActions handleStatusChange (BUG-020)", () => {
  beforeEach(() => {
    updateMutateMock.mockReset();
  });

  it("非破壊的 status は {status} のみで即 PATCH する", () => {
    const { result } = setup();
    const reservation = makeReservation({ status: "confirmed" });

    act(() => {
      result.current.handleStatusChange(reservation, "checked_in");
    });

    expect(updateMutateMock).toHaveBeenCalledTimes(1);
    expect(updateMutateMock).toHaveBeenCalledWith(
      { id: "r1", req: { status: "checked_in" } },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    const payload = updateMutateMock.mock.calls[0][0];
    expect(payload.req).toEqual({ status: "checked_in" });
    expect(payload.req).not.toHaveProperty("start_time");
    expect(payload.req).not.toHaveProperty("end_time");
    expect(payload.req).not.toHaveProperty("doctor_id");
  });

  it("キャンセルは確認ダイアログを開き、確定後に status-only PATCH する", () => {
    const { result, rerender, setStatusConfirmOpen, getStatusConfirmTarget } = setup();
    const reservation = makeReservation({ status: "confirmed" });
    const next: ReservationStatus = "cancelled";

    act(() => {
      result.current.handleStatusChange(reservation, next);
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(setStatusConfirmOpen).toHaveBeenCalledWith(true);
    const pending = getStatusConfirmTarget();
    expect(pending).toEqual({ reservation, status: next });

    rerender({ statusConfirmTarget: pending, appointments: [] });

    act(() => {
      result.current.executeStatusChange();
    });

    expect(updateMutateMock).toHaveBeenCalledTimes(1);
    expect(updateMutateMock).toHaveBeenCalledWith(
      { id: "r1", req: { status: "cancelled" } },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    const payload = updateMutateMock.mock.calls[0][0];
    expect(Object.keys(payload.req)).toEqual(["status"]);
  });

  it("同一 status では API を呼ばない", () => {
    const { result } = setup();
    const reservation = makeReservation({ status: "confirmed" });

    act(() => {
      result.current.handleStatusChange(reservation, "confirmed");
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
  });

  it("受付済→診療中はカルテ作成必須のため PATCH しない", () => {
    const { result } = setup();
    const reservation = makeReservation({ status: "checked_in" });

    act(() => {
      result.current.handleStatusChange(reservation, "in_consultation");
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
  });
});

describe("buildReservationUpdateRequest (BUG-012)", () => {
  const start = new Date("2026-05-29T03:30:00.000Z");
  const end = new Date("2026-05-29T04:00:00.000Z");
  const current: ReservationFormData = {
    id: "r1",
    start,
    end,
    visitType: "first",
    type: "1",
    doctor: "1",
    isDesignated: false,
    status: "checked_in",
    notes: "memo",
  };

  it("メモのみ変更なら notes だけを送る", () => {
    const payload = buildReservationUpdateRequest(
      current,
      { ...current, notes: "updated memo" },
      "1",
    );
    expect(payload).toEqual({
      id: "r1",
      req: { notes: "updated memo" },
    });
  });

  it("時刻変更があるときだけ start_time/end_time を含める", () => {
    const nextEnd = new Date("2026-05-29T04:30:00.000Z");
    const payload = buildReservationUpdateRequest(
      current,
      { ...current, end: nextEnd },
      "1",
    );
    expect(payload?.req).toHaveProperty("end_time");
    expect(payload?.req).not.toHaveProperty("doctor_id");
    expect(payload?.req).not.toHaveProperty("start_time");
  });
});


describe("useReservationActions multi-pet retry", () => {
  beforeEach(() => {
    createMutateAsyncMock.mockReset();
  });

  it("retries only pets that failed after a partial create", async () => {
    createMutateAsyncMock
      .mockResolvedValueOnce(makeReservation({ id: "created-10" }))
      .mockRejectedValueOnce(new Error("temporary failure"))
      .mockResolvedValueOnce(makeReservation({ id: "created-11" }));
    const { result, rerender } = setup();
    const data: ReservationFormData = {
      start: new Date("2026-05-29T03:30:00.000Z"),
      end: new Date("2026-05-29T04:00:00.000Z"),
      visitType: "first",
      type: "1",
      doctor: "1",
      status: "confirmed",
    };
    const pets = [
      { id: "10", ownerId: "20", name: "ポチ" },
      { id: "11", ownerId: "20", name: "タマ" },
    ];

    let firstResult: string | null = null;
    await act(async () => {
      firstResult = await result.current.handleSave(data, pets);
    });
    expect(firstResult).toBeTruthy();
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(2);

    rerender({
      statusConfirmTarget: null,
      appointments: [makeReservation({ id: "created-10", petId: "10" })],
    });
    await act(async () => {
      await result.current.handleSave({ ...data, notes: "修正後" }, pets);
    });
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(3);
    expect(createMutateAsyncMock.mock.calls[2]?.[0]).toEqual(
      expect.objectContaining({ pet_id: 11 }),
    );
  });

  it("clears partial-create progress when the create modal is closed", async () => {
    createMutateAsyncMock
      .mockResolvedValueOnce(makeReservation({ id: "created-10" }))
      .mockRejectedValueOnce(new Error("temporary failure"))
      .mockResolvedValue(makeReservation({ id: "retry" }));
    const { result } = setup();
    const data: ReservationFormData = {
      start: new Date("2026-05-29T03:30:00.000Z"),
      end: new Date("2026-05-29T04:00:00.000Z"),
      visitType: "first",
      type: "1",
      doctor: "1",
      status: "confirmed",
    };
    const pets = [
      { id: "10", ownerId: "20", name: "ポチ" },
      { id: "11", ownerId: "20", name: "タマ" },
    ];

    await act(async () => { await result.current.handleSave(data, pets); });
    act(() => result.current.resetCreateProgress());
    await act(async () => { await result.current.handleSave(data, pets); });
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(4);
  });

  it("does not exclude other overlapping reservations after a partial create", async () => {
    createMutateAsyncMock
      .mockResolvedValueOnce(makeReservation({ id: "created-10" }))
      .mockRejectedValueOnce(new Error("temporary failure"));
    const { result, rerender } = setup();
    const data: ReservationFormData = {
      start: new Date("2026-05-29T03:30:00.000Z"),
      end: new Date("2026-05-29T04:00:00.000Z"),
      visitType: "first",
      type: "1",
      doctor: "1",
      status: "confirmed",
    };
    const pets = [
      { id: "10", ownerId: "20", name: "ポチ" },
      { id: "11", ownerId: "20", name: "タマ" },
    ];

    await act(async () => { await result.current.handleSave(data, pets); });
    const correctedData = {
      ...data,
      start: new Date("2026-05-29T05:00:00.000Z"),
      end: new Date("2026-05-29T05:30:00.000Z"),
      doctor: "2",
    };
    rerender({
      statusConfirmTarget: null,
      appointments: [
        makeReservation({ id: "created-10", petId: "10" }),
        makeReservation({
          id: "unrelated",
          petId: "10",
          doctor: "2",
          doctorId: "2",
          start: correctedData.start,
          end: correctedData.end,
        }),
      ],
    });

    let retryResult: string | null = null;
    await act(async () => {
      retryResult = await result.current.handleSave(correctedData, pets);
    });
    expect(retryResult).toBe("指定された時間帯には既に予約が入っています");
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(2);
  });

  it("resets partial progress when new-owner success closes before reopening", async () => {
    createMutateAsyncMock
      .mockResolvedValueOnce(makeReservation({ id: "created-10" }))
      .mockRejectedValueOnce(new Error("temporary failure"))
      .mockResolvedValueOnce(makeReservation({ id: "created-new-owner" }))
      .mockResolvedValueOnce(makeReservation({ id: "created-10-reopened" }));
    const createOwnerFn = vi.fn().mockResolvedValue({ id: 30 });
    const createPetFn = vi.fn().mockResolvedValue({ id: 40 });
    const { result, handleCloseForm } = setup({ createOwnerFn, createPetFn });
    const data: ReservationFormData = {
      start: new Date("2026-05-29T03:30:00.000Z"),
      end: new Date("2026-05-29T04:00:00.000Z"),
      visitType: "first",
      type: "1",
      doctor: "1",
      status: "confirmed",
    };
    const pets = [
      { id: "10", ownerId: "20", name: "ポチ" },
      { id: "11", ownerId: "20", name: "タマ" },
    ];

    await act(async () => { await result.current.handleSave(data, pets); });
    await act(async () => {
      await result.current.handleSave(data, [], {
        ownerName: "新規飼主",
        phone: "09012345678",
        petName: "新規ペット",
        animalSpeciesId: 1,
        chiefComplaint: "相談",
      });
    });
    expect(handleCloseForm).toHaveBeenCalledTimes(1);

    await act(async () => {
      await result.current.handleSave(data, [pets[0]]);
    });
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(4);
    expect(createMutateAsyncMock.mock.calls[3]?.[0]).toEqual(
      expect.objectContaining({ pet_id: 10 }),
    );
  });

});
