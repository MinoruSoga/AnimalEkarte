import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { MutableRefObject } from "react";
import { toast } from "sonner";

import type { Reservation, ReservationFormData, ReservationStatus } from "../types";
import {
  buildReservationUpdateRequest,
  isDestructiveReservationStatus,
  PERMISSION_DENIED_MESSAGE,
  useReservationActions,
  type ReservationMutationPermissions,
  type StatusConfirmTarget,
} from "./use-reservation-actions";

const updateMutateMock = vi.fn();
const createMutateAsyncMock = vi.fn();
const createBatchMutateAsyncMock = vi.fn();
const deleteMutateMock = vi.fn();

vi.mock("../api/create-reservation", () => ({
  useCreateReservation: () => ({ mutateAsync: createMutateAsyncMock }),
  useCreateReservationBatch: () => ({ mutateAsync: createBatchMutateAsyncMock }),
}));

vi.mock("../api/delete-reservation", () => ({
  useDeleteReservation: () => ({ mutate: deleteMutateMock }),
}));

vi.mock("../api/update-reservation", () => ({
  useUpdateReservation: () => ({ mutate: updateMutateMock }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

// FE-RC-204: 既存テスト群は「権限あり」の通常フローを検証するため、既定で全許可する。
// 権限拒否の fail-closed 挙動は専用の describe で検証する。
const ALLOW_ALL_PERMISSIONS: Readonly<ReservationMutationPermissions> = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
};

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
    doctor: "山田獣医師",
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
  permissions?: Readonly<ReservationMutationPermissions>;
  deleteTarget?: Reservation | null;
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
    statusConfirmTarget = typeof value === "function" ? value(statusConfirmTarget) : value;
  });

  const handleCloseForm = vi.fn();
  const createOwnerFn = options.createOwnerFn ?? vi.fn();
  const createPetFn = options.createPetFn ?? vi.fn();
  const { result, rerender } = renderHook(
    (props: {
      statusConfirmTarget: StatusConfirmTarget | null;
      appointments: Reservation[];
      permissions?: Readonly<ReservationMutationPermissions>;
      deleteTarget: Reservation | null;
    }) =>
      useReservationActions({
        appointments: props.appointments,
        editingAppointmentRef,
        deleteTarget: props.deleteTarget,
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
        permissions: props.permissions,
      }),
    {
      initialProps: {
        statusConfirmTarget: null as StatusConfirmTarget | null,
        appointments: options.appointments ?? [],
        permissions: options.permissions,
        deleteTarget: options.deleteTarget ?? null,
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
    editingAppointmentRef,
  };
}

function makeSaveFormData(): ReservationFormData {
  return {
    start: new Date("2026-05-29T03:30:00.000Z"),
    end: new Date("2026-05-29T04:00:00.000Z"),
    visitType: "first",
    type: "1",
    doctor: "1",
    status: "confirmed",
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
    const { result } = setup({ permissions: ALLOW_ALL_PERMISSIONS });
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
    const { result, rerender, setStatusConfirmOpen, getStatusConfirmTarget } = setup({
      permissions: ALLOW_ALL_PERMISSIONS,
    });
    const reservation = makeReservation({ status: "confirmed" });
    const next: ReservationStatus = "cancelled";

    act(() => {
      result.current.handleStatusChange(reservation, next);
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(setStatusConfirmOpen).toHaveBeenCalledWith(true);
    const pending = getStatusConfirmTarget();
    expect(pending).toEqual({ reservation, status: next });

    rerender({
      statusConfirmTarget: pending,
      appointments: [],
      permissions: ALLOW_ALL_PERMISSIONS,
      deleteTarget: null,
    });

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
    const { result } = setup({ permissions: ALLOW_ALL_PERMISSIONS });
    const reservation = makeReservation({ status: "confirmed" });

    act(() => {
      result.current.handleStatusChange(reservation, "confirmed");
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
  });

  it("受付済→診療中はカルテ作成必須のため PATCH しない", () => {
    const { result } = setup({ permissions: ALLOW_ALL_PERMISSIONS });
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
    const payload = buildReservationUpdateRequest(current, { ...current, end: nextEnd }, "1");
    expect(payload?.req).toHaveProperty("end_time");
    expect(payload?.req).not.toHaveProperty("doctor_id");
    expect(payload?.req).not.toHaveProperty("start_time");
  });
});

describe("useReservationActions multi-pet batch retry", () => {
  beforeEach(() => {
    createBatchMutateAsyncMock.mockReset();
  });
  it("posts selected pets once as an atomic batch and retries the same payload after a transport failure", async () => {
    createBatchMutateAsyncMock
      .mockRejectedValueOnce(new Error("network lost"))
      .mockResolvedValueOnce([]);
    const { result } = setup({ permissions: ALLOW_ALL_PERMISSIONS });
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
    await act(async () => {
      expect(await result.current.handleSave(data, pets)).toBeTruthy();
    });
    await act(async () => {
      expect(await result.current.handleSave(data, pets)).toBeNull();
    });
    expect(createBatchMutateAsyncMock).toHaveBeenCalledTimes(2);
    expect(createBatchMutateAsyncMock.mock.calls[0][0]).toEqual(
      createBatchMutateAsyncMock.mock.calls[1][0],
    );
    expect(createBatchMutateAsyncMock.mock.calls[0][0]).toEqual(
      expect.objectContaining({
        pets: [
          { pet_id: 10, owner_id: 20 },
          { pet_id: 11, owner_id: 20 },
        ],
      }),
    );
  });
});

describe("useReservationActions new-owner retry", () => {
  it("retries a new-owner reservation without recreating its committed owner or pet", async () => {
    const createOwnerFn = vi.fn().mockResolvedValue({ id: 30 });
    const createPetFn = vi.fn().mockResolvedValue({ id: 40 });
    createMutateAsyncMock
      .mockRejectedValueOnce(new Error("reservation failed"))
      .mockResolvedValueOnce(makeReservation({ id: "created-40", ownerId: "30", petId: "40" }));
    const { result } = setup({ createOwnerFn, createPetFn, permissions: ALLOW_ALL_PERMISSIONS });
    const data: ReservationFormData = {
      start: new Date("2026-05-29T03:30:00.000Z"),
      end: new Date("2026-05-29T04:00:00.000Z"),
      visitType: "first",
      type: "1",
      doctor: "1",
      status: "confirmed",
    };
    const newOwner = {
      ownerName: "新規飼主",
      phone: "09012345678",
      petName: "新規ペット",
      animalSpeciesId: 1,
      chiefComplaint: "相談",
    };

    await act(async () => {
      await result.current.handleSave(data, [], newOwner);
    });
    await act(async () => {
      await result.current.handleSave(data, [], newOwner);
    });
    expect(createOwnerFn).toHaveBeenCalledTimes(1);
    expect(createPetFn).toHaveBeenCalledTimes(1);
    expect(createMutateAsyncMock).toHaveBeenCalledTimes(2);
  });
});

// ──────────────────────────────────────────────────────────
// FE-RC-204: permissionsRef + isMutationAllowed による mutation 直前の権限再検査。
// UI の非表示/disabled をバイパスされても fail-closed で API を呼ばない。
// ──────────────────────────────────────────────────────────

describe("useReservationActions permissions (FE-RC-204 fail-closed)", () => {
  const pets = [{ id: "10", ownerId: "20", name: "ポチ" }];

  beforeEach(() => {
    updateMutateMock.mockReset();
    createMutateAsyncMock.mockReset();
    createBatchMutateAsyncMock.mockReset();
    deleteMutateMock.mockReset();
    vi.mocked(toast.error).mockClear();
    vi.mocked(toast.success).mockClear();
    createMutateAsyncMock.mockResolvedValue({});
    updateMutateMock.mockImplementation((_payload, opts?: { onSuccess?: () => void }) => {
      opts?.onSuccess?.();
    });
  });

  it("canCreate=false（新規作成）→ create mutate は呼ばれない", async () => {
    const { result } = setup({
      permissions: { canCreate: false, canEdit: true, canDelete: true },
    });

    await act(async () => {
      await result.current.handleSave(makeSaveFormData(), pets);
    });

    expect(createMutateAsyncMock).not.toHaveBeenCalled();
    expect(createBatchMutateAsyncMock).not.toHaveBeenCalled();
    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false（編集保存）→ update mutate は呼ばれない", async () => {
    const { result, editingAppointmentRef } = setup({
      permissions: { canCreate: true, canEdit: false, canDelete: true },
    });
    editingAppointmentRef.current = { ...makeSaveFormData(), id: "r1" };

    await act(async () => {
      await result.current.handleSave(makeSaveFormData(), pets);
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(createMutateAsyncMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false（DnD）→ update mutate は呼ばれない", () => {
    const { result } = setup({
      permissions: { canCreate: true, canEdit: false, canDelete: true },
    });
    const reservation = makeReservation();

    act(() => {
      result.current.handleReservationUpdate(
        reservation,
        new Date("2026-05-29T05:00:00.000Z"),
        new Date("2026-05-29T05:30:00.000Z"),
      );
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false（status 変更）→ update mutate は呼ばれない", () => {
    const { result } = setup({
      permissions: { canCreate: true, canEdit: false, canDelete: true },
    });

    act(() => {
      result.current.handleStatusChange(makeReservation({ status: "confirmed" }), "checked_in");
    });

    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canDelete=false（削除）→ delete mutate は呼ばれない", () => {
    const reservation = makeReservation();
    const { result } = setup({
      permissions: { canCreate: true, canEdit: true, canDelete: false },
      deleteTarget: reservation,
    });

    act(() => {
      result.current.executeDelete();
    });

    expect(deleteMutateMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("permissions 未指定（既定 deny）→ create/update/delete いずれも呼ばれない", async () => {
    const reservation = makeReservation();
    const { result, editingAppointmentRef } = setup({ deleteTarget: reservation });

    await act(async () => {
      await result.current.handleSave(makeSaveFormData(), pets);
    });
    act(() => {
      result.current.handleReservationUpdate(
        reservation,
        new Date("2026-05-29T05:00:00.000Z"),
        new Date("2026-05-29T05:30:00.000Z"),
      );
    });
    act(() => {
      result.current.handleStatusChange(reservation, "checked_in");
    });
    act(() => {
      result.current.executeDelete();
    });

    editingAppointmentRef.current = { ...makeSaveFormData(), id: "r1" };
    await act(async () => {
      await result.current.handleSave(makeSaveFormData(), pets);
    });

    expect(createMutateAsyncMock).not.toHaveBeenCalled();
    expect(createBatchMutateAsyncMock).not.toHaveBeenCalled();
    expect(updateMutateMock).not.toHaveBeenCalled();
    expect(deleteMutateMock).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("ALLOW ALL の新規作成は create mutate を呼ぶ", async () => {
    const { result } = setup({ permissions: ALLOW_ALL_PERMISSIONS });

    await act(async () => {
      await result.current.handleSave(makeSaveFormData(), pets);
    });

    expect(createMutateAsyncMock).toHaveBeenCalledTimes(1);
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("ALLOW ALL の DnD は update mutate を呼ぶ", () => {
    const { result } = setup({ permissions: ALLOW_ALL_PERMISSIONS });

    act(() => {
      result.current.handleReservationUpdate(
        makeReservation(),
        new Date("2026-05-29T05:00:00.000Z"),
        new Date("2026-05-29T05:30:00.000Z"),
      );
    });

    expect(updateMutateMock).toHaveBeenCalledTimes(1);
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});
