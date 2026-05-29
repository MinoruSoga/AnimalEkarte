import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useReceptionModalHandlers } from "./useReceptionModalHandlers";
import type { ReceptionAppointment } from "../api/types";

const { updateReservationMock } = vi.hoisted(() => ({
  updateReservationMock: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn() },
}));

vi.mock("@/features/reservations", () => ({
  useUpdateReservation: () => ({
    mutate: updateReservationMock,
  }),
}));

const baseAppointment: ReceptionAppointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-06-01",
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
};

function renderHandlers(updateAppointment = vi.fn()) {
  return renderHook(() => useReceptionModalHandlers({
    advanceStatus: vi.fn(),
    cancelAppointment: vi.fn(),
    updateAppointment,
  }));
}

describe("useReceptionModalHandlers", () => {
  beforeEach(() => {
    updateReservationMock.mockReset();
    updateReservationMock.mockImplementation((_payload, options?: { onSuccess?: () => void }) => {
      options?.onSuccess?.();
    });
  });

  it("受付予約編集では appointment の visitDate と time から開始日時を作る", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });

    expect(result.current.editingAppointment?.start?.toISOString()).toBe("2026-06-01T00:45:00.000Z");
    expect(result.current.editingAppointment?.end?.toISOString()).toBe("2026-06-01T01:45:00.000Z");
    expect(result.current.editingAppointment?.type).toBe("1");
    expect(result.current.editingAppointment?.doctor).toBe("33");
  });

  it("編集保存では予約更新 API を呼び、成功後にローカル appointment も更新する", () => {
    const updateAppointment = vi.fn();
    const { result } = renderHandlers(updateAppointment);

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date("2026-06-01T09:45:00+09:00"),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [],
      );
    });

    expect(updateReservationMock).toHaveBeenCalledWith(
      {
        id: "101",
        req: expect.objectContaining({
          start_time: "2026-06-01T00:45:00.000Z",
          end_time: "2026-06-01T01:45:00.000Z",
          visit_type: "revisit",
          reservation_type_id: 1,
          doctor_id: 33,
          is_designated: false,
          status: "confirmed",
        }),
      },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    expect(updateAppointment).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "101",
        time: "09:45",
        visitDate: "2026-06-01",
      }),
    );
  });

  it("編集保存では予約区分IDと担当者IDが不正値なら undefined にする", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date("2026-06-01T09:45:00+09:00"),
          visitType: "revisit",
          type: "一般診察",
          doctor: "担当者A",
        },
        [],
      );
    });

    expect(updateReservationMock).toHaveBeenCalledWith(
      {
        id: "101",
        req: expect.objectContaining({
          reservation_type_id: undefined,
          doctor_id: undefined,
        }),
      },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });
});
