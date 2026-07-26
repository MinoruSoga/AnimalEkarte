import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";
import { PetStatusDeceased } from "@/types/generated/models";

import { useReceptionModalHandlers } from "./use-reception-modal-handlers";
import type { ReceptionAppointment } from "../api/types";

const { updateReservationMock } = vi.hoisted(() => ({
  updateReservationMock: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn() },
}));

vi.mock("@/hooks/use-update-reservation", () => ({
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

interface HandlerProps {
  canEditReservation?: boolean;
  canDeleteReservation?: boolean;
}

function renderHandlers(
  updateAppointment = vi.fn(),
  initialPermissions: HandlerProps = {
    canEditReservation: true,
    canDeleteReservation: true,
  },
) {
  const advanceStatus = vi.fn();
  const cancelAppointment = vi.fn();
  const view = renderHook(
    (permissions: HandlerProps) => useReceptionModalHandlers({
      advanceStatus,
      cancelAppointment,
      updateAppointment,
      ...permissions,
    }),
    { initialProps: initialPermissions },
  );
  return { ...view, advanceStatus, cancelAppointment };
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

    expect(result.current.editingAppointment?.start?.getFullYear()).toBe(2026);
    expect(result.current.editingAppointment?.start?.getMonth()).toBe(5);
    expect(result.current.editingAppointment?.start?.getDate()).toBe(1);
    expect(result.current.editingAppointment?.start?.getHours()).toBe(9);
    expect(result.current.editingAppointment?.start?.getMinutes()).toBe(45);
    expect(result.current.editingAppointment?.end?.getHours()).toBe(10);
    expect(result.current.editingAppointment?.end?.getMinutes()).toBe(45);
    expect(result.current.editingAppointment?.type).toBe("1");
    expect(result.current.editingAppointment?.doctor).toBe("33");
    expect(result.current.editingAppointment?.status).toBe("checked_in");
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
          start: new Date(2026, 5, 1, 9, 45, 0),
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
          pet_id: 10,
          owner_id: 20,
          reservation_type_id: 1,
          doctor_id: 33,
          is_designated: false,
          status: "checked_in",
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
          start: new Date(2026, 5, 1, 9, 45, 0),
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

  it("編集対象に新しく選択されたペットが死亡の場合は予約更新を拒否する", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [
          {
            id: "11",
            ownerId: "21",
            status: "死亡",
          } as Pet,
        ],
      );
    });

    expect(updateReservationMock).not.toHaveBeenCalled();
  });

  it("受付予約カラムの編集では pending 表示を confirmed として保存する", () => {
    const { result } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment({ ...baseAppointment, status: "pending" });
    });
    act(() => {
      result.current.handleEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
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
          status: "confirmed",
        }),
      },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });

  it("死亡 appointment の card click・status・edit・cancel callback を拒否する", () => {
    const deceasedAppointment = {
      ...baseAppointment,
      petStatus: PetStatusDeceased,
    };
    const { result, advanceStatus, cancelAppointment } = renderHandlers();

    act(() => {
      result.current.handleCardClick(deceasedAppointment);
      result.current.handleAdvanceStatus();
      result.current.handleEditAppointment(deceasedAppointment);
      result.current.handleCancelAppointment(deceasedAppointment);
      result.current.executeCancel();
    });

    expect(result.current.modalOpen).toBe(false);
    expect(result.current.isEditModalOpen).toBe(false);
    expect(result.current.cancelConfirmOpen).toBe(false);
    expect(advanceStatus).not.toHaveBeenCalled();
    expect(cancelAppointment).not.toHaveBeenCalled();
    expect(updateReservationMock).not.toHaveBeenCalled();
  });

  it("captured status callback は commit 後の最新 edit 権限を再確認する", () => {
    const { result, rerender, advanceStatus } = renderHandlers();

    act(() => {
      result.current.handleCardClick(baseAppointment);
    });
    const capturedAdvanceStatus = result.current.handleAdvanceStatus;

    rerender({ canEditReservation: false, canDeleteReservation: true });
    act(() => {
      capturedAdvanceStatus();
    });

    expect(advanceStatus).not.toHaveBeenCalled();
  });

  it("captured edit save callback は commit 後の最新 edit 権限を再確認する", () => {
    const { result, rerender } = renderHandlers();

    act(() => {
      result.current.handleEditAppointment(baseAppointment);
    });
    const capturedEditSave = result.current.handleEditSave;

    rerender({ canEditReservation: undefined, canDeleteReservation: true });
    act(() => {
      capturedEditSave(
        {
          start: new Date(2026, 5, 1, 9, 45, 0),
          visitType: "revisit",
          type: "1",
          doctor: "33",
        },
        [],
      );
    });

    expect(updateReservationMock).not.toHaveBeenCalled();
  });

  it("captured cancel callback は commit 後の最新 delete 権限を再確認する", () => {
    const { result, rerender, cancelAppointment } = renderHandlers();

    act(() => {
      result.current.handleCancelAppointment(baseAppointment);
    });
    const capturedExecuteCancel = result.current.executeCancel;

    rerender({ canEditReservation: true, canDeleteReservation: false });
    act(() => {
      capturedExecuteCancel();
    });

    expect(cancelAppointment).not.toHaveBeenCalled();
  });
});
